package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/jamesbehr/torpedo/editor"
	"github.com/jamesbehr/torpedo/picker"
	"github.com/jamesbehr/torpedo/scan"
	"github.com/jamesbehr/torpedo/tmux"
	"github.com/jamesbehr/torpedo/workspace"
)

type WorkspaceNewCmd struct {
	Name string `arg:""`
}

func (cmd *WorkspaceNewCmd) Run(info *scan.Info) error {
	if err := info.ValidateIsRoot(); err != nil {
		return err
	}

	dir, err := info.Root.WorkspaceDir(cmd.Name)
	if err != nil {
		return err
	}

	if _, err := workspace.Initialize(dir); err != nil {
		return err
	}

	if err := info.Root.AddWorkspace(cmd.Name); err != nil {
		return err
	}

	return info.Root.SaveConfig()
}

type WorkspaceJumplistEditCmd struct{}

func (cmd *WorkspaceJumplistEditCmd) Run(info *scan.Info) error {
	if err := info.ValidateIsRoot(); err != nil {
		return err
	}

	f, err := os.CreateTemp("", "torpedo.jumplist*")
	if err != nil {
		return err
	}

	if err := info.Root.WriteJumplist(f); err != nil {
		return err
	}

	if err := editor.Open(f.Name()); err != nil {
		return err
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if err := info.Root.ReadJumplist(f); err != nil {
		return err
	}

	return info.Root.SaveConfig()
}

type WorkspaceJumplistGetCmd struct {
	Index int `arg:""`
}

func (cmd *WorkspaceJumplistGetCmd) Run(info *scan.Info) error {
	if err := info.ValidateIsRoot(); err != nil {
		return err
	}

	wsc, err := info.Root.JumplistItem(cmd.Index)
	if err != nil {
		return err
	}

	fmt.Println(wsc.Name)

	return nil
}

type WorkspaceJumplistGotoCmd struct {
	Index int `arg:""`
}

func (cmd *WorkspaceJumplistGotoCmd) Run(info *scan.Info, t *tmux.Client) error {
	if err := info.ValidateIsRoot(); err != nil {
		return err
	}

	wsc, err := info.Root.JumplistItem(cmd.Index)
	if err != nil {
		return err
	}

	dir, err := info.Root.WorkspaceDir(wsc.Name)
	if err != nil {
		return err
	}

	ws, err := workspace.New(dir)
	if err != nil {
		return err
	}

	return ws.AttachWorkspaceSession(t)
}

type WorkspaceJumplistCmd struct {
	Goto WorkspaceJumplistGotoCmd `cmd:""`
	Edit WorkspaceJumplistEditCmd `cmd:""`
	Get  WorkspaceJumplistGetCmd  `cmd:""`
}

type WorkspacePickCmd struct{}

func (cmd *WorkspacePickCmd) Run(info *scan.Info, t *tmux.Client) error {
	if err := info.ValidateIsRoot(); err != nil {
		return err
	}

	items := []string{}
	for _, ws := range info.Root.Workspaces() {
		items = append(items, ws.Name)
	}

	choice, err := picker.Pick(items)
	if err != nil {
		return err
	}

	dir, err := info.Root.WorkspaceDir(choice)
	if err != nil {
		return err
	}

	ws, err := workspace.New(dir)
	if err != nil {
		return err
	}

	return ws.AttachWorkspaceSession(t)
}

type WorkspaceCmd struct {
	New      WorkspaceNewCmd      `cmd:"" help:"Create a new workspace"`
	Pick     WorkspacePickCmd     `cmd:"" help:"Pick a workspace in the picker"`
	Jumplist WorkspaceJumplistCmd `cmd:"" help:"Manage the workspace jumplist"`
}
