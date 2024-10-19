package editor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Editor() string {
	if v, ok := os.LookupEnv("TERM"); ok && v != "dumb" {
		v, ok := os.LookupEnv("VISUAL")
		if ok {
			return v
		}
	}

	v, ok := os.LookupEnv("EDITOR")
	if ok {
		return v
	}

	return "vi"
}

func Edit(path string) error {
	cmd := exec.Command(Editor(), path)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor: unable to start editor: %w", err)
	}

	return nil
}

func EditList(path string, items []string) ([]string, error) {
	data := strings.Join(items, "\n")
	if err := os.WriteFile(path, []byte(data), 0744); err != nil {
		return nil, fmt.Errorf("editor: unable to write list to file: %w", err)
	}

	if err := Edit(path); err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("editor: unable to read list: %w", err)
	}

	return strings.Split(string(bytes), "\n"), nil
}
