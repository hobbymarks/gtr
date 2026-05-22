package yandex

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hobbymarks/gtr/internal/engine"
)

// IdentifyLanguage requests auto→English translation and reads the detected
// source language from the JSON "lang" field (e.g. "de-en" → "de").
func (e *Engine) IdentifyLanguage(ctx context.Context, text, hostLang string) (string, error) {
	if hostLang == "" {
		hostLang = "en"
	}
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("yandex: empty text")
	}
	in := engine.TranslateInput{
		Text:     text,
		Source:   "auto",
		Target:   "en",
		HostLang: hostLang,
		Brief:    true,
	}
	body, status, err := e.translatePost(ctx, in)
	if err != nil {
		return "", err
	}
	if status >= 400 {
		return "", fmt.Errorf("yandex: identify HTTP %d", status)
	}
	return parseDetectedSourceFromLangJSON(body)
}

func parseDetectedSourceFromLangJSON(raw []byte) (string, error) {
	lang, err := parseLangField(raw)
	if err != nil {
		return "", err
	}
	i := strings.Index(lang, "-")
	if i <= 0 || i >= len(lang)-1 {
		return "", fmt.Errorf("yandex: unexpected lang field %q", lang)
	}
	return lang[:i], nil
}

func parseLangField(raw []byte) (string, error) {
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return "", fmt.Errorf("yandex: invalid JSON: %w", err)
	}
	lang, ok := root["lang"].(string)
	if !ok || strings.TrimSpace(lang) == "" {
		return "", fmt.Errorf("yandex: missing lang in response")
	}
	return lang, nil
}

var _ engine.LanguageIdentifier = (*Engine)(nil)
