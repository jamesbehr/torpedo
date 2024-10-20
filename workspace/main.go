package workspace

import (
	"encoding/json"
	"fmt"

	"github.com/jamesbehr/torpedo/filesystem"
	"github.com/jamesbehr/torpedo/tmux"
)

type Configuration struct {
}

type Service struct {
	FS     filesystem.FS
	Config Configuration
}

const configDir = ".torpedo"
const configName = ".torpedo/workspace.json"

func New(fs filesystem.FS) (*Service, error) {
	var cfg Configuration

	f, err := fs.Open(configName)
	if err != nil {
		return nil, fmt.Errorf("root: unable to open config file: %w", err)
	}

	defer f.Close()

	dec := json.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("root: unable to deserialize config: %w", err)
	}

	svc := Service{
		FS:     fs,
		Config: cfg,
	}

	return &svc, nil
}

func Initialize(fs filesystem.FS) (*Service, error) {
	svc := Service{
		FS:     fs,
		Config: Configuration{},
	}

	if err := svc.FS.MkdirAll(configDir, 0777); err != nil {
		return nil, fmt.Errorf("workspace: failed to create workspace: %w", err)
	}

	f, err := svc.FS.Create(configName)
	if err != nil {
		return nil, fmt.Errorf("workspace: failed to create workspace config: %w", err)
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	if err := enc.Encode(svc.Config); err != nil {
		return nil, fmt.Errorf("workspace: failed to serialize workspace config: %w", err)
	}

	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("workspace: failed to close workspace config: %w", err)
	}

	return &svc, nil
}

func (svc *Service) AttachWorkspaceSession(t *tmux.Client) error {
	path, err := svc.FS.Path(".")
	if err != nil {
		return fmt.Errorf("workspace: failed to get workspace dir: %w", err)
	}

	session := tmux.Session{
		Name: path,
		Path: path,
		Environment: []string{
			"TORPEDO_WORKSPACE=" + path,
		},
	}

	if err = t.AttachOrCreateSession(&session); err != nil {
		return fmt.Errorf("workspace: could not attach to session: %w", err)
	}

	return nil
}
