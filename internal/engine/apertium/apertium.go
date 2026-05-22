// Package apertium implements the Apertium APy translation engine for gtr.
package apertium

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/hobbymarks/gtr/internal/engine"
	"github.com/hobbymarks/gtr/internal/httpx"
)

const translateBase = "https://www.apertium.org"

// Engine calls Apertium APy translate endpoint (translate-shell parity).
type Engine struct {
	HTTP *http.Client
}

func New(c *http.Client) *Engine {
	if c == nil {
		c = httpx.NewClient()
	}
	return &Engine{HTTP: c}
}

func (e *Engine) Name() string { return "apertium" }

func (e *Engine) Translate(ctx context.Context, in engine.TranslateInput) (engine.TranslateOutput, error) {
	sl := strings.TrimSpace(in.Source)
	if sl == "auto" {
		sl = "en"
	}
	tl := strings.TrimSpace(in.Target)

	u, err := url.Parse(translateBase + "/apy/translate")
	if err != nil {
		return engine.TranslateOutput{}, err
	}
	q := url.Values{}
	q.Set("langpair", sl+"|"+tl)
	q.Set("q", in.Text)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return engine.TranslateOutput{}, err
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")

	if in.Debug {
		_, _ = fmt.Fprintf(os.Stderr, "apertium debug: GET %s\n", u.String())
	}

	resp, err := e.HTTP.Do(req)
	if err != nil {
		return engine.TranslateOutput{}, fmt.Errorf("apertium: request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, engine.MaxReadBody+1))
	if err != nil {
		return engine.TranslateOutput{}, fmt.Errorf("apertium: read body: %w", err)
	}
	if int64(len(body)) > engine.MaxReadBody {
		return engine.TranslateOutput{}, fmt.Errorf("apertium: response body exceeds %d bytes", engine.MaxReadBody)
	}

	if in.Dump {
		return engine.TranslateOutput{Text: string(body)}, nil
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return engine.TranslateOutput{}, fmt.Errorf("apertium: rate limiting is in effect (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		return engine.TranslateOutput{}, fmt.Errorf("apertium: HTTP %d: %s", resp.StatusCode, httpx.Truncate(body, 200))
	}

	text, err := parseTranslateBody(body)
	if err != nil {
		return engine.TranslateOutput{}, err
	}
	if in.Brief {
		text = strings.TrimSpace(text)
	}
	return engine.TranslateOutput{Text: text}, nil
}

func parseTranslateBody(raw []byte) (string, error) {
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return "", fmt.Errorf("apertium: invalid JSON: %w", err)
	}
	if ex, ok := root["exception"]; ok {
		return "", fmt.Errorf("apertium: %v", ex)
	}
	if errStr, ok := root["error"].(string); ok && strings.TrimSpace(errStr) != "" {
		return "", fmt.Errorf("apertium: %s", errStr)
	}
	data, ok := root["responseData"].(map[string]any)
	if !ok || data == nil {
		return "", fmt.Errorf("apertium: unsupported language pair or empty responseData")
	}
	txt, ok := data["translatedText"].(string)
	if !ok || strings.TrimSpace(txt) == "" {
		return "", fmt.Errorf("apertium: empty translatedText (pair may be unsupported)")
	}
	return txt, nil
}
