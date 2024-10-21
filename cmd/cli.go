package cmd

import (
	"github.com/alecthomas/kong"
	"github.com/jamesbehr/torpedo/core"
	"github.com/jamesbehr/torpedo/tmux"
)

type CLI struct {
	Workspace WorkspaceCmd `cmd:"" help:"Manage workspaces"`
	Project   ProjectCmd   `cmd:"" help:"Manage projects inside a workspace"`
}

var cli CLI

func Execute() {
	ctx := kong.Parse(&cli)

	svc, err := core.Scan()
	if err != nil {
		ctx.FatalIfErrorf(err)
	}

	ctx.Bind(svc)
	ctx.Bind(&tmux.Client{})

	ctx.FatalIfErrorf(ctx.Run())
}
