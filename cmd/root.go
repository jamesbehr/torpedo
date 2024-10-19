package cmd

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/jamesbehr/torpedo/config"
	"github.com/jamesbehr/torpedo/tmux"
)

type ProjectCmd struct {
}

type CLI struct {
	Config    string       `help:"Path to config file" default:"${config}" type:"path"`
	Workspace WorkspaceCmd `cmd:"" help:"Manage workspaces"`
	Project   ProjectCmd   `cmd:"" help:"Manage projects inside a workspace"`
}

func (cli *CLI) BeforeApply(cfg *config.File) error {
	c, err := config.Load(cli.Config)
	if err != nil {
		return err
	}

	*cfg = *c

	return nil
}

var cli CLI

func Execute() {
	pwd, err := os.Getwd()
	if err != nil {
		panic("Failed to get current working directory")
	}

	ctx := kong.Parse(&cli, kong.Vars{
		"pwd":    pwd,
		"config": "~/.config/torpedo/config",
	}, kong.Bind(&config.File{}), kong.Bind(&tmux.Client{}))

	ctx.FatalIfErrorf(ctx.Run())
}
