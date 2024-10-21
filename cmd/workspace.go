package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/jamesbehr/torpedo/core"
	"github.com/jamesbehr/torpedo/editor"
	"github.com/jamesbehr/torpedo/picker"
	"github.com/jamesbehr/torpedo/tmux"
)

type WorkspaceNewCmd struct {
	Name string `arg:""`
}

func (cmd *WorkspaceNewCmd) Run(svc *core.Service) error {
	return svc.CreateWorkspace(cmd.Name)
}

type WorkspaceJumplistEditCmd struct{}

func (cmd *WorkspaceJumplistEditCmd) Run(svc *core.Service) error {
	f, err := os.CreateTemp("", "torpedo.jumplist*")
	if err != nil {
		return err
	}

	name := f.Name()

	defer os.Remove(name)
	defer f.Close()

	jl := svc.WorkspaceJumplist()
	if err := jl.Serialize(f); err != nil {
		return err
	}

	if err := editor.Open(name); err != nil {
		return err
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	loaded, err := core.ParseJumplist(f)
	if err != nil {
		return err
	}

	if err := svc.UpdateWorkspaceJumplist(loaded); err != nil {
		return err
	}

	return nil
}

type WorkspaceJumplistGetCmd struct {
	Index int `arg:""`
}

func (cmd *WorkspaceJumplistGetCmd) Run(svc *core.Service) error {
	jl := svc.WorkspaceJumplist()

	name, err := jl.Get(cmd.Index)
	if err != nil {
		return err
	}

	fmt.Println(name)

	return nil
}

type WorkspaceJumplistGotoCmd struct {
	Index int `arg:""`
}

func (cmd *WorkspaceJumplistGotoCmd) Run(svc *core.Service, t *tmux.Client) error {
	jl := svc.WorkspaceJumplist()

	name, err := jl.Get(cmd.Index)
	if err != nil {
		return err
	}

	return svc.AttachWorkspaceSession(name, t)
}

type WorkspaceJumplistCmd struct {
	Goto WorkspaceJumplistGotoCmd `cmd:""`
	Edit WorkspaceJumplistEditCmd `cmd:""`
	Get  WorkspaceJumplistGetCmd  `cmd:""`
}

type WorkspacePickCmd struct{}

func (cmd *WorkspacePickCmd) Run(svc *core.Service, t *tmux.Client) error {
	name, err := picker.Pick(svc.WorkspaceNames())
	if err != nil {
		return err
	}

	return svc.AttachWorkspaceSession(name, t)
}

type WorkspaceCmd struct {
	New      WorkspaceNewCmd      `cmd:"" help:"Create a new workspace"`
	Pick     WorkspacePickCmd     `cmd:"" help:"Pick a workspace in the picker"`
	Jumplist WorkspaceJumplistCmd `cmd:"" help:"Manage the workspace jumplist"`
}
