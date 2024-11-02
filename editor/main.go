package editor

import (
	"fmt"
	"os"
	"os/exec"
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

func Open(path string) error {
	cmd := exec.Command(Editor(), path)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor: unable to start editor: %w", err)
	}

	return nil
}
