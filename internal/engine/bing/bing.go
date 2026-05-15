// Package bing implements the Bing Web Translator engine for gtr,
// using HTML scraping for setup tokens and the ttranslatev3 POST endpoint.
package bing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hobbymarks/gtr/internal/engine"
	"github.com/hobbymarks/gtr/internal/httpx"
)

const translateBase = "https://www.bing.com"

const setupCacheTTL = 5 * time.Minute

type setupCacheEntry struct {
	cookie, ig, iid, token, key string
	expires                     time.Time
}

// Engine uses Bing Web Translator (translate-shell parity: www.bing.com).
type Engine struct {
	HTTP       *http.Client
	mu         sync.Mutex
	setupCache map[string]*setupCacheEntry
	cacheMu    sync.Mutex
}

func New(c *http.Client) *Engine {
	if c == nil {
		c = httpx.NewClient()
	}
	return &Engine{HTTP: c, setupCache: make(map[string]*setupCacheEntry)}
}

func (e *Engine) Name() string { return "bing" }

func (e *Engine) Translate(ctx context.Context, in engine.TranslateInput) (engine.TranslateOutput, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	hosts := []string{translateBase, "https://cn.bing.com"}
	var errs []error
	for _, origin := range hosts {
		out, err := e.translateOnHost(ctx, origin, in)
		if err == nil {
			return out, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", origin, err))
	}
	if len(errs) == 0 {
		return engine.TranslateOutput{}, fmt.Errorf("bing: no host succeeded")
	}
	return engine.TranslateOutput{}, fmt.Errorf("bing: all hosts failed: %v", errs)
}

// translateRequest performs setup and the translate POST, returning the raw body and HTTP status.
func (e *Engine) translateRequest(ctx context.Context, origin, text, sl, tl string, debug bool) (body []byte, statusCode int, err error) {
	cookie, ig, iid, token, key, err := e.cachedSetup(ctx, origin)
	if err != nil {
		return nil, 0, err
	}

	patchLangCodes(&sl, &tl)

	postURL := origin + "/ttranslatev3?IG=" + url.QueryEscape(ig) + "&IID=" + url.QueryEscape(iid)
	form := url.Values{}
	form.Set("text", text)
	form.Set("fromLang", sl)
	form.Set("to", tl)
	form.Set("token", token)
	form.Set("key", key)
	bodyStr := "&" + form.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, postURL, strings.NewReader(bodyStr))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", origin+"/translator")
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	if debug {
		_, _ = fmt.Fprintf(os.Stderr, "bing debug: POST %s\n", postURL)
	}

	resp, err := e.HTTP.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("bing: POST: %w", err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, engine.MaxReadBody+1)
	body, err = io.ReadAll(limited)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("bing: read body: %w", err)
	}
	if int64(len(body)) > engine.MaxReadBody {
		return nil, resp.StatusCode, fmt.Errorf("bing: response body exceeds %d bytes", engine.MaxReadBody)
	}
	return body, resp.StatusCode, nil
}

func (e *Engine) translateOnHost(ctx context.Context, origin string, in engine.TranslateInput) (engine.TranslateOutput, error) {
	body, statusCode, err := e.translateRequest(ctx, origin, in.Text, in.Source, in.Target, in.Debug)
	if err != nil {
		return engine.TranslateOutput{}, err
	}

	if in.Dump {
		return engine.TranslateOutput{Text: string(body)}, nil
	}

	if statusCode == http.StatusTooManyRequests {
		return engine.TranslateOutput{}, fmt.Errorf("bing: rate limiting is in effect (HTTP %d)", statusCode)
	}
	if statusCode >= 400 {
		return engine.TranslateOutput{}, fmt.Errorf("bing: HTTP %d: %s", statusCode, httpx.Truncate(body, 200))
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return engine.TranslateOutput{}, fmt.Errorf("bing: empty response body")
	}

	text, err := parseTranslateResponse(body)
	if err != nil {
		return engine.TranslateOutput{}, err
	}
	if in.Brief {
		text = strings.TrimSpace(text)
	}
	return engine.TranslateOutput{Text: text}, nil
}

var (
	reIG    = regexp.MustCompile(`IG:"([^"]+)"`)
	reIID   = regexp.MustCompile(`data-iid="([^"]+)"`)
	reAbuse = regexp.MustCompile(`params_AbusePreventionHelper\s*=\s*(\[[^\]]+\]);`)
)

func (e *Engine) cachedSetup(ctx context.Context, origin string) (cookie, ig, iid, token, key string, err error) {
	e.cacheMu.Lock()
	if entry, ok := e.setupCache[origin]; ok && time.Now().Before(entry.expires) {
		cookie, ig, iid, token, key = entry.cookie, entry.ig, entry.iid, entry.token, entry.key
		e.cacheMu.Unlock()
		return
	}
	e.cacheMu.Unlock()

	cookie, ig, iid, token, key, err = e.setup(ctx, origin)
	if err != nil {
		return "", "", "", "", "", err
	}

	e.cacheMu.Lock()
	e.setupCache[origin] = &setupCacheEntry{
		cookie:  cookie,
		ig:      ig,
		iid:     iid,
		token:   token,
		key:     key,
		expires: time.Now().Add(setupCacheTTL),
	}
	e.cacheMu.Unlock()
	return
}

