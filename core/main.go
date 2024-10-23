package core

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesbehr/torpedo/filesystem"
	"github.com/jamesbehr/torpedo/tmux"
)

var (
	ErrNotInRoot      = errors.New("core: not in root")
	ErrNotInWorkspace = errors.New("core: not in workspace")
)

func encode(fs filesystem.FS, name string, v any) error {
	if err := fs.MkdirAll(filepath.Dir(name), 0777); err != nil {
		return err
	}

	f, err := fs.Create(name)
	if err != nil {
		return err
	}

	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	if err := enc.Encode(v); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

func decode(fs filesystem.FS, name string, v any) error {
	f, err := fs.Open(name)
	if err != nil {
		return err
	}

	defer f.Close()

	if err := json.NewDecoder(f).Decode(v); err != nil {
		return err
	}

	return nil
}

type workspace struct {
	Name       string `json:"path"`
	Jumplisted bool   `json:"jumplisted"`
}

type state struct {
	Workspaces []*workspace `json:"workspaces"`
}

func (state *state) Save(fs filesystem.FS) error {
	if err := encode(fs, stateName, state); err != nil {
		return fmt.Errorf("core: failed to save state: %w", err)
	}

	return nil
}

type project struct {
	Name       string `json:"name"`
	Jumplisted bool   `json:"jumplisted"`
}

type workspaceConfig struct {
	Projects []*project `json:"projects"`
}

func (cfg *workspaceConfig) Save(fs filesystem.FS) error {
	if err := encode(fs, workspaceConfigName, cfg); err != nil {
		return fmt.Errorf("core: failed to save workspace config: %w", err)
	}

	return nil
}

const (
	stateName           = ".torpedo.json"
	configName          = "torpedo/config.json"
	workspaceConfigName = "torpedo.json"
)

type currentWorkspace struct {
	name   string
	config workspaceConfig
}

type Service struct {
	rootFs           filesystem.FS
	currentWorkspace *currentWorkspace
	state            state
}

func Scan() (*Service, error) {
	svc := Service{}

	home := os.Getenv("HOME")
	if home == "" {
		return nil, errors.New("core: $HOME not set")
	}

	svc.rootFs = filesystem.DirFS(home)

	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("core: could not get working directory: %w", err)
	}

	rel, err := filepath.Rel(home, pwd)
	if err != nil {
		return nil, ErrNotInRoot
	}

	if strings.HasPrefix(rel, "..") {
		return nil, ErrNotInRoot
	}

	for dir := rel; dir != "."; dir = filepath.Dir(dir) {
		children, err := svc.rootFs.ReadDir(dir)
		if err != nil {
			return nil, err
		}

		isWorkspace := false
		for _, child := range children {
			if child.Name() == workspaceConfigName && child.Type().IsRegular() {
				isWorkspace = true
				break
			}
		}

		if isWorkspace {
			svc.currentWorkspace = &currentWorkspace{
				name:   dir,
				config: workspaceConfig{},
			}

			if err := decode(svc.rootFs, filepath.Join(dir, workspaceConfigName), &svc.currentWorkspace.config); err != nil {
				return nil, fmt.Errorf("core: unable to read workspace config: %w", err)
			}
		}
	}

	if err := decode(svc.rootFs, stateName, &svc.state); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &svc, nil
		}

		return nil, fmt.Errorf("core: unable to read state: %w", err)
	}

	return &svc, nil
}

func (svc *Service) CreateProject(name string) error {
	if svc.currentWorkspace == nil {
		return ErrNotInWorkspace
	}

	if err := svc.rootFs.MkdirAll(filepath.Join(svc.currentWorkspace.name, name), 0777); err != nil {
		return fmt.Errorf("core: could not create project dir: %w", err)
	}

	svc.currentWorkspace.config.Projects = append(svc.currentWorkspace.config.Projects, &project{
		Name:       name,
		Jumplisted: false,
	})

	return svc.currentWorkspace.config.Save(svc.rootFs)
}

func (svc *Service) AttachProjectSession(name string, t *tmux.Client) error {
	if svc.currentWorkspace == nil {
		return ErrNotInWorkspace
	}

	fullName := filepath.Join(svc.currentWorkspace.name, name)

	path, err := svc.rootFs.Path(fullName)
	if err != nil {
		return fmt.Errorf("core: failed to get workspace dir: %w", err)
	}

	session := tmux.Session{
		Name: fullName,
		Path: path,
	}

	if err = t.AttachOrCreateSession(&session); err != nil {
		return fmt.Errorf("core: could not attach to session: %w", err)
	}

	return nil
}

