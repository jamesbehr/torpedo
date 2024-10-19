package picker

import (
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

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("picker: cannot create input pipe: %w", err)
	}

	errc := make(chan error, 1)
	go func() {
		defer stdin.Close()
		defer close(errc)

		data := []byte(strings.Join(items, "\n"))

		if _, err := stdin.Write(data); err != nil {
			errc <- fmt.Errorf("picker: error writing to stdin: %w", err)
			return
		}
	}()

	chosen, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("picker: error while running the command: %w", err)
	}

	if err, ok := <-errc; ok {
		return "", err
	}

	return strings.TrimSuffix(string(chosen), "\n"), nil
}
