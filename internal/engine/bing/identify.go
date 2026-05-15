package bing

import (
	"context"
	"fmt"
	"strings"

	"github.com/hobbymarks/gtr/internal/engine"
)

// IdentifyLanguage runs auto-detect to a pivot target and returns Bing's detected source code.
func (e *Engine) IdentifyLanguage(ctx context.Context, text, hostLang string) (string, error) {
	_ = hostLang // Bing translate POST does not take UI language in this client path.
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("bing: empty text")
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	hosts := []string{"https://www.bing.com", "https://cn.bing.com"}
	var lastErr error
	for _, origin := range hosts {
		body, status, err := e.translateRequest(ctx, origin, text, "auto", "en", false)
		if err != nil {
			lastErr = err
			continue
		}
		if status >= 400 {
			lastErr = fmt.Errorf("bing: identify HTTP %d", status)
			continue
		}
		_, detected, err := parseBingResponse(body)
		if err != nil {
			lastErr = err
			continue
		}
		if strings.TrimSpace(detected) == "" {
			lastErr = fmt.Errorf("bing: no detected language in response")
			continue
		}
		return detected, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("bing: identify failed")
	}
	return "", lastErr
}

var _ engine.LanguageIdentifier = (*Engine)(nil)