func (e *Engine) setup(ctx context.Context, origin string) (cookie, ig, iid, token, key string, err error) {
	pageURL := origin + "/translator"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", "", "", "", "", err
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Referer", pageURL)

	resp, err := e.HTTP.Do(req)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("bing setup GET: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		limitedErr := io.LimitReader(resp.Body, 200)
		errBody, _ := io.ReadAll(limitedErr)
		return "", "", "", "", "", fmt.Errorf("bing setup GET: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(errBody)))
	}

	var setCookies []string
	for _, c := range resp.Cookies() {
		if c.Name != "" {
			setCookies = append(setCookies, c.Name+"="+c.Value)
		}
	}
	cookie = strings.Join(setCookies, "; ")

	body, err := io.ReadAll(io.LimitReader(resp.Body, engine.MaxReadBody+1))
	if err != nil {
		return "", "", "", "", "", err
	}
	if int64(len(body)) > engine.MaxReadBody {
		return "", "", "", "", "", fmt.Errorf("bing setup GET: response body exceeds %d bytes", engine.MaxReadBody)
	}
	html := string(body)

	if m := reIG.FindStringSubmatch(html); len(m) > 1 {
		ig = m[1]
	}
	if m := reIID.FindStringSubmatch(html); len(m) > 1 {
		iid = m[1]
	}
	if ig == "" || iid == "" {
		return "", "", "", "", "", fmt.Errorf("bing: could not parse IG/IID from translator page")
	}

	if m := reAbuse.FindStringSubmatch(html); len(m) > 1 {
		var arr []interface{}
		if err := json.Unmarshal([]byte(m[1]), &arr); err == nil && len(arr) >= 2 {
			key = fmt.Sprint(arr[0])
			token = fmt.Sprint(arr[1])
		}
	}
	if token == "" || key == "" {
		return "", "", "", "", "", fmt.Errorf("bing: could not parse token/key from translator page")
	}
	return cookie, ig, iid, token, key, nil
}

func patchLangCodes(sl, tl *string) {
	patchFromLang(sl)
	patchToLang(tl)
}

func patchFromLang(sl *string) {
	switch *sl {
	case "auto":
		*sl = "auto-detect"
	case "tl":
		*sl = "fil"
	case "hmn":
		*sl = "mww"
	case "ku":
		*sl = "kmr"
	case "ckb":
		*sl = "ku"
	case "mn":
		*sl = "mn-Cyrl"
	case "no":
		*sl = "nb"
	case "pt-BR":
		*sl = "pt"
	case "pt-PT":
		*sl = "pt"
	case "zh-CN":
		*sl = "zh-Hans"
	case "zh-TW":
		*sl = "zh-Hant"
	}
}

func patchToLang(tl *string) {
	switch *tl {
	case "tl":
		*tl = "fil"
	case "hmn":
		*tl = "mww"
	case "ku":
		*tl = "kmr"
	case "ckb":
		*tl = "ku"
	case "mn":
		*tl = "mn-Cyrl"
	case "no":
		*tl = "nb"
	case "pt-BR":
		*tl = "pt"
	case "pt-PT":
		*tl = "pt-pt"
	case "zh-CN":
		*tl = "zh-Hans"
	case "zh-TW":
		*tl = "zh-Hant"
	}
}

func parseTranslateResponse(raw []byte) (string, error) {
	text, _, err := parseBingResponse(raw)
	return text, err
}

// parseBingResponse extracts translation text and optional detected source language.
func parseBingResponse(raw []byte) (text string, detected string, err error) {
	var root []map[string]interface{}
	if err := json.Unmarshal(raw, &root); err != nil {
		return "", "", fmt.Errorf("bing: invalid JSON: %w", err)
	}
	if len(root) == 0 {
		return "", "", fmt.Errorf("bing: unexpected JSON root")
	}
	first := root[0]
	if dl, ok := first["detectedLanguage"].(map[string]interface{}); ok {
		if lang, ok := dl["language"].(string); ok {
			detected = lang
		}
	}
	if sc, ok := first["statusCode"].(float64); ok && sc == 400 {
		return "", "", fmt.Errorf("bing: does not support the specified language(s)")
	}
	trans, ok := first["translations"].([]interface{})
	if !ok || len(trans) == 0 {
		return "", "", fmt.Errorf("bing: missing translations in response")
	}
	t0, ok := trans[0].(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("bing: malformed translations[0]")
	}
	txt, ok := t0["text"].(string)
	if !ok || strings.TrimSpace(txt) == "" {
		return "", "", fmt.Errorf("bing: empty translation text")
	}
	return txt, detected, nil
}
