package editor

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jamesbehr/torpedo/filesystem"
)

type Options struct {
	FS    filesystem.FS
	Name  string
	Items []string
	Help  string
}

func editor() string {
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
	cmd := exec.Command(editor(), path)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor: unable to start editor: %w", err)
	}

	return nil
}
