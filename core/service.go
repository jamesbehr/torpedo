package core

import (
	"bufio"
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
	"github.com/jamesbehr/torpedo/util"
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

func (svc *Service) AttachProject(projectPath string, windows []Window) error {
	sessionName := util.UnexpandPath(projectPath)

	hasSession := tmux.HasSession(sessionName)
	if err := svc.tmux.Run(hasSession); err != nil {
		if !tmux.IsExitError(err) {
			return err
		}
	} else {
		return nil
	}

	cmds := []tmux.Command{}

	newSession := tmux.NewSession{
		SessionName: sessionName,
		Detached:    true,
	}

	cmds = append(cmds, &newSession)

	activeWindow := -1

	for wi, window := range windows {
		activePane := -1

		if window.Active {
			activeWindow = wi
		}

		if wi == 0 {
			for pi, pane := range window.Panes {
				if pi == 0 {
					newSession.StartDirectory = projectPath
					newSession.Environment = pane.Env
					newSession.Command = pane.Cmd
				} else {
					cmds = append(cmds, &tmux.SplitWindow{
						StartDirectory: projectPath,
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
				WindowName: window.Name,
			}

			cmds = append(cmds, &newWindow)

			for pi, pane := range window.Panes {
				if pi == 0 {
					newWindow.StartDirectory = projectPath
					newWindow.Environment = pane.Env
					newWindow.Command = pane.Cmd
				} else {
					cmds = append(cmds, &tmux.SplitWindow{
						StartDirectory: projectPath,
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
		return err
	}

	if tmux.InSession() {
		switchClient := tmux.SwitchClient{
			SessionName: sessionName,
		}

		return svc.tmux.Run(&switchClient)
	}

	attachSession := tmux.AttachSession{
		SessionName: sessionName,
	}

	return svc.tmux.Run(&attachSession)
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

func (svc *Service) RunProjectCommand(projectPath string, cmdName string, args []string) error {
	path := filepath.Join(projectPath, projectDataDir, "cmd", cmdName)

	cmd := exec.Command(path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// TODO: Connect signals

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (svc *Service) SendKeysToProcess(processName string, keys []string) error {
	listPanes := tmux.ListPanes{
		Session: true,
		Filter:  fmt.Sprintf("#{==:#{pane_current_command},%s}", processName),
		Format:  "#{pane_id}",
	}

	output, err := svc.tmux.Output(&listPanes)
	if err != nil {
		return err
	}

	panes := strings.Split(output, "\n")
	if len(panes) == 0 {
		return fmt.Errorf("could not find process %q in current Tmux session", processName)
	}

	paneId := panes[0]

	sendKeys := tmux.SendKeys{
		TargetPane: paneId,
		Keys:       keys,
	}

	return svc.tmux.Run(&sendKeys)
}

func (svc *Service) FileMarks(projectPath string) (*Marks, error) {
	return NewMarks(filepath.Join(projectPath, projectDataDir, "marks"))
}

type Config struct {
	Windows []Window `json:"windows"`
}

type Pane struct {
	Pwd    string   `json:"pwd,omitempty"`
	Env    []string `json:"env,omitempty"`
	Cmd    []string `json:"cmd,omitempty"`
	Active bool     `json:"active,omitempty"`
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

type Marks struct {
	path string
}

func NewMarks(path string) (*Marks, error) {
	if path == "" {
		return nil, errors.New("invalid marks file")
	}

	return &Marks{path}, nil
}

func (m *Marks) read() ([]string, error) {
	f, err := os.Open(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	defer f.Close()

	lines := []string{}

	s := bufio.NewScanner(f)

	for s.Scan() {
		lines = append(lines, s.Text())
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	if err := f.Close(); err != nil {
		return nil, err
	}

	return lines, nil
}

func (m *Marks) write(lines []string) error {
	if err := os.MkdirAll(filepath.Dir(m.path), 0777); err != nil {
		return err
	}

	f, err := os.Create(m.path)
	if err != nil {
		return err
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(f, line); err != nil {
			return errors.Join(err, f.Close())
		}
	}

	return f.Close()
}

func (m *Marks) Add(mark string) error {
	lines, err := m.read()
	if err != nil {
		return err
	}

	for _, line := range lines {
		if cleanMarkLine(line) == mark {
			return nil
		}
	}

	return m.write(append(lines, mark))
}

func (m *Marks) Remove(mark string) error {
	lines, err := m.read()
	if err != nil {
		return err
	}

	for i, line := range lines {
		if cleanMarkLine(line) == mark {
			return m.write(slices.Delete(lines, i, i+1))
		}
	}

	return nil
}

func (m *Marks) List() ([]string, error) {
	lines, err := m.read()
	if err != nil {
		return nil, err
	}

	marks := make([]string, 0, len(lines))
	for _, line := range lines {
		cleaned := cleanMarkLine(line)

		if cleaned != "" {
			marks = append(marks, cleaned)
		}
	}

	return marks, nil
}

func (m *Marks) Get(index int) (string, error) {
	lines, err := m.read()
	if err != nil {
		return "", err
	}

	i := 0
	for _, line := range lines {
		cleaned := cleanMarkLine(line)

		if cleaned != "" {
			if i == index {
				return cleaned, nil
			}

			i++
		}
	}

	return "", fmt.Errorf("invalid mark index: %d", index)
}

func (m *Marks) Edit(editor string) error {
	cmd := exec.Command(editor, m.path)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor: unable to start editor: %w", err)
	}

	return nil
}

type FileMark struct {
	Path string
	Line uint64
}

func NewFileMark(path string, line uint64) FileMark {
	return FileMark{path, line}
}

func (fm FileMark) String() string {
	return fmt.Sprintf("%s:%d", fm.Path, fm.Line)
}

func ParseFileMark(mark string) (*FileMark, error) {
	if mark == "" {
		return nil, errors.New("invalid mark")
	}

	pieces := strings.SplitN(mark, ":", 2)
	if len(pieces) != 2 {
		return nil, errors.New("invalid mark")
	}

	n, err := strconv.ParseUint(pieces[1], 10, 64)
	if err != nil {
		return nil, err
	}

	result := FileMark{
		Path: pieces[0],
		Line: n,
	}

	return &result, nil
}
