package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/hobbymarks/gtr/internal/engine"
	"github.com/hobbymarks/gtr/internal/engine/auto"
	"github.com/hobbymarks/gtr/internal/engine/bing"
	"github.com/hobbymarks/gtr/internal/engine/google"
	"github.com/hobbymarks/gtr/internal/httpx"
)

func ttsURLForEngine(eng engine.Engine, in engine.TranslateInput, translated string) (string, error) {
	switch e := eng.(type) {
	case *google.Engine:
		_ = e
		return google.BuildTTSURL(translated, in.Target)
	case *bing.Engine:
		_ = e
		return bing.BuildTTSURL(translated, in.Target)
	case *auto.Engine:
		_ = e
		switch auto.PickBackend(in.Source, in.Target) {
		case "google":
			return google.BuildTTSURL(translated, in.Target)
		case "bing":
			return bing.BuildTTSURL(translated, in.Target)
		default:
			return "", fmt.Errorf("TTS is only wired when auto routes to google or bing (picked %s)", auto.PickBackend(in.Source, in.Target))
		}
	default:
		return "", fmt.Errorf("engine %q does not support TTS in this build", eng.Name())
	}
}

func fetchTTSResponse(ctx context.Context, u string) (*http.Response, error) {
	client := httpx.NewClient()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("TTS fetch: %w", err)
	}
	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return nil, fmt.Errorf("TTS HTTP %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "" && !strings.HasPrefix(ct, "audio/") {
		resp.Body.Close()
		return nil, fmt.Errorf("TTS: unexpected Content-Type %q (expected audio/*)", ct)
	}
	return resp, nil
}

func playTTS(ctx context.Context, u string) error {
	resp, err := fetchTTSResponse(ctx, u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Try streaming directly to player via stdin (faster, no temp file)
	for _, args := range [][]string{
		{"mpv", "--no-video", "--really-quiet", "-"},
		{"ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet", "-"},
	} {
		bin, err := exec.LookPath(args[0])
		if err != nil {
			continue
		}
		cmd := exec.CommandContext(ctx, bin, args[1:]...)
		cmd.Stdin = io.LimitReader(resp.Body, 8<<20)
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		return cmd.Run()
	}

	// Fallback: write to temp file and play
	f, err := os.CreateTemp("", "gtr-tts-*.mp3")
	if err != nil {
		return err
	}
	path := f.Name()
	defer os.Remove(path)
	if _, err := io.Copy(f, io.LimitReader(resp.Body, 8<<20)); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return playAudioFile(ctx, path)
}

func downloadTTSFile(ctx context.Context, u, dstPath string) error {
	resp, err := fetchTTSResponse(ctx, u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	f, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, io.LimitReader(resp.Body, 8<<20)); err != nil {
		return err
	}
	return nil
}

func playAudioFile(ctx context.Context, path string) error {
	candidates := [][]string{
		{"mpv", "--no-video", "--really-quiet", path},
		{"ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet", path},
		{"paplay", path},
		{"aplay", path},
		{"play", "-q", path},
		{"cvlc", "--play-and-exit", "--no-video", "--intf", "dummy", path},
	}
	if runtime.GOOS == "darwin" {
		candidates = append([][]string{{"afplay", path}}, candidates...)
	}
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/c", "start", "", "\""+path+"\"").Run()
	}
	var errs []string
	for _, argv := range candidates {
		bin, err := exec.LookPath(argv[0])
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: not found", argv[0]))
			continue
		}
		cmd := exec.CommandContext(ctx, bin, argv[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", argv[0], err))
			continue
		}
		return nil
	}
	if len(errs) == 0 {
		errs = append(errs, "no audio player found")
	}
	return fmt.Errorf("play audio: %s", strings.Join(errs, "; "))
}
