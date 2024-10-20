package cmd

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/jamesbehr/torpedo/filesystem"
	"github.com/jamesbehr/torpedo/root"
	"github.com/jamesbehr/torpedo/tmux"
)

type ProjectCmd struct {
}

type CLI struct {
	Workspace WorkspaceCmd `cmd:"" help:"Manage workspaces"`
	Project   ProjectCmd   `cmd:"" help:"Manage projects inside a workspace"`
}

var cli CLI

func Execute() {
	ctx := kong.Parse(&cli)

	pwd, err := os.Getwd()
	if err != nil {
		ctx.FatalIfErrorf(err)
	}

	rt, err := root.New(filesystem.DirFS(pwd))
	if err != nil {
		ctx.FatalIfErrorf(err)
	}

	ctx.Bind(rt)
	ctx.Bind(&tmux.Client{})

	ctx.FatalIfErrorf(ctx.Run())
}