func (svc *Service) ProjectJumplist() (*Jumplist, error) {
	if svc.currentWorkspace == nil {
		return nil, ErrNotInWorkspace
	}

	jl := Jumplist{
		Header: []string{
			"Select the projects the you want to be present in your jumplist",
			"Blank lines or lines starting with '#' will be ignored. The first non-empty line",
			"will be the the 0th item in the jumplist.",
		},
	}

	for _, p := range svc.currentWorkspace.config.Projects {
		if p.Jumplisted {
			jl.Items = append(jl.Items, p.Name)
		} else {
			jl.Trailer = append(jl.Trailer, p.Name)
		}
	}

	return &jl, nil
}

func (svc *Service) UpdateProjectJumplist(paths []string) error {
	if svc.currentWorkspace == nil {
		return ErrNotInWorkspace
	}

	for _, p := range svc.currentWorkspace.config.Projects {
		p.Jumplisted = false
	}

	for _, path := range paths {
		found := false
		for _, p := range svc.currentWorkspace.config.Projects {
			if p.Name == path {
				p.Jumplisted = true
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("core: workspace does not exist %q", path)
		}

	}

	return svc.currentWorkspace.config.Save(svc.rootFs)
}

func (svc *Service) ProjectNames() ([]string, error) {
	if svc.currentWorkspace == nil {
		return nil, ErrNotInWorkspace
	}

	names := make([]string, len(svc.currentWorkspace.config.Projects))

	for i, p := range svc.currentWorkspace.config.Projects {
		names[i] = p.Name
	}

	return names, nil
}

func (svc *Service) CreateWorkspace(name string) error {
	fs, err := svc.rootFs.Sub(name)
	if err != nil {
		return fmt.Errorf("core: invalid workspace: %w", err)
	}

	if err := fs.MkdirAll(".", 0777); err != nil {
		return fmt.Errorf("core: failed to create workspace dir: %w", err)
	}

	cfg := workspaceConfig{}
	if err := cfg.Save(fs); err != nil {
		return fmt.Errorf("core: failed to write workspace config: %w", err)
	}

	svc.state.Workspaces = append(svc.state.Workspaces, &workspace{
		Name:       name,
		Jumplisted: false,
	})

	return svc.state.Save(svc.rootFs)
}

func (svc *Service) AttachWorkspaceSession(name string, t *tmux.Client) error {
	path, err := svc.rootFs.Path(name)
	if err != nil {
		return fmt.Errorf("core: failed to get workspace dir: %w", err)
	}

	session := tmux.Session{
		Name: name,
		Path: path,
	}

	if err = t.AttachOrCreateSession(&session); err != nil {
		return fmt.Errorf("core: could not attach to session: %w", err)
	}

	return nil
}

func (svc *Service) WorkspaceJumplist() *Jumplist {
	jl := Jumplist{
		Header: []string{
			"Select the workspaces the you want to be present in your jumplist",
			"Blank lines or lines starting with '#' will be ignored. The first non-empty line",
			"will be the the 0th item in the jumplist.",
		},
	}

	for _, ws := range svc.state.Workspaces {
		if ws.Jumplisted {
			jl.Items = append(jl.Items, ws.Name)
		} else {
			jl.Trailer = append(jl.Trailer, ws.Name)
		}
	}

	return &jl
}

func (svc *Service) UpdateWorkspaceJumplist(paths []string) error {
	for _, ws := range svc.state.Workspaces {
		ws.Jumplisted = false
	}

	for _, path := range paths {
		found := false
		for _, ws := range svc.state.Workspaces {
			if ws.Name == path {
				ws.Jumplisted = true
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("core: workspace does not exist %q", path)
		}

	}

	return svc.state.Save(svc.rootFs)
}

func (svc *Service) WorkspaceNames() []string {
	names := make([]string, len(svc.state.Workspaces))

	for i, ws := range svc.state.Workspaces {
		names[i] = ws.Name
	}

	return names
}

type Jumplist struct {
	Header  []string
	Items   []string
	Trailer []string
}

func (j *Jumplist) Get(index int) (string, error) {
	if index < len(j.Items) {
		return j.Items[index], nil
	}

	return "", fmt.Errorf("jumplist: item does not exist %d", index)
}

func (j *Jumplist) Serialize(w io.Writer) error {
	for _, line := range j.Header {
		if _, err := fmt.Fprintf(w, "# %s\n", line); err != nil {
			return fmt.Errorf("jumplist: failed to serialize: %w", err)
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("jumplist: failed to serialize: %w", err)
	}

	for _, line := range j.Items {
		if _, err := fmt.Fprintf(w, "%s\n", line); err != nil {
			return fmt.Errorf("jumplist: failed to serialize: %w", err)
		}
	}

	for _, line := range j.Trailer {
		if _, err := fmt.Fprintf(w, "# %s\n", line); err != nil {
			return fmt.Errorf("jumplist: failed to serialize: %w", err)
		}
	}

	return nil
}

func ParseJumplist(r io.Reader) ([]string, error) {
	result := []string{}

	s := bufio.NewScanner(r)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		result = append(result, line)
	}

	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("jumplist: failed to unmarshal: %w", err)
	}

	return result, nil
}
