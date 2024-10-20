package cmd

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/jamesbehr/torpedo/scan"
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

	entries, err := scan.TraverseParents(pwd)
	if err != nil {
		ctx.FatalIfErrorf(err)
	}

	info, err := scan.Discover(entries)
	if err != nil {
		ctx.FatalIfErrorf(err)
	}

	ctx.Bind(info)
	ctx.Bind(&tmux.Client{})

	ctx.FatalIfErrorf(ctx.Run())
}
