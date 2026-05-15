package yandex

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
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

const maxReadBody = 4 << 20

const translateBase = "https://translate.yandex.net"

// Engine uses Yandex public mobile-style JSON API (translate-shell parity).
type Engine struct {
	HTTP *http.Client
	ucid string
}

// New returns a Yandex engine; each instance gets a fresh ucid like translate-shell yandexInit.
func New(c *http.Client) *Engine {
	if c == nil {
		c = httpx.NewClient()
	}
	return &Engine{HTTP: c, ucid: newUCID()}
}

func newUCID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strings.ReplaceAll("00000000-0000-0000-0000-000000000000", "-", "")
	}
	return hex.EncodeToString(b[:])
}

func (e *Engine) Name() string { return "yandex" }

func (e *Engine) translatePost(ctx context.Context, in engine.TranslateInput) (body []byte, statusCode int, err error) {
	sl, tl := stripDigraph(in.Source), stripDigraph(in.Target)
	lang := tl
	if sl != "auto" {
		lang = sl + "-" + tl
	}

	v := url.Values{}
	v.Set("ucid", e.ucid)
	v.Set("srv", "android")
	v.Set("text", in.Text)
	v.Set("lang", lang)
	postURL := translateBase + "/api/v1/tr.json/translate?" + v.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, postURL, strings.NewReader(""))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	if in.Debug {
		_, _ = fmt.Fprintf(os.Stderr, "yandex debug: POST %s\n", postURL)
	}

	resp, err := e.HTTP.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("yandex: request: %w", err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(io.LimitReader(resp.Body, maxReadBody+1))
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("yandex: read body: %w", err)
	}
	if int64(len(body)) > maxReadBody {
		return nil, resp.StatusCode, fmt.Errorf("yandex: response body exceeds %d bytes", maxReadBody)
	}
	return body, resp.StatusCode, nil
}

func (e *Engine) Translate(ctx context.Context, in engine.TranslateInput) (engine.TranslateOutput, error) {
	body, statusCode, err := e.translatePost(ctx, in)
	if err != nil {
		return engine.TranslateOutput{}, err
	}

	if in.Dump {
		return engine.TranslateOutput{Text: string(body)}, nil
	}

	if statusCode == http.StatusTooManyRequests {
		return engine.TranslateOutput{}, fmt.Errorf("yandex: rate limiting is in effect (HTTP %d)", statusCode)
	}
	if statusCode >= 400 {
		return engine.TranslateOutput{}, fmt.Errorf("yandex: HTTP %d: %s", statusCode, truncate(body, 200))
	}

	text, err := parseTranslateJSON(body)
	if err != nil {
		return engine.TranslateOutput{}, err
	}
	if in.Brief {
		text = strings.TrimSpace(text)
	}
	return engine.TranslateOutput{Text: text}, nil
}

func truncate(b []byte, n int) string {
	s := string(bytes.TrimSpace(b))
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func stripDigraph(code string) string {
	code = strings.TrimSpace(code)
	if i := strings.Index(code, "-"); i > 0 {
		return code[:i]
	}
	return code
}

func parseTranslateJSON(raw []byte) (string, error) {
	var root map[string]interface{}
	if err := json.Unmarshal(raw, &root); err != nil {
		return "", fmt.Errorf("yandex: invalid JSON: %w", err)
	}
	code := jsonNumberToInt(root["code"])
	if code != 0 && code != 200 {
		msg := ""
		if m, ok := root["message"].(string); ok {
			msg = m
		} else {
			msg = fmt.Sprint(root["message"])
		}
		if strings.TrimSpace(msg) == "" {
			msg = fmt.Sprintf("code %d", code)
		}
		return "", fmt.Errorf("yandex: %s", msg)
	}
	switch t := root["text"].(type) {
	case []interface{}:
		if len(t) == 0 {
			return "", fmt.Errorf("yandex: empty text array")
		}
		if s, ok := t[0].(string); ok {
			return s, nil
		}
		return "", fmt.Errorf("yandex: unexpected text[0] type")
	case string:
		return t, nil
	default:
		return "", fmt.Errorf("yandex: missing or invalid text field")
	}
}

func jsonNumberToInt(v interface{}) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case string:
		var n int
		_, _ = fmt.Sscanf(x, "%d", &n)
		return n
	default:
		return 0
	}
}
