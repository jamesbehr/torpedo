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
	"slices"
	"strings"

	"github.com/jamesbehr/torpedo/filesystem"
	"github.com/jamesbehr/torpedo/tmux"
)

var (
	ErrNotInRoot      = errors.New("core: not in root")
	ErrNotInWorkspace = errors.New("core: not in workspace")
	ErrNotInProject   = errors.New("core: not in project")
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

type workspace struct{}

type state struct {
	Workspaces map[string]*workspace `json:"workspaces"`
	Jumplist   []string              `json:"jumplist"`
}

func (state *state) Save(fs filesystem.FS) error {
	if err := encode(fs, stateName, state); err != nil {
		return fmt.Errorf("core: failed to save state: %w", err)
	}

	return nil
}

type project struct {
	Jumplist []string `json:"jumplist"`
}

type workspaceConfig struct {
	Projects map[string]*project `json:"projects"`
	Jumplist []string            `json:"jumplist"`
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
	name           string
	config         workspaceConfig
	currentProject *string
}

func (ws *currentWorkspace) Save(fs filesystem.FS) error {
	subfs, err := fs.Sub(ws.name)
	if err != nil {
		return err
	}

	return ws.config.Save(subfs)
}

type Service struct {
	rootFs           filesystem.FS
	currentWorkspace *currentWorkspace
	state            state
}

func Scan() (*Service, error) {
	svc := Service{
		state: state{
			Workspaces: map[string]*workspace{},
			Jumplist:   []string{},
		},
	}

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

	// Find the current workspace, if any
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
				name: dir,
				config: workspaceConfig{
					Projects: map[string]*project{},
					Jumplist: []string{},
				},
			}

			if err := decode(svc.rootFs, filepath.Join(dir, workspaceConfigName), &svc.currentWorkspace.config); err != nil {
				return nil, fmt.Errorf("core: unable to read workspace config: %w", err)
			}
		}
	}

	// Find the current project, if any
	if svc.currentWorkspace != nil {
		for dir := rel; dir != "."; dir = filepath.Dir(dir) {
			for name := range svc.currentWorkspace.config.Projects {
				path := filepath.Join(svc.currentWorkspace.name, name)

				if dir == path {
					svc.currentWorkspace.currentProject = &name
				}
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

func (svc *Service) currentProject() (*project, error) {
	if svc.currentWorkspace == nil || svc.currentWorkspace.currentProject == nil {
		return nil, ErrNotInProject
	}

	return svc.currentWorkspace.config.Projects[*svc.currentWorkspace.currentProject], nil
}

func (svc *Service) FileJumplist() (*Jumplist, error) {
	p, err := svc.currentProject()
	if err != nil {
		return nil, err
	}

	jl := Jumplist{
		header: []string{
			"Select the files the you want to be present in your jumplist",
			"Blank lines or lines starting with '#' will be ignored. The first non-empty line",
			"will be the the 0th item in the jumplist.",
		},
	}

	for _, name := range p.Jumplist {
		jl.items = append(jl.items, name)
	}

	return &jl, nil
}

func (svc *Service) UpdateFileJumplist(names []string) error {
	p, err := svc.currentProject()
	if err != nil {
		return err
	}

	p.Jumplist = names

	return svc.currentWorkspace.Save(svc.rootFs)
}

func (svc *Service) CreateProject(name string) error {
	if svc.currentWorkspace == nil {
		return ErrNotInWorkspace
	}

	if err := svc.rootFs.MkdirAll(filepath.Join(svc.currentWorkspace.name, name), 0777); err != nil {
		return fmt.Errorf("core: could not create project dir: %w", err)
	}

	svc.currentWorkspace.config.Projects[name] = &project{
		Jumplist: []string{},
	}

	return svc.currentWorkspace.Save(svc.rootFs)
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
		header: []string{
			"Select the projects the you want to be present in your jumplist",
			"Blank lines or lines starting with '#' will be ignored. The first non-empty line",
			"will be the the 0th item in the jumplist.",
		},
	}

	for name := range svc.currentWorkspace.config.Projects {
		if slices.Contains(svc.currentWorkspace.config.Jumplist, name) {
			jl.items = append(jl.items, name)
		} else {
			jl.trailer = append(jl.trailer, name)
		}
	}

	return &jl, nil
}

func (svc *Service) UpdateProjectJumplist(names []string) error {
	if svc.currentWorkspace == nil {
		return ErrNotInWorkspace
	}

	for _, name := range names {
		if _, ok := svc.currentWorkspace.config.Projects[name]; !ok {
			return fmt.Errorf("core: project does not exist %q", name)
		}
	}

	svc.currentWorkspace.config.Jumplist = names

	return svc.currentWorkspace.Save(svc.rootFs)
}

func (svc *Service) ProjectNames() ([]string, error) {
	if svc.currentWorkspace == nil {
		return nil, ErrNotInWorkspace
	}

	names := make([]string, 0, len(svc.currentWorkspace.config.Projects))

	for name := range svc.currentWorkspace.config.Projects {
		names = append(names, name)
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

	svc.state.Workspaces[name] = &workspace{}

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
		header: []string{
			"Select the workspaces the you want to be present in your jumplist",
			"Blank lines or lines starting with '#' will be ignored. The first non-empty line",
			"will be the the 0th item in the jumplist.",
		},
	}

	for name := range svc.state.Workspaces {
		if slices.Contains(svc.state.Jumplist, name) {
			jl.items = append(jl.items, name)
		} else {
			jl.trailer = append(jl.items, name)
		}
	}

	return &jl
}

func (svc *Service) UpdateWorkspaceJumplist(names []string) error {
	for _, name := range names {
		if _, ok := svc.state.Workspaces[name]; !ok {
			return fmt.Errorf("core: workspace does not exist %q", name)
		}
	}

	svc.state.Jumplist = names

	return svc.state.Save(svc.rootFs)
}

func (svc *Service) WorkspaceNames() []string {
	names := make([]string, 0, len(svc.state.Workspaces))

	for name := range svc.state.Workspaces {
		names = append(names, name)
	}

	return names
}

type Jumplist struct {
	header  []string
	items   []string
	trailer []string
}

func (j *Jumplist) Items() []string { return j.items }

func (j *Jumplist) Add(item string) {
	j.items = append(j.items, item)
}

func (j *Jumplist) Get(index int) (string, error) {
	if index < len(j.items) {
		return j.items[index], nil
	}

	return "", fmt.Errorf("jumplist: item does not exist %d", index)
}

func (j *Jumplist) Serialize(w io.Writer) error {
	for _, line := range j.header {
		if _, err := fmt.Fprintf(w, "# %s\n", line); err != nil {
			return fmt.Errorf("jumplist: failed to serialize: %w", err)
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("jumplist: failed to serialize: %w", err)
	}

	for _, line := range j.items {
		if _, err := fmt.Fprintf(w, "%s\n", line); err != nil {
			return fmt.Errorf("jumplist: failed to serialize: %w", err)
		}
	}

	for _, line := range j.trailer {
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
