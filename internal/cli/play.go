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
	"github.com/hobbymarks/gtr/internal/engine/google"
	"github.com/hobbymarks/gtr/internal/httpx"
)

func googleTTSURLForEngine(eng engine.Engine, in engine.TranslateInput, translated string) (string, error) {
	switch e := eng.(type) {
	case *google.Engine:
		_ = e
		return google.BuildTTSURL(translated, in.Target)
	case *auto.Engine:
		_ = e
		if auto.PickBackend(in.Source, in.Target) != "google" {
			return "", fmt.Errorf("TTS is only wired when auto routes to google (picked %s)", auto.PickBackend(in.Source, in.Target))
		}
		return google.BuildTTSURL(translated, in.Target)
	default:
		return "", fmt.Errorf("engine %q does not expose Google TTS in this build", eng.Name())
	}
}

func playGoogleTTS(ctx context.Context, u string) error {
	client := httpx.NewClient()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("TTS fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("TTS HTTP %d", resp.StatusCode)
	}
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

func playAudioFile(ctx context.Context, path string) error {
	candidates := [][]string{
		{"mpv", "--no-video", path},
		{"ffplay", "-nodisp", "-autoexit", path},
	}
	if runtime.GOOS == "darwin" {
		candidates = append([][]string{{"afplay", path}}, candidates...)
	}
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/c", "start", "", "\""+path+"\"").Run()
	}
	var lastErr error
	for _, argv := range candidates {
		bin, err := exec.LookPath(argv[0])
		if err != nil {
			lastErr = err
			continue
		}
		cmd := exec.CommandContext(ctx, bin, argv[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no audio player found (tried mpv, ffplay)")
	}
	return fmt.Errorf("play audio: %w", lastErr)
}

// openPagerWriter returns a WriteCloser that feeds the system pager (PAGER or less -R / more).
func openPagerWriter(ctx context.Context) (io.WriteCloser, func(), error) {
	pager := strings.TrimSpace(os.Getenv("PAGER"))
	if pager == "" {
		if runtime.GOOS == "windows" {
			pager = "more"
		} else {
			pager = "less -R"
		}
	}
	argv := strings.Fields(pager)
	if len(argv) == 0 {
		return nil, func() {}, fmt.Errorf("empty PAGER")
	}
	bin, err := exec.LookPath(argv[0])
	if err != nil {
		return nil, func() {}, fmt.Errorf("pager %q: %w", argv[0], err)
	}
	cmd := exec.CommandContext(ctx, bin, argv[1:]...)
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, func() {}, err
	}
	if err := cmd.Start(); err != nil {
		return nil, func() {}, err
	}
	cleanup := func() {
		_ = stdin.Close()
		_ = cmd.Wait()
	}
	return stdin, cleanup, nil
}
