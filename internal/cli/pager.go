package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

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
