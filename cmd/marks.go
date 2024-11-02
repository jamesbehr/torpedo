package cmd

import (
	"fmt"
	"html/template"

	"github.com/jamesbehr/torpedo/core"
	"github.com/jamesbehr/torpedo/util"
)

type MarksRemoveCmd struct {
	Directory string `default:"."`
}

func (cmd *MarksRemoveCmd) Run(ctx *Context) error {
	dir, err := ctx.Service.FindCurrentProject(cmd.Directory)
	if err != nil {
		return err
	}

	marks, err := core.NewMarks(cli.Marks.MarksFile)
	if err != nil {
		return err
	}

	return marks.Remove(util.UnexpandPath(dir))
}

type MarksEditCmd struct {
	Directory string `default:"."`
}

func (cmd *MarksEditCmd) Run(ctx *Context) error {
	marks, err := core.NewMarks(cli.Marks.MarksFile)
	if err != nil {
		return err
	}

	return marks.Edit()
}

type MarksAddCmd struct {
	Directory string `default:"."`
}

func (cmd *MarksAddCmd) Run(ctx *Context) error {
	dir, err := ctx.Service.FindCurrentProject(cmd.Directory)
	if err != nil {
		return err
	}

	marks, err := core.NewMarks(cli.Marks.MarksFile)
	if err != nil {
		return err
	}

	return marks.Add(util.UnexpandPath(dir))
}

type MarksListCmd struct {
	Format string `default:"{{.Index}}: {{.Mark}}"`
}

type projectMark struct {
	Index int
	Mark  string
	Path  string
}

func (cmd *MarksListCmd) Run(ctx *Context) error {
	tmpl, err := template.New("format").Parse(cmd.Format)
	if err != nil {
		return err
	}

	marks, err := core.NewMarks(cli.Marks.MarksFile)
	if err != nil {
		return err
	}

	items, err := marks.List()
	if err != nil {
		return err
	}

	for i, mark := range items {
		data := projectMark{
			Index: i,
			Mark:  mark,
			Path:  util.ExpandPath(mark),
		}

		if err := tmpl.Execute(ctx.Stdout, &data); err != nil {
			return err
		}

		if _, err := fmt.Fprintln(ctx.Stdout); err != nil {
			return err
		}
	}

	return nil
}

type MarksGetCmd struct {
	Index int `arg:""`
}

func (cmd *MarksGetCmd) Run(ctx *Context) error {
	marks, err := core.NewMarks(cli.Marks.MarksFile)
	if err != nil {
		return err
	}

	mark, err := marks.Get(cmd.Index)
	if err != nil {
		return err
	}

	projectDir := util.ExpandPath(mark)

	if _, err := fmt.Fprintln(ctx.Stdout, projectDir); err != nil {
		return err
	}

	return nil
}

type MarksJumpCmd struct {
	Index int `arg:""`
}

func (cmd *MarksJumpCmd) Run(ctx *Context) error {
	marks, err := core.NewMarks(cli.Marks.MarksFile)
	if err != nil {
		return err
	}

	mark, err := marks.Get(cmd.Index)
	if err != nil {
		return err
	}

	projectDir := util.ExpandPath(mark)

	cfg, err := ctx.Service.ParseProjectConfig(projectDir)
	if err != nil {
		return err
	}

	return ctx.Service.AttachProject(projectDir, cfg.Windows)
}

type MarksCmd struct {
	MarksFile string // TODO: Set default

	Add  MarksAddCmd    `cmd:"" help:"Add the current project to your marks"`
	Rm   MarksRemoveCmd `cmd:"" help:"Remove the current project from your marks"`
	Edit MarksEditCmd   `cmd:"" help:"Edit your project marks in your editor"`
	List MarksListCmd   `cmd:"" help:"List your marks"`
	Get  MarksGetCmd    `cmd:"" help:"Get a marked project by index"`
	Jump MarksJumpCmd   `cmd:"" help:"Jump to a marked project by index"`
}
