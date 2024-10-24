package cmd

import (
	"fmt"

	"github.com/jamesbehr/torpedo/core"
)

type FileJumplistEditCmd struct{}

func (cmd *FileJumplistEditCmd) Run(svc *core.Service) error {
	jl, err := svc.FileJumplist()
	if err != nil {
		return err
	}

	loaded, err := editJumplist(jl)
	if err != nil {
		return err
	}

	return svc.UpdateFileJumplist(loaded)
}

type FileJumplistGetCmd struct {
	Index int `arg:""`
}

func (cmd *FileJumplistGetCmd) Run(svc *core.Service) error {
	jl, err := svc.FileJumplist()
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

type FileJumplistAddCmd struct {
	Name string `arg:""`
}

func (cmd *FileJumplistAddCmd) Run(svc *core.Service) error {
	jl, err := svc.FileJumplist()
	if err != nil {
		return err
	}

	jl.Add(cmd.Name)

	return svc.UpdateFileJumplist(jl.Items())
}

type FileJumplistCmd struct {
	Add  FileJumplistAddCmd  `cmd:""`
	Edit FileJumplistEditCmd `cmd:""`
	Get  FileJumplistGetCmd  `cmd:""`
}

type FileCmd struct {
	Jumplist FileJumplistCmd `cmd:"" help:"Manage the file jumplist"`
}
