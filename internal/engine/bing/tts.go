package bing

import (
	"fmt"
	"net/url"
	"strings"
)

const maxTTSTextLen = 1500

// BuildTTSURL returns the Bing TTS endpoint URL (translate-shell parity).
func BuildTTSURL(text, tl string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("bing: empty TTS text")
	}
	if tl == "" {
		return "", fmt.Errorf("bing: empty target language for TTS")
	}
	if len(text) > maxTTSTextLen {
		text = text[:maxTTSTextLen]
	}
	v := url.Values{}
	v.Set("format", "audio/mp3")
	v.Set("language", tl)
	v.Set("text", text)
	return "https://www.bing.com/tspeak?" + v.Encode(), nil
}
