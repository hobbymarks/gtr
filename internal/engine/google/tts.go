package google

import (
	"fmt"
	"net/url"
)

// BuildTTSURL returns the public Google Translate TTS endpoint URL (translate-shell parity).
func BuildTTSURL(text, tl string) (string, error) {
	if text == "" {
		return "", fmt.Errorf("google: empty TTS text")
	}
	if tl == "" {
		return "", fmt.Errorf("google: empty target language for TTS")
	}
	v := url.Values{}
	v.Set("ie", "UTF-8")
	v.Set("client", "gtx")
	v.Set("tl", tl)
	v.Set("q", text)
	return "https://translate.googleapis.com/translate_tts?" + v.Encode(), nil
}
