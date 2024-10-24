package cmd

import (
	"fmt"

	"github.com/jamesbehr/torpedo/core"
	"github.com/jamesbehr/torpedo/picker"
	"github.com/jamesbehr/torpedo/tmux"
)

type ProjectNewCmd struct {
	Name string `arg:""`
}

func (cmd *ProjectNewCmd) Run(svc *core.Service) error {
	return svc.CreateProject(cmd.Name)
}

type ProjectJumplistEditCmd struct{}

func (cmd *ProjectJumplistEditCmd) Run(svc *core.Service) error {
	jl, err := svc.ProjectJumplist()
	if err != nil {
		return err
	}

	loaded, err := editJumplist(jl)
	if err != nil {
		return err
	}

	return svc.UpdateProjectJumplist(loaded)
}

type ProjectJumplistGetCmd struct {
	Index int `arg:""`
}

func (cmd *ProjectJumplistGetCmd) Run(svc *core.Service) error {
	jl, err := svc.ProjectJumplist()
	if err != nil {
		return err
	}

	name, err := jl.Get(cmd.Index)
	if err != nil {
		return err
	}

	fmt.Println(name)

	return nil
}

type ProjectJumplistGotoCmd struct {
	Index int `arg:""`
}

func (cmd *ProjectJumplistGotoCmd) Run(svc *core.Service, t *tmux.Client) error {
	jl, err := svc.ProjectJumplist()
	if err != nil {
		return err
	}

	name, err := jl.Get(cmd.Index)
	if err != nil {
		return err
	}

	return svc.AttachProjectSession(name, t)
}

type ProjectJumplistCmd struct {
	Goto ProjectJumplistGotoCmd `cmd:""`
	Edit ProjectJumplistEditCmd `cmd:""`
	Get  ProjectJumplistGetCmd  `cmd:""`
}

type ProjectPickCmd struct{}

func (cmd *ProjectPickCmd) Run(svc *core.Service, t *tmux.Client) error {
	names, err := svc.ProjectNames()
	if err != nil {
		return err
	}

	name, err := picker.Pick(names)
	if err != nil {
		return err
	}

	return svc.AttachWorkspaceSession(name, t)
}

type ProjectCmd struct {
	New      ProjectNewCmd      `cmd:"" help:"Create a new project"`
	Pick     ProjectPickCmd     `cmd:"" help:"Pick a project in the picker"`
	Jumplist ProjectJumplistCmd `cmd:"" help:"Manage the project jumplist"`
	File     FileCmd            `cmd:"" help:"Manage project files"`
}
