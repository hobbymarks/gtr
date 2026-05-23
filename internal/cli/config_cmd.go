package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/hobbymarks/gtr/internal/config"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "View and manage gtr configuration (~/.gtrrc)",
		Long: strings.TrimSpace(`
Manage gtr configuration stored in ~/.gtrrc.

Without arguments, displays all current settings including effective values
from environment variables.`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showConfig(cmd)
		},
	}

	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigUnsetCmd())
	cmd.AddCommand(newConfigPathCmd())

	return cmd
}

func showConfig(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	path := configFilePath()

	fmt.Fprintf(out, "Config file: %s\n", path)
	if path == "" {
		fmt.Fprintln(out, "(could not determine home directory)")
	} else if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Fprintln(out, "(file does not exist)")
	} else if err != nil {
		fmt.Fprintf(out, "(error: %v)\n", err)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Key                  File value    Effective value")
	fmt.Fprintln(out, "---                  ----------    ---------------")

	keys := config.KnownConfigKeys()
	for _, k := range keys {
		fileVal := config.ConfigFileValueForPath(path, k)
		envVal := os.Getenv(k)
		effective := envVal
		if effective == "" {
			effective = fileVal
		}
		if effective == "" {
			effective = "-"
		}
		if fileVal == "" {
			fileVal = "-"
		}
		fmt.Fprintf(out, "%-20s %-12s  %s", k, fileVal, effective)
		if envVal != "" {
			fmt.Fprintf(out, "  (from env)")
		}
		fmt.Fprintln(out)
	}
	return nil
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long:  "Set a configuration key in ~/.gtrrc. Supported keys: GTR_DEFAULT_ENGINE, GTR_DEFAULT_TARGET, GTR_TIMEOUT.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := strings.TrimSpace(args[0])
			value := strings.TrimSpace(args[1])
			if !config.IsKnownConfigKey(key) {
				return fmt.Errorf("unknown config key %q (supported: %s)", key, strings.Join(config.KnownConfigKeys(), ", "))
			}
			if err := configSet(key, value); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Set %s=%s\n", key, value)
			return nil
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := strings.TrimSpace(args[0])
			if !config.IsKnownConfigKey(key) {
				return fmt.Errorf("unknown config key %q (supported: %s)", key, strings.Join(config.KnownConfigKeys(), ", "))
			}
			val := configFileValue(key)
			if val == "" {
				fmt.Fprintf(cmd.OutOrStdout(), "%s is not set\n", key)
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), val)
			return nil
		},
	}
}

func newConfigUnsetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unset <key>",
		Short: "Remove a configuration key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := strings.TrimSpace(args[0])
			if !config.IsKnownConfigKey(key) {
				return fmt.Errorf("unknown config key %q (supported: %s)", key, strings.Join(config.KnownConfigKeys(), ", "))
			}
			if err := configUnset(key); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed %s\n", key)
			return nil
		},
	}
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show the config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := configFilePath()
			if p == "" {
				return fmt.Errorf("could not determine config file location")
			}
			fmt.Fprintln(cmd.OutOrStdout(), p)
			return nil
		},
	}
}

func configFilePath() string {
	if f := configFilePathFn; f != nil {
		return f()
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".gtrrc")
}

var configFilePathFn func() string

func configFileValue(key string) string {
	path := configFilePath()
	if path == "" {
		return ""
	}
	return config.ConfigFileValueForPath(path, key)
}

func configSet(key, value string) error {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)

	path := configFilePath()
	if path == "" {
		return fmt.Errorf("cannot determine config file location")
	}

	lines, err := readConfigLines(path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	found := false

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == key {
			lines[i] = key + "=" + value
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, key+"="+value)
	}

	return writeConfigLines(path, lines)
}

func configUnset(key string) error {
	key = strings.TrimSpace(key)
	path := configFilePath()
	if path == "" {
		return fmt.Errorf("cannot determine config file location")
	}

	lines, err := readConfigLines(path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	var out []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			out = append(out, line)
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == key {
			continue
		}
		out = append(out, line)
	}

	return writeConfigLines(path, out)
}

func readConfigLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	return lines, nil
}

func writeConfigLines(path string, lines []string) error {
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	content := strings.Join(lines, "\n") + "\n"

	f, err := os.CreateTemp(filepath.Dir(path), ".gtrrc-*")
	if err != nil {
		return err
	}
	tmpPath := f.Name()
	if _, err := f.Write([]byte(content)); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}
