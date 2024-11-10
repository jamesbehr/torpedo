package cmd

import (
	"fmt"

	"github.com/jamesbehr/torpedo/format"
	"github.com/jamesbehr/torpedo/marks"
)

type MarksDelCmd struct {
	Key string `arg:""`
}

func (cmd *MarksDelCmd) Run(ctx *Context) error {
	path := ctx.ConfigFilePath("marks.json")

	m, err := marks.ReadMarks(path)
	if err != nil {
		return err
	}

	delete(m, cmd.Key)

	return m.Write(path)
}

type MarksSetCmd struct {
	Key string `arg:""`
}

func (cmd *MarksSetCmd) Run(ctx *Context) error {
	dir, err := ctx.Service.FindCurrentProject(ctx.WorkingDirectory)
	if err != nil {
		return err
	}

	path := ctx.ConfigFilePath("marks.json")

	m, err := marks.ReadMarks(path)
	if err != nil {
		return err
	}

	m[cmd.Key] = ctx.UnexpandPath(dir)

	return m.Write(path)
}

type MarksListCmd struct {
	Fields []string `default:"mark,project" enum:"mark,path,project"`
	Format string   `default:"text"`
}

func (cmd *MarksListCmd) Run(ctx *Context) error {
	path := ctx.ConfigFilePath("marks.json")

	m, err := marks.ReadMarks(path)
	if err != nil {
		return err
	}

	formatter, err := format.New(cmd.Format, cmd.Fields, ctx.Stdout)
	if err != nil {
		return err
	}

	for key, value := range m {
		data := map[string]any{
			"mark":    key,
			"project": value,
			"path":    ctx.ExpandPath(value),
		}

		if err := formatter.Write(data); err != nil {
			return err
		}
	}

	return formatter.Close()
}

type MarksJumpCmd struct {
	Key string `arg:""`
}

func (cmd *MarksJumpCmd) Run(ctx *Context) error {
	path := ctx.ConfigFilePath("marks.json")

	m, err := marks.ReadMarks(path)
	if err != nil {
		return err
	}

	value, ok := m[cmd.Key]
	if !ok {
		return fmt.Errorf("no such mark %q", cmd.Key)
	}

	projectDir := ctx.ExpandPath(value)

	cfg, err := ctx.Service.ParseProjectConfig(projectDir)
	if err != nil {
		return err
	}

	sessionName := ctx.UnexpandPath(projectDir)

	exists, err := ctx.Service.HasSession(sessionName)
	if err != nil {
		return err
	}

	if !exists {
		if err := ctx.Service.CreateSession(sessionName, projectDir, cfg.Windows); err != nil {
			return err
		}
	}

	return ctx.Service.AttachSession(sessionName)
}

type MarksCmd struct {
	Set  MarksSetCmd  `cmd:"" help:"Set a mark"`
	Del  MarksDelCmd  `cmd:"" help:"Delete a mark"`
	List MarksListCmd `cmd:"" help:"List your marks"`
	Jump MarksJumpCmd `cmd:"" help:"Jump to a marked project by index"`
}
