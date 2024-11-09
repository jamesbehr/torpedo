package cmd

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/jamesbehr/torpedo/core"
	"github.com/jamesbehr/torpedo/picker"
)

var defaultTemplate embed.FS

type Context struct {
	Service          *core.Service
	Stdout           io.Writer
	Shell            string
	WorkingDirectory string
	Home             string
	ConfigRoot       string
}

func (ctx *Context) ExpandPath(p string) string {
	if p == "" {
		return ""
	}

	if filepath.IsAbs(ctx.Home) {
		rel, err := filepath.Rel("~", p)
		if err == nil && filepath.IsLocal(rel) {
			return filepath.Join(ctx.Home, rel)
		}
	}

	return p
}

func (ctx *Context) UnexpandPath(p string) string {
	if p == "" {
		return ""
	}

	rel, err := filepath.Rel(ctx.Home, p)
	if err == nil && filepath.IsLocal(rel) {
		return filepath.Join("~", rel)
	}

	return p
}

func (ctx *Context) ConfigFilePath(name string) string {
	return filepath.Join(ctx.ConfigRoot, "torpedo", name)
}

type InitCmd struct {
	Template string `default:"default"`
}

func (cmd *InitCmd) Run(ctx *Context) error {
	if _, err := ctx.Service.FindCurrentProject(ctx.WorkingDirectory); err != nil {
		if err != core.ErrProjectNotFound {
			return err
		}
	} else {
		return errors.New("projects should not be nested")
	}

	searchPaths := []string{
		"/etc/torpedo/templates",
		ctx.ConfigFilePath("templates"),
	}

	if v, ok := os.LookupEnv("TORPEDO_TEMPLATE_PATH"); ok {
		searchPaths = filepath.SplitList(v)
	}

	dir, err := ctx.Service.FindTemplateDir(cmd.Template, searchPaths)
	if err != nil {
		return err
	}

	return ctx.Service.InitializeProject(ctx.WorkingDirectory, dir)
}

type PickCmd struct {
	Paths []string
}

func (cmd *PickCmd) Run(ctx *Context) error {
	projects, err := ctx.Service.FindProjects(cmd.Paths)
	if err != nil {
		return err
	}

	for i := range projects {
		projects[i] = ctx.UnexpandPath(projects[i])
	}

	choice, err := picker.Pick(projects)
	if err != nil {
		return err
	}

	projectPath := ctx.ExpandPath(choice)

	cfg, err := ctx.Service.ParseProjectConfig(projectPath)
	if err != nil {
		return err
	}

	sessionName := ctx.UnexpandPath(projectPath)
	return ctx.Service.AttachProject(sessionName, projectPath, cfg.Windows)
}

type RunCmd struct {
	Directory string   `default:"."`
	Command   string   `arg:""`
	Args      []string `arg:"" optional:""`
}

func (cmd *RunCmd) Run(ctx *Context) error {
	projectDir, err := ctx.Service.FindCurrentProject(cmd.Directory)
	if err != nil {
		return err
	}

	config, err := ctx.Service.ParseProjectConfig(projectDir)
	if err != nil {
		return err
	}

	script, ok := config.Commands[cmd.Command]
	if !ok {
		return fmt.Errorf("unknown command %q", cmd.Command)
	}

	return ctx.Service.RunProjectScript(projectDir, "sh", script, cmd.Args)
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
	ctx := kong.Parse(&cli)

	home := os.Getenv("HOME")
	if !filepath.IsAbs(home) {
		ctx.Fatalf("$HOME is not set")
	}

	configHome := os.Getenv("XDG_CONFIG_HOME")
	if !filepath.IsAbs(configHome) {
		configHome = filepath.Join(home, ".config")
	}

	wd, err := os.Getwd()
	if err != nil {
		ctx.Fatalf(err.Error())
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}

	context := Context{
		Service:          core.New(),
		Stdout:           os.Stdout,
		WorkingDirectory: wd,
		Home:             home,
		ConfigRoot:       configHome,
		Shell:            shell,
	}
	ctx.Bind(&context)
	ctx.FatalIfErrorf(ctx.Run())
}
