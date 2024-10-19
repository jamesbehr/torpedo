package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/jamesbehr/torpedo/config"
	"github.com/jamesbehr/torpedo/editor"
	"github.com/jamesbehr/torpedo/picker"
	"github.com/jamesbehr/torpedo/tmux"
	"github.com/jamesbehr/torpedo/workspace"
)

type WorkspaceNewCmd struct {
	Directory string `type:"exisitingdir" help:"Directory to create the workspace in. Defaults to pwd" default:"${pwd}"`
	Name      string `arg:""`
}

func (cmd *WorkspaceNewCmd) Run(cfg *config.File) error {
	path := filepath.Join(cmd.Directory, cmd.Name)

	if err := workspace.Create(path); err != nil {
		return err
	}

	cfg.Workspaces = append(cfg.Workspaces, &config.Workspace{
		Path: path,
	})

	return cfg.Save()
}

type WorkspaceListCmd struct{}

func (cmd *WorkspaceListCmd) Run(cfg *config.File) error {
	index := 0
	for _, ws := range cfg.Workspaces {
		if ws.Jumplisted {
			fmt.Printf("%s (%d)\n", ws.Path, index)
			index++
		} else {
			fmt.Println(ws.Path)
		}
	}

	return nil
}

type WorkspaceJumplistEditCmd struct{}

func (cmd *WorkspaceJumplistEditCmd) Run(cfg *config.File) error {
	if len(cfg.Workspaces) == 0 {
		return errors.New("no workspaces to edit")
	}

	lines := []string{
		"# Select the workspaces that you want to be present in your jumplist. Blank",
		"# lines or lines starting with '#' will be ignored. The first non-empty line",
		"# will be the 0th item in the jumplist.",
		"",
	}

	for _, ws := range cfg.Workspaces {
		if ws.Jumplisted {
			lines = append(lines, ws.Path)
		}
	}

	for _, ws := range cfg.Workspaces {
		if !ws.Jumplisted {
			lines = append(lines, "# "+ws.Path)
		}

		ws.Jumplisted = false
	}

	path := kong.ExpandPath("~/.config/torpedo/JUMPLIST")
	lines, err := editor.EditList(path, lines)
	if err != nil {
		return err
	}

	workspaces := []*config.Workspace{}

	for _, line := range lines {
		line := strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		found := false
		for _, ws := range cfg.Workspaces {
			if ws.Path == line {
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
	for _, ws := range cfg.Workspaces {
		if !ws.Jumplisted {
			workspaces = append(workspaces, ws)
		}
	}

	cfg.Workspaces = workspaces

	return cfg.Save()
}

type WorkspaceJumplistGetCmd struct {
	Index int `arg:""`
}

func (cmd *WorkspaceJumplistGetCmd) Run(cfg *config.File) error {
	index := 0
	for _, ws := range cfg.Workspaces {
		if ws.Jumplisted {
			if index == cmd.Index {
				fmt.Println(ws.Path)
				return nil
			}

			index++
		}
	}

	return fmt.Errorf("workspace: could not find jumplist item %d", cmd.Index)
}

type WorkspaceJumplistGotoCmd struct {
	Index int `arg:""`
}

func (cmd *WorkspaceJumplistGotoCmd) Run(cfg *config.File, t *tmux.Client) error {
	index := 0
	for _, ws := range cfg.Workspaces {
		if ws.Jumplisted {
			if index == cmd.Index {
				return t.AttachOrCreateSession(&tmux.Session{
					Name: ws.Path,
					Path: ws.Path,
					Environment: []string{
						"TORPEDO_WORKSPACE=" + ws.Path,
					},
				})
			}

			index++
		}
	}

	return fmt.Errorf("workspace: could not find jumplist item %d", cmd.Index)
}

type WorkspaceJumplistCmd struct {
	Goto WorkspaceJumplistGotoCmd `cmd:""`
	Edit WorkspaceJumplistEditCmd `cmd:""`
	Get  WorkspaceJumplistGetCmd  `cmd:""`
}

type WorkspacePickCmd struct{}

func (cmd *WorkspacePickCmd) Run(cfg *config.File, t *tmux.Client) error {
	items := []string{}

	for _, ws := range cfg.Workspaces {
		items = append(items, ws.Path)
	}

	choice, err := picker.Pick(items)
	if err != nil {
		return err
	}

	return t.AttachOrCreateSession(&tmux.Session{
		Name: choice,
		Path: choice,
		Environment: []string{
			"TORPEDO_WORKSPACE=" + choice,
		},
	})
}

type WorkspaceCmd struct {
	New      WorkspaceNewCmd      `cmd:"" help:"Create a new workspace"`
	List     WorkspaceListCmd     `cmd:"" help:"List all known workspaces"`
	Pick     WorkspacePickCmd     `cmd:"" help:"Pick a workspace in the picker"`
	Jumplist WorkspaceJumplistCmd `cmd:"" help:"Manage the workspace jumplist"`
}
