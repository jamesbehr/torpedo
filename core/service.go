package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/jamesbehr/torpedo/tmux"
)

type Service struct {
	tmux *tmux.Client
}

func New() *Service {
	return &Service{
		tmux: &tmux.Client{},
	}
}

const projectDataDir = ".torpedo"
const maxParents = 64

var ErrProjectNotFound = errors.New("could not find project root")

// FindCurrentProject finds the absolute root path to the project that is a
// parent of currentDirectory, if one exits.
// It will look for a .torpedo directory starting in currentDirectory and
// checking each parent of currentDirectory until there are no more parents or
// the maximum number of traversals is exceeded.
// If a .torpedo directory is found, it return the absolute path to the
// immediate parent of that directory.
// If no such directory is found, it will return an [ErrProjectNotFound] error
func (svc *Service) FindCurrentProject(currentDirectory string) (string, error) {
	for i := 0; i < maxParents; i++ {
		dataDir := filepath.Join(currentDirectory, projectDataDir)
		if info, err := os.Stat(dataDir); err == nil {
			if info.IsDir() {
				return filepath.Abs(currentDirectory)
			}
		} else if !os.IsNotExist(err) {
			return "", err
		}

		parent := filepath.Dir(currentDirectory)
		if parent == currentDirectory {
			break
		}

		currentDirectory = parent
	}

	return "", ErrProjectNotFound
}

// FindTemplateDir finds an absolute path to a template given a template name or path.
// If templateName is an absolute path, then it will be returned as is.
// If templateName is a relative path such that [filepath.IsLocal](path) would
// return false, then it is resolved relative to the current directory
// Otherwise, it attempts to resolve the template first against the current directory.
// If that fails it will look relative to each path in searchPaths, starting with the first one.
// If all of these attempts fail to find a valid directory, then an error is returned.
func (svc *Service) FindTemplateDir(templateName string, searchPaths []string) (string, error) {
	if filepath.IsAbs(templateName) {
		return templateName, nil
	}

	templatePath, err := filepath.Abs(templateName)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(templatePath)
	if err == nil {
		if info.IsDir() {
			return templatePath, nil
		}
	} else {
		if !os.IsNotExist(err) {
			return "", err
		}
	}

	if filepath.IsLocal(templateName) {
		for _, dir := range searchPaths {
			templatePath := filepath.Join(dir, templateName)
			info, err := os.Stat(templatePath)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}

				return "", err
			}

			if info.IsDir() {
				return templatePath, nil
			}
		}
	}

	return "", fmt.Errorf("template does not exist %q", searchPaths)
}

// InitializeProject walks all the files in templateDir, copying them into
// targetDir and preserving their permissions.
// If targetDir does not exist, then this returns an error.
func (svc *Service) InitializeProject(targetDir string, templateDir string) error {
	if _, err := os.Stat(targetDir); err != nil {
		return err
	}

	return filepath.WalkDir(targetDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		dstPath := filepath.Join(targetDir, path)

		info, err := entry.Info()
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return os.Mkdir(dstPath, info.Mode())
		}

		srcPath := filepath.Join(templateDir, path)
		data, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}

