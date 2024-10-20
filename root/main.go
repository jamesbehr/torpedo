package root

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/jamesbehr/torpedo/filesystem"
	"github.com/sblinch/kdl-go"
)

type Workspace struct {
	Name       string `kdl:"name"`
	Jumplisted bool   `kdl:"jumplisted"`
}

type Configuration struct {
	Workspaces []*Workspace `kdl:"workspaces"`
}

type Service struct {
	FS     filesystem.FS
	Config Configuration
}

const configName = ".torpedo.kdl"

func New(fs filesystem.FS) (*Service, error) {
	var cfg Configuration

	f, err := fs.Open(configName)
	if err != nil {
		return nil, fmt.Errorf("root: unable to open config file: %w", err)
	}

	defer f.Close()

	dec := kdl.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("root: unable to deserialize config: %w", err)
	}

	svc := Service{
		FS:     fs,
		Config: cfg,
	}

	return &svc, nil
}

func (svc *Service) SaveConfig() error {
	f, err := svc.FS.Create(configName)
	if err != nil {
		return fmt.Errorf("root: could not open config file: %w", err)
	}

	if err := kdl.NewEncoder(f).Encode(svc.Config); err != nil {
		return fmt.Errorf("root: could not serialize config: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("root: could not write config: %w", err)
	}

	return nil
}

func (svc *Service) AddWorkspace(name string) error {
	svc.Config.Workspaces = append(svc.Config.Workspaces, &Workspace{
		Name: name,
	})

	return nil
}

func (svc *Service) WriteJumplist(w io.Writer) error {
	if len(svc.Config.Workspaces) == 0 {
		return errors.New("root: no workspaces to edit")
	}

	lines := []string{
		"# Select the workspaces that you want to be present in your jumplist. Blank",
		"# lines or lines starting with '#' will be ignored. The first non-empty line",
		"# will be the 0th item in the jumplist.",
		"",
	}

	for _, ws := range svc.Config.Workspaces {
		if ws.Jumplisted {
			lines = append(lines, ws.Name)
		}
	}

	for _, ws := range svc.Config.Workspaces {
		if !ws.Jumplisted {
			lines = append(lines, "# "+ws.Name)
		}

		ws.Jumplisted = false
	}

	data := strings.Join(lines, "\n")

	if _, err := w.Write([]byte(data)); err != nil {
		return fmt.Errorf("root: failed to write jumplist")
	}

	return nil
}

func (svc *Service) ReadJumplist(r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("root: cannot read jumplist")
	}

	lines := strings.Split(string(b), "\n")

	workspaces := []*Workspace{}

	for _, line := range lines {
		line := strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		found := false
		for _, ws := range svc.Config.Workspaces {
			if ws.Name == line {
				ws.Jumplisted = true
				workspaces = append(workspaces, ws)
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("invalid workspace: %s", line)
		}
	}

	// Keep the workspaces that weren't selected in the original order
	for _, ws := range svc.Config.Workspaces {
		if !ws.Jumplisted {
			workspaces = append(workspaces, ws)
		}
	}

	svc.Config.Workspaces = workspaces

	return nil
}

func (svc *Service) JumplistItem(index int) (*Workspace, error) {
	n := 0
	for _, ws := range svc.Config.Workspaces {
		if ws.Jumplisted {
			if n == index {
				return ws, nil
			}

			n++
		}
	}

	return nil, fmt.Errorf("config: could not find jumplist item %d", index)
}

func (svc *Service) WorkspaceDir(name string) (filesystem.FS, error) {
	fs, err := svc.FS.Sub(name)
	if err != nil {
		return nil, fmt.Errorf("root: unable to get workspace dir: %w", err)
	}

	return fs, nil
}

func (svc *Service) Workspaces() []*Workspace {
	return svc.Config.Workspaces
}
