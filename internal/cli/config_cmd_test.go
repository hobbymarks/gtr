package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func testConfigRoot() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.AddCommand(newConfigCmd())
	return cmd
}

func TestConfig_path(t *testing.T) {
	cmd := testConfigRoot()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"config", "path"})

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	got := strings.TrimSpace(out.String())
	if runtime.GOOS == "windows" {
		if !strings.Contains(got, ".gtrrc") {
			t.Fatalf("expected .gtrrc in path, got %q", got)
		}
	} else {
		if !strings.HasSuffix(got, ".gtrrc") {
			t.Fatalf("expected .gtrrc suffix, got %q", got)
		}
	}
}

func TestConfig_get_unset(t *testing.T) {
	cmd := testConfigRoot()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"config", "get", "GTR_DEFAULT_TARGET"})

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "is not set") {
		t.Fatalf("expected 'is not set', got %q", out.String())
	}
}

func TestConfig_show(t *testing.T) {
	cmd := testConfigRoot()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"config"})

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Key") {
		t.Fatalf("expected table header, got %q", out.String())
	}
}

func TestConfig_set(t *testing.T) {
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	cmd := testConfigRoot()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"config", "set", "GTR_DEFAULT_TARGET", "de"})

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Set GTR_DEFAULT_TARGET=de") {
		t.Fatalf("expected set confirmation, got %q", out.String())
	}

	// Verify file was written
	data, err := os.ReadFile(filepath.Join(tmp, ".gtrrc"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "GTR_DEFAULT_TARGET=de") {
		t.Fatalf("expected GTR_DEFAULT_TARGET=de in config file, got %q", string(data))
	}
}

func TestConfig_unset(t *testing.T) {
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	// First set a value
	err := os.WriteFile(filepath.Join(tmp, ".gtrrc"), []byte("GTR_DEFAULT_TARGET=de\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cmd := testConfigRoot()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"config", "unset", "GTR_DEFAULT_TARGET"})

	err = cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Removed GTR_DEFAULT_TARGET") {
		t.Fatalf("expected unset confirmation, got %q", out.String())
	}

	// Verify file no longer has the key
	data, err := os.ReadFile(filepath.Join(tmp, ".gtrrc"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "GTR_DEFAULT_TARGET") {
		t.Fatalf("key should be removed, got %q", string(data))
	}
}

func TestConfig_set_updateExisting(t *testing.T) {
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	err := os.WriteFile(filepath.Join(tmp, ".gtrrc"), []byte("GTR_DEFAULT_TARGET=de\nGTR_TIMEOUT=10\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cmd := testConfigRoot()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetArgs([]string{"config", "set", "GTR_DEFAULT_TARGET", "fr"})

	err = cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, ".gtrrc"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "GTR_DEFAULT_TARGET=fr") {
		t.Fatalf("expected updated value, got %q", string(data))
	}
	// Should preserve other keys
	if !strings.Contains(string(data), "GTR_TIMEOUT=10") {
		t.Fatalf("should preserve GTR_TIMEOUT, got %q", string(data))
	}
	// Should not have the old value
	if strings.Contains(string(data), "GTR_DEFAULT_TARGET=de") {
		t.Fatalf("old value should be replaced, got %q", string(data))
	}
}

func TestConfig_envOverride(t *testing.T) {
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	os.Setenv("GTR_DEFAULT_TARGET", "ja")
	defer os.Unsetenv("GTR_DEFAULT_TARGET")

	err := os.WriteFile(filepath.Join(tmp, ".gtrrc"), []byte("GTR_DEFAULT_TARGET=de\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cmd := testConfigRoot()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"config"})

	err = cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "ja") {
		t.Fatalf("expected effective value from env, got %q", out.String())
	}
	if !strings.Contains(out.String(), "from env") {
		t.Fatalf("expected env marker, got %q", out.String())
	}
}
