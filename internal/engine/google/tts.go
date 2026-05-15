package google

import (
	"fmt"
	"net/url"
	"strings"
)

// maxTTSTextLen caps the text sent to Google TTS to avoid URL-length issues
// and degraded playback quality with very long inputs.
const maxTTSTextLen = 1500

// BuildTTSURL returns the public Google Translate TTS endpoint URL (translate-shell parity).
func BuildTTSURL(text, tl string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("google: empty TTS text")
	}
	if tl == "" {
		return "", fmt.Errorf("google: empty target language for TTS")
	}
	if len(text) > maxTTSTextLen {
		text = text[:maxTTSTextLen]
	}
	v := url.Values{}
	v.Set("ie", "UTF-8")
	v.Set("client", "gtx")
	v.Set("tl", tl)
	v.Set("q", text)
	return "https://translate.googleapis.com/translate_tts?" + v.Encode(), nil
}
