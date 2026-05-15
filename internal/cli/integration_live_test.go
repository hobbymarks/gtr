//go:build integration

package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hobbymarks/gtr/internal/engine"
	_ "github.com/hobbymarks/gtr/internal/engine/apertium"
	_ "github.com/hobbymarks/gtr/internal/engine/auto"
	_ "github.com/hobbymarks/gtr/internal/engine/bing"
	_ "github.com/hobbymarks/gtr/internal/engine/google"
	_ "github.com/hobbymarks/gtr/internal/engine/yandex"
)

func TestIntegrationGoogleTranslate(t *testing.T) {
	factory, ok := engine.Lookup("google")
	if !ok {
		t.Fatal("google engine not registered")
	}
	eng, err := factory()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	out, err := eng.Translate(ctx, engine.TranslateInput{
		Text:   "hello",
		Source: "en",
		Target: "fr",
		Brief:  true,
	})
	if err != nil {
		t.Fatalf("google translate: %v", err)
	}
	if strings.TrimSpace(out.Text) == "" {
		t.Fatal("empty translation")
	}
	t.Logf("translate en->fr: %q", out.Text)
}

func TestIntegrationGoogleIdentify(t *testing.T) {
	factory, ok := engine.Lookup("google")
	if !ok {
		t.Fatal("google engine not registered")
	}
	eng, err := factory()
	if err != nil {
		t.Fatal(err)
	}
	li, ok := eng.(engine.LanguageIdentifier)
	if !ok {
		t.Fatal("google does not implement LanguageIdentifier")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	lang, err := li.IdentifyLanguage(ctx, "Bonjour le monde", "en")
	if err != nil {
		t.Fatal(err)
	}
	if lang == "" {
		t.Fatal("empty detected language")
	}
	t.Logf("detected: %s", lang)
}

func TestIntegrationAutoRouter(t *testing.T) {
	factory, ok := engine.Lookup("auto")
	if !ok {
		t.Fatal("auto engine not registered")
	}
	eng, err := factory()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	out, err := eng.Translate(ctx, engine.TranslateInput{
		Text:   "good morning",
		Source: "en",
		Target: "de",
		Brief:  true,
	})
	if err != nil {
		t.Fatalf("auto translate: %v", err)
	}
	if strings.TrimSpace(out.Text) == "" {
		t.Fatal("empty translation")
	}
	t.Logf("translate en->de: %q", out.Text)
}

func TestIntegrationListEngines(t *testing.T) {
	names := engine.Names()
	if len(names) == 0 {
		t.Fatal("no engines registered")
	}
	t.Logf("engines: %v", names)
}

func TestIntegrationJSONOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRoot()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"-e", "google", "-t", "es", "-b", "--json", "hello"})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := cmd.ExecuteContext(ctx)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"text"`) {
		t.Fatalf("no text field in JSON output: %s", output)
	}
	t.Logf("JSON: %s", output)
}
