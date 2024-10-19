package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Workspace struct {
	Path       string
	Jumplisted bool
}

type Configuration struct {
	Workspaces []*Workspace
}

type File struct {
	Configuration
	Path string
}

func Load(path string) (*File, error) {
	var cfg Configuration
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("config: could not parse config: %w", err)
		}
	}

	return &File{cfg, path}, nil
}

func (f *File) Save() error {
	if err := os.MkdirAll(filepath.Dir(f.Path), 0744); err != nil {
		return fmt.Errorf("config: failed to create config directory: %w", err)
	}

	fh, err := os.Create(f.Path)
	if err != nil {
		return fmt.Errorf("config: could not open config for writing: %w", err)
	}

	if err := toml.NewEncoder(fh).Encode(f.Configuration); err != nil {
		return fmt.Errorf("config: could not write config: %w", err)
	}

	if err := fh.Close(); err != nil {
		return fmt.Errorf("config: could not write config: %w", err)
	}

	return nil
}
