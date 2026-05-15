package google

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hobbymarks/gtr/internal/engine"
)

// IdentifyLanguage performs a minimal auto→pivot translate and returns the
// detected source language code from the JSON payload.
func (e *Engine) IdentifyLanguage(ctx context.Context, text, hostLang string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("google: empty text")
	}
	hl := hostLang
	if hl == "" {
		hl = "en"
	}
	in := engine.TranslateInput{
		Text:     text,
		Source:   "auto",
		Target:   "en",
		HostLang: hl,
		Brief:    true,
	}
	u, err := buildSingleRequestURL(in.Text, in.Source, in.Target, in.HostLang, false)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	resp, err := e.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("google: identify: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxReadBody+1))
	if err != nil {
		return "", fmt.Errorf("google: read body: %w", err)
	}
	if int64(len(body)) > maxReadBody {
		return "", fmt.Errorf("google: response body exceeds %d bytes", maxReadBody)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("google: identify HTTP %d", resp.StatusCode)
	}
	return ParseDetectedSourceLanguage(body)
}

var _ engine.LanguageIdentifier = (*Engine)(nil)
