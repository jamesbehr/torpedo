package cmd

import (
	"embed"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/jamesbehr/torpedo/core"
	"github.com/jamesbehr/torpedo/picker"
	"github.com/jamesbehr/torpedo/util"
)

var defaultTemplate embed.FS

type Context struct {
	Service *core.Service
	Stdout  io.Writer
}

type InitCmd struct {
	Template  string `default:"default"`
	Directory string `default:"."`
}

func (cmd *InitCmd) Run(ctx *Context) error {
	if _, err := ctx.Service.FindCurrentProject(cmd.Directory); err != nil {
		if err != core.ErrProjectNotFound {
			return err
		}
	} else {
		return errors.New("projects should not be nested")
	}

	searchPaths := []string{} // TODO; Defaults
	if v, ok := os.LookupEnv("TORPEDO_TEMPLATE_PATH"); ok {
		searchPaths = filepath.SplitList(v)
	}

	dir, err := ctx.Service.FindTemplateDir(cmd.Template, searchPaths)
	if err != nil {
		return err
	}

	return ctx.Service.InitializeProject(cmd.Directory, dir)
}

type PickCmd struct {
	Paths []string `name:"path" default:"."`
}

func (cmd *PickCmd) Run(ctx *Context) error {
	projects, err := ctx.Service.FindProjects(cmd.Paths)
	if err != nil {
		return err
	}

	for i := range projects {
		projects[i] = util.UnexpandPath(projects[i])
	}

	choice, err := picker.Pick(projects)
	if err != nil {
		return err
	}

	projectPath := util.ExpandPath(choice)

	cfg, err := ctx.Service.ParseProjectConfig(projectPath)
	if err != nil {
		return err
	}

	return ctx.Service.AttachProject(projectPath, cfg.Windows)
}

type RunCmd struct {
	Directory string   `default:"."`
	Command   string   `arg:""`
	Args      []string `arg:""`
}

func (cmd *RunCmd) Run(ctx *Context) error {
	projectDir, err := ctx.Service.FindCurrentProject(cmd.Directory)
	if err != nil {
		return err
	}

	return ctx.Service.RunProjectCommand(projectDir, cmd.Command, cmd.Args)
}

type CLI struct {
	Init      InitCmd      `cmd:"" help:"Initialize a project"`
	Pick      PickCmd      `cmd:"" help:"Find project and jump to it"`
	Marks     MarksCmd     `cmd:"" help:"Manage project marks"`
	FileMarks FileMarksCmd `cmd:"" help:"Manage file marks within a project"`
	Run       RunCmd       `cmd:"" help:"Run a project command"`
}

var cli CLI

func Execute() {
	// TODO: kong can read defaults from a configuration
	ctx := kong.Parse(&cli)
	context := Context{
		Service: &core.Service{}, // TODO: Create Tmux client
		Stdout:  os.Stdout,
	}
	ctx.Bind(&context)
	ctx.FatalIfErrorf(ctx.Run())
}
