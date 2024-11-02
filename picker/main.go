package picker

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Pick(items []string) (string, error) {
	path, ok := os.LookupEnv("TORPEDO_PICKER")
	if !ok {
		path = "fzf"
	}

	cmd := exec.Command(path)
	cmd.Stderr = os.Stderr
	cmd.Stdin = bytes.NewBuffer([]byte(strings.Join(items, "\n")))

	chosen, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("picker: error while running the command: %w", err)
	}

	needle := strings.TrimSuffix(string(chosen), "\n")
	for _, item := range items {
		if needle == item {
			return item, nil
		}
	}

	return "", errors.New("picker: command returned invalid item")
}