func (svc *Service) FindProjects(paths []string) ([]string, error) {
	projects := []string{}

	for _, path := range paths {
		err := filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if entry.IsDir() && entry.Name() == projectDataDir {
				project, err := filepath.Abs(filepath.Dir(path))
				if err != nil {
					return err
				}

				projects = append(projects, project)
				return fs.SkipDir
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return projects, nil
}

func (svc *Service) AttachSession(sessionName string) error {
	if tmux.InSession() {
		switchClient := tmux.SwitchClient{
			SessionName: sessionName,
		}

		if err := svc.tmux.Run(&switchClient); err != nil {
			return fmt.Errorf("AttachSession: unable to switch client: %w", err)
		}
	}

	attachSession := tmux.AttachSession{
		SessionName: sessionName,
	}

	if err := svc.tmux.Run(&attachSession); err != nil {
		return fmt.Errorf("AttachSession: unable to attach to session: %w", err)
	}

	return nil
}

func fields(s string) [][]string {
	fields := [][]string{}
	row := []string{}
	field := ""
	escaping := false

	for _, r := range s {
		if escaping {
			// TODO: Handle escape chars properly
			field += string(r)
			escaping = false
		} else {
			switch r {
			case '\n':
				row = append(row, field)
				fields = append(fields, row)
				field = ""
				row = []string{}
			case '\r':
			case '\\':
				escaping = true
			case ' ', '\t':
				row = append(row, field)
				field = ""
			default:
				field += string(r)
			}
		}
	}

	return fields
}

func parse(row []string, args ...any) error {
	if len(args) != len(row) {
		return fmt.Errorf("expected %d fields but got %d", len(args), len(row))
	}

	for i, arg := range args {
		s := row[i]

		switch v := arg.(type) {
		case *bool:
			b, err := strconv.ParseBool(s)
			if err != nil {
				return err
			}

			*v = b
		case *int:
			n, err := strconv.Atoi(s)
			if err != nil {
				return err
			}

			*v = n
		case *string:
			*v = s
		default:
			return fmt.Errorf("bad argument type %T", arg)
		}
	}

	return nil
}

func parseNullTerminatedList(data []byte) []string {
	result := []string{}

	for i := 0; i < 10; i++ {
		idx := bytes.IndexByte(data, 0)
		if idx < 0 {
			break
		}

		result = append(result, string(data[:idx]))

		if idx+1 >= len(data) {
			break
		}

		data = data[idx+1:]
	}

	return result
}

func (svc *Service) DumpSession(sessionName string) ([]Window, error) {
	dumpSession := tmux.Multi{
		&tmux.ListPanes{
			Session: true,
			Target:  sessionName,
			Format:  "pane #{window_id} #{pane_start_path} #{pane_active} #{pane_pid}",
		},
		&tmux.ListWindows{
			TargetSession: sessionName,
			Format:        "window #{window_id} #{window_name} #{window_active} #{window_layout}",
		},
	}

	output, err := svc.tmux.Output(dumpSession)
	if err != nil {
		return nil, fmt.Errorf("DumpSession: %w", err)
	}

	panes := map[string][]Pane{}
	windows := []Window{}

	env := os.Environ()

	for _, row := range fields(string(output)) {
		if len(row) == 0 {
			continue
		}

		switch row[0] {
		case "pane":
			var pane Pane
			var id, pid string
			if err := parse(row[1:], &id, &pane.Pwd, &pane.Active, &pid); err != nil {
				return nil, fmt.Errorf("DumpSession: %w", err)
			}

			cmdline, err := os.ReadFile(filepath.Join("/proc", pid, "cmdline"))
			if err != nil {
				return nil, fmt.Errorf("DumpSession: unable to read cmdline args: %w", err)
			}

			pane.Cmd = parseNullTerminatedList(cmdline)

			environ, err := os.ReadFile(filepath.Join("/proc", pid, "environ"))
			if err != nil {
				return nil, fmt.Errorf("DumpSession: unable to read environment variables: %w", err)
			}

			pane.Env = []string{}
			for _, v := range parseNullTerminatedList(environ) {
				if strings.HasPrefix(v, "TMUX=") || slices.Contains(env, v) {
					continue
				}

				pane.Env = append(pane.Env, v)
			}

			panes[id] = append(panes[id], pane)
		case "window":
			var window Window
			var id string
			if err := parse(row[1:], &id, &window.Name, &window.Active, &window.Layout); err != nil {
				return nil, fmt.Errorf("DumpSession: %w", err)
			}

			window.Panes = panes[id]
			windows = append(windows, window)
		default:
			continue
		}
	}

	return windows, nil
}

func (svc *Service) HasSession(sessionName string) (bool, error) {
	hasSession := tmux.HasSession{
		SessionName: sessionName,
	}

	ok, err := svc.tmux.Success(&hasSession)
	if err != nil {
		return false, fmt.Errorf("HasSession: %w", err)
	}

	return ok, err
}

func (svc *Service) CreateSession(sessionName, projectPath string, windows []Window) error {
	cmds := []tmux.Command{}

	newSession := tmux.NewSession{
		SessionName:    sessionName,
		StartDirectory: projectPath,
		Detached:       true,
	}

	cmds = append(cmds, &newSession)

	activeWindow := -1

	for wi, window := range windows {
		activePane := -1

		if window.Active {
			activeWindow = wi
		}

		if wi == 0 {
			newSession.WindowName = window.Name

			for pi, pane := range window.Panes {
				if pi == 0 {
					newSession.StartDirectory = pane.StartDirectory(projectPath)
					newSession.Environment = pane.Env
					newSession.Command = pane.Cmd
				} else {
					cmds = append(cmds, &tmux.SplitWindow{
						StartDirectory: pane.StartDirectory(projectPath),
						Environment:    pane.Env,
						Command:        pane.Cmd,
					})
				}

				if pane.Active {
					activePane = pi
				}
			}
		} else {
			newWindow := tmux.NewWindow{
				WindowName:   window.Name,
				TargetWindow: fmt.Sprintf("%s:", sessionName),
			}

			cmds = append(cmds, &newWindow)

			for pi, pane := range window.Panes {
				if pi == 0 {
					newWindow.StartDirectory = pane.StartDirectory(projectPath)
					newWindow.Environment = pane.Env
					newWindow.Command = pane.Cmd
				} else {
					cmds = append(cmds, &tmux.SplitWindow{
						StartDirectory: pane.StartDirectory(projectPath),
						Environment:    pane.Env,
						Command:        pane.Cmd,
					})
				}

				if pane.Active {
					activePane = pi
				}
			}
		}

		if window.Layout != "" {
			cmds = append(cmds, &tmux.SelectLayout{
				LayoutName: window.Layout,
			})
		}

		if activePane >= 0 {
			cmds = append(cmds, &tmux.SelectPane{
				TargetPane: strconv.Itoa(activePane),
			})
		}
	}

	if activeWindow >= 0 {
		cmds = append(cmds, &tmux.SelectWindow{
			TargetWindow: strconv.Itoa(activeWindow),
		})
	}

	if err := svc.tmux.Run(tmux.Multi(cmds)); err != nil {
		return fmt.Errorf("CreateSession: unable to create session: %w", err)
	}

	return nil
}

func (svc *Service) ParseProjectConfig(projectPath string) (*Config, error) {
	configPath := filepath.Join(projectPath, projectDataDir, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (svc *Service) RunProjectScript(projectPath string, shell, script string, args []string) error {
	// Let the shell do the parameter escaping
	for i := range args {
		script += fmt.Sprintf(" \"$%d\"", i+1)
	}

	shellArgs := []string{"-c", script, "torpedo"}
	shellArgs = append(shellArgs, args...)

	cmd := exec.Command(shell, shellArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func (svc *Service) ProjectDataFilePath(projectPath string, filename string) string {
	return filepath.Join(projectPath, projectDataDir, filename)
}

type Config struct {
	Commands map[string]string `json:"commands,omitempty"`
	Windows  []Window          `json:"windows,omitempty"`
}

type Pane struct {
	Pwd    string   `json:"pwd,omitempty"`
	Env    []string `json:"env,omitempty"`
	Cmd    []string `json:"cmd,omitempty"`
	Active bool     `json:"active,omitempty"`
}

func (p *Pane) StartDirectory(projectDir string) string {
	if filepath.IsLocal(p.Pwd) {
		return filepath.Join(projectDir, p.Pwd)
	}

	return projectDir
}

type Window struct {
	Name   string `json:"name,omitempty"`
	Layout string `json:"layout,omitempty"`
	Panes  []Pane `json:"panes,omitempty"`
	Active bool   `json:"active,omitempty"`
}

func cleanMarkLine(line string) string {
	commentStart := strings.IndexRune(line, '#')
	if commentStart >= 0 {
		line = line[:commentStart]
	}

	return strings.TrimSpace(line)
}
