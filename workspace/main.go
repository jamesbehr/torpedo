package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const configDir = ".torpedo"

type Configuration struct {
}

type Manager struct {
	Path string
}

func Create(path string) error {
	if err := os.MkdirAll(filepath.Join(path, configDir), 0744); err != nil {
		return fmt.Errorf("workspace: failed to create workspace: %w", err)
	}

	f, err := os.OpenFile(filepath.Join(path, configDir, "config"), os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("workspace: failed to create workspace config: %w", err)
	}

	cfg := Configuration{}

	enc := toml.NewEncoder(f)
	if err := enc.Encode(cfg); err != nil {
		return fmt.Errorf("workspace: failed to serialize workspace config: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("workspace: failed to close workspace config: %w", err)
	}

	return nil
}
