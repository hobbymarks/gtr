package cli

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestOpenPagerWriter_cat(t *testing.T) {
	orig := os.Getenv("PAGER")
	os.Setenv("PAGER", "cat")
	defer os.Setenv("PAGER", orig)

	wc, cleanup, err := openPagerWriter(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	_, err = wc.Write([]byte("hello\n"))
	if err != nil {
		t.Fatal(err)
	}
	err = wc.Close()
	if err != nil {
		t.Fatal(err)
	}
	cleanup()
}

func TestOpenPagerWriter_nonexistent(t *testing.T) {
	orig := os.Getenv("PAGER")
	os.Setenv("PAGER", "nonexistent-binary-xyz")
	defer os.Setenv("PAGER", orig)

	_, _, err := openPagerWriter(context.Background())
	if err == nil {
		t.Fatal("expected error for nonexistent pager")
	}
}

func TestPlayAudioFile_errorsOnMissingPlayer(t *testing.T) {
	// On a system with no audio players, this should error.
	// We check that the function runs without panic and returns a descriptive error.
	err := playAudioFile(context.Background(), "/dev/null")
	if err == nil && (playerExists("mpv") || playerExists("ffplay")) {
		t.Skip("player found, skipping missing-player test")
	}
	if err != nil {
		if !strings.Contains(err.Error(), "play audio") &&
			!strings.Contains(err.Error(), "no audio player found") &&
			!strings.Contains(err.Error(), "exec:") {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func playerExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
