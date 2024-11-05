package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/jamesbehr/torpedo/core"
)

type FileMarksRemoveCmd struct {
	Directory string `default:"."`
	File      string `arg:""`
	Line      uint64 `arg:""`
}

func (cmd *FileMarksRemoveCmd) Run(ctx *Context) error {
	dir, err := ctx.Service.FindCurrentProject(cmd.Directory)
	if err != nil {
		return err
	}

	path, err := filepath.Abs(cmd.File)
	if err != nil {
		return err
	}

	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return err
	}

	if !filepath.IsLocal(rel) {
		return errors.New("file must be within the current project")
	}

	marks, err := ctx.Service.FileMarks(dir)
	if err != nil {
		return err
	}

	mark := core.NewFileMark(rel, cmd.Line)
	return marks.Remove(mark.String())
}

type FileMarksEditCmd struct {
	Directory string `default:"."`
	Editor    string `env:"VISUAL,EDITOR" default:"vi"`
}

func (cmd *FileMarksEditCmd) Run(ctx *Context) error {
	dir, err := ctx.Service.FindCurrentProject(cmd.Directory)
	if err != nil {
		return err
	}

	marks, err := ctx.Service.FileMarks(dir)
	if err != nil {
		return err
	}

	return marks.Edit(cmd.Editor)
}

type FileMarksAddCmd struct {
	Directory string `default:"."`
	File      string `arg:""`
	Line      uint64 `arg:""`
}

func (cmd *FileMarksAddCmd) Run(ctx *Context) error {
	dir, err := ctx.Service.FindCurrentProject(cmd.Directory)
	if err != nil {
		return err
	}

	path, err := filepath.Abs(cmd.File)
	if err != nil {
		return err
	}

	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return err
	}

	if !filepath.IsLocal(rel) {
		return errors.New("file must be within the current project")
	}

	marks, err := ctx.Service.FileMarks(dir)
	if err != nil {
		return err
	}

	mark := core.NewFileMark(rel, cmd.Line)
	return marks.Add(mark.String())
}

type FileMarksListCmd struct {
	Directory string `default:"."`
	Format    string `default:"{{.Index}}: {{.Mark}}"`
}

type fileMark struct {
	Index    int
	Mark     string
	Filename string
	Line     uint64
	Path     string
}

func (cmd *FileMarksListCmd) Run(ctx *Context) error {
	tmpl, err := template.New("format").Parse(cmd.Format)
	if err != nil {
		return err
	}

	dir, err := ctx.Service.FindCurrentProject(cmd.Directory)
	if err != nil {
		return err
	}

	marks, err := ctx.Service.FileMarks(dir)
	if err != nil {
		return err
	}

	items, err := marks.List()
	if err != nil {
		return err
	}

	for i, mark := range items {
		fm, err := core.ParseFileMark(mark)
		if err != nil {
			continue
		}

		data := fileMark{
			Index:    i,
			Mark:     mark,
			Filename: fm.Path,
			Line:     fm.Line,
			Path:     filepath.Join(dir, fm.Path),
		}

		data.Path = filepath.Join(dir, data.Filename)

		if err := tmpl.Execute(ctx.Stdout, &data); err != nil {
			return err
		}

		if _, err := fmt.Fprintln(ctx.Stdout); err != nil {
			return err
		}
	}

	return nil
}

type FileMarksGetCmd struct {
	Directory string `default:"."`
	Index     int    `arg:""`
}

func (cmd *FileMarksGetCmd) Run(ctx *Context) error {
	dir, err := ctx.Service.FindCurrentProject(cmd.Directory)
	if err != nil {
		return err
	}

	marks, err := ctx.Service.FileMarks(dir)
	if err != nil {
		return err
	}

	mark, err := marks.Get(cmd.Index)
	if err != nil {
		return err
	}

	path := filepath.Join(dir, mark)

	if _, err := fmt.Fprintln(ctx.Stdout, path); err != nil {
		return err
	}

	return nil
}

type FileMarksJumpCmd struct {
	Directory string   `default:"."`
	Index     int      `arg:""`
	Editor    string   `env:"VISUAL,EDITOR" default:"vi"`
	Keys      []string `default:"Escape,:e +{{.Line}} {{.Path}},Enter"`
}

func (cmd *FileMarksJumpCmd) Run(ctx *Context) error {
	dir, err := ctx.Service.FindCurrentProject(cmd.Directory)
	if err != nil {
		return err
	}

	marks, err := ctx.Service.FileMarks(dir)
	if err != nil {
		return err
	}

	mark, err := marks.Get(cmd.Index)
	if err != nil {
		return err
	}

	fm, err := core.ParseFileMark(mark)
	if err != nil {
		return err
	}

	keys := make([]string, len(cmd.Keys))

	for i := range cmd.Keys {
		tmpl, err := template.New("format").Parse(cmd.Keys[i])
		if err != nil {
			return err
		}

		var result bytes.Buffer

		if err := tmpl.Execute(&result, fm); err != nil {
			return err
		}

		keys[i] = result.String()
	}

	return ctx.Service.SendKeysToProcess(cmd.Editor, keys)
}

type FileMarksCmd struct {
	Add  FileMarksAddCmd    `cmd:"" help:"Add a file to your marks"`
	Rm   FileMarksRemoveCmd `cmd:"" help:"Remove a file from your marks"`
	Edit FileMarksEditCmd   `cmd:"" help:"Edit your file marks in your editor"`
	List FileMarksListCmd   `cmd:"" help:"List your file marks"`
	Get  FileMarksGetCmd    `cmd:"" help:"Get a marked file by index"`
	Jump FileMarksJumpCmd   `cmd:"" help:"Jump to a marked file by index"`
}
