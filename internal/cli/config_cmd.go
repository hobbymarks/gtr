package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
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
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Fprintln(out, "(file does not exist)")
	} else if err != nil {
		fmt.Fprintf(out, "(error: %v)\n", err)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Key                  File value    Effective value")
	fmt.Fprintln(out, "---                  ----------    ---------------")

	keys := []string{"GTR_DEFAULT_ENGINE", "GTR_DEFAULT_TARGET", "GTR_TIMEOUT"}
	for _, k := range keys {
		fileVal := configFileValue(k)
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
			fmt.Fprintln(cmd.OutOrStdout(), configFilePath())
			return nil
		},
	}
}

func configFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "~/.gtrrc"
	}
	return filepath.Join(home, ".gtrrc")
}

func configFileValue(key string) string {
	path := configFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) != key {
			continue
		}
		return strings.TrimSpace(parts[1])
	}
	return ""
}

func configSet(key, value string) error {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)

	path := configFilePath()
	lines := readConfigLines(path)
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
	lines := readConfigLines(path)

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

func readConfigLines(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	return lines
}

func writeConfigLines(path string, lines []string) error {
	// Trim trailing blank lines
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(path, []byte(content), 0644)
}
