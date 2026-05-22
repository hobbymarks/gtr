// Package google implements the Google Translate engine for gtr,
// using the translate_a/single public endpoint with phonetic and dictionary support.
package google

import (
	"bytes"
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

// Engine calls Google Translate web-style endpoint (translate-shell parity: client=gtx).
type Engine struct {
	HTTP *http.Client
}

// New returns a Google engine using the given HTTP client (typically [httpx.NewClient]).
func New(c *http.Client) *Engine {
	if c == nil {
		c = httpx.NewClient()
	}
	return &Engine{HTTP: c}
}

func (e *Engine) Name() string { return "google" }

func (e *Engine) Translate(ctx context.Context, in engine.TranslateInput) (engine.TranslateOutput, error) {
	u, err := buildSingleRequestURL(in.Text, in.Source, in.Target, in.HostLang, in.NoAutocorrect)
	if err != nil {
		return engine.TranslateOutput{}, err
	}
	if in.Debug {
		_, _ = fmt.Fprintf(os.Stderr, "gtr debug: GET %s\n", u)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return engine.TranslateOutput{}, err
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := e.HTTP.Do(req)
	if err != nil {
		return engine.TranslateOutput{}, fmt.Errorf("google: request: %w", err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, engine.MaxReadBody+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return engine.TranslateOutput{}, fmt.Errorf("google: read body: %w", err)
	}
	if int64(len(body)) > engine.MaxReadBody {
		return engine.TranslateOutput{}, fmt.Errorf("google: response body exceeds %d bytes", engine.MaxReadBody)
	}

	if in.Dump {
		return engine.TranslateOutput{Text: string(body)}, nil
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return engine.TranslateOutput{}, fmt.Errorf("google: rate limiting is in effect (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		return engine.TranslateOutput{}, fmt.Errorf("google: HTTP %d: %s", resp.StatusCode, httpx.Truncate(body, 200))
	}

	text, phonetic, err := ParseTranslateSingleResponse(body)
	if err != nil {
		return engine.TranslateOutput{}, err
	}
	if in.Brief {
		text = strings.TrimSpace(text)
		phonetic = strings.TrimSpace(phonetic)
	}
	out := engine.TranslateOutput{Text: text, Phonetic: phonetic}
	if in.Dictionary {
		if d, err := FormatDictionaryPayload(body); err == nil && d != "" {
			out.Dictionary = d
		}
	}
	return out, nil
}

func buildSingleRequestURL(text, sl, tl, hl string, noAutocorrect bool) (string, error) {
	qc := "qca"
	if noAutocorrect {
		qc = "qc"
	}
	base, err := url.Parse("https://translate.googleapis.com/translate_a/single")
	if err != nil {
		return "", err
	}
	v := url.Values{}
	v.Set("client", "gtx")
	v.Set("ie", "UTF-8")
	v.Set("oe", "UTF-8")
	for _, dt := range []string{"bd", "ex", "ld", "md", "rw", "rm", "ss", "t", "at", "gt"} {
		v.Add("dt", dt)
	}
	v.Add("dt", qc)
	v.Set("sl", sl)
	v.Set("tl", tl)
	v.Set("hl", hl)
	v.Set("q", text)
	base.RawQuery = v.Encode()
	return base.String(), nil
}

// ParseTranslateSingleResponse extracts the primary translation and phonetic
// romanization from a translate_a/single JSON payload (same logical paths as
// translate-shell flattened indices for sentence translations).
func ParseTranslateSingleResponse(raw []byte) (text string, phonetic string, err error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return "", "", fmt.Errorf("google: empty response body")
	}
	var root any
	if err := json.Unmarshal(raw, &root); err != nil {
		return "", "", fmt.Errorf("google: invalid JSON: %w", err)
	}
	out, phon, err := joinSentenceTranslations(root)
	if err != nil {
		return "", "", err
	}
	if strings.TrimSpace(out) == "" {
		return "", "", fmt.Errorf("google: could not parse translation from response")
	}
	return out, phon, nil
}

func joinSentenceTranslations(v any) (text string, phonetic string, err error) {
	root, ok := v.([]any)
	if !ok || len(root) == 0 {
		return "", "", fmt.Errorf("google: unexpected JSON root shape")
	}
	sentences, ok := root[0].([]any)
	if !ok {
		return "", "", fmt.Errorf("google: missing sentence list at index 0")
	}
	var b strings.Builder
	for _, s := range sentences {
		seg, ok := s.([]any)
		if !ok || len(seg) < 1 {
			continue
		}
		t, ok := seg[0].(string)
		if !ok {
			continue
		}
		b.WriteString(t)
	}
	text = b.String()

	if len(sentences) >= 2 {
		if seg, ok := sentences[1].([]any); ok && len(seg) > 2 {
			if p, ok := seg[2].(string); ok && strings.TrimSpace(p) != "" {
				phonetic = p
			}
		}
	}
	return text, phonetic, nil
}
