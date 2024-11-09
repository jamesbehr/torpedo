package cmd

import (
	"errors"
	"path/filepath"

	"github.com/jamesbehr/torpedo/format"
	"github.com/jamesbehr/torpedo/marks"
)

type FileMarksDelCmd struct {
	Key string `arg:""`
}

func (cmd *FileMarksDelCmd) Run(ctx *Context) error {
	dir, err := ctx.Service.FindCurrentProject(ctx.WorkingDirectory)
	if err != nil {
		return err
	}

	marksPath := ctx.Service.ProjectDataFilePath(dir, "marks.json")
	m, err := marks.ReadFileMarks(marksPath)
	if err != nil {
		return err
	}

	delete(m, cmd.Key)

	return m.Write(marksPath)
}

type FileMarksSetCmd struct {
	Key    string `arg:""`
	File   string `arg:""`
	Line   uint64 `arg:""`
	Column uint64 `arg:""`
}

func (cmd *FileMarksSetCmd) Run(ctx *Context) error {
	dir, err := ctx.Service.FindCurrentProject(ctx.WorkingDirectory)
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

	marksPath := ctx.Service.ProjectDataFilePath(dir, "marks.json")
	m, err := marks.ReadFileMarks(marksPath)
	if err != nil {
		return err
	}

	m[cmd.Key] = marks.File(cmd.File, cmd.Line, cmd.Column)

	return m.Write(marksPath)
}

type FileMarksListCmd struct {
	Fields []string `default:"mark,line,col,file" enum:"mark,line,col,file,path"`
	Format string   `default:"text"`
}

func (cmd *FileMarksListCmd) Run(ctx *Context) error {
	dir, err := ctx.Service.FindCurrentProject(ctx.WorkingDirectory)
	if err != nil {
		return err
	}

	marksPath := ctx.Service.ProjectDataFilePath(dir, "marks.json")
	m, err := marks.ReadFileMarks(marksPath)
	if err != nil {
		return err
	}

	formatter, err := format.New(cmd.Format, cmd.Fields, ctx.Stdout)
	if err != nil {
		return err
	}

	for key, value := range m {
		data := map[string]any{
			"mark": key,
			"col":  value.Column,
			"line": value.Line,
			"file": value.Path,
			"path": filepath.Join(dir, value.Path),
		}

		if err := formatter.Write(data); err != nil {
			return err
		}
	}

	return formatter.Close()
}

type FileMarksCmd struct {
	Set  FileMarksSetCmd  `cmd:"" help:"Set a file mark"`
	Del  FileMarksDelCmd  `cmd:"" help:"Delete a file mark"`
	List FileMarksListCmd `cmd:"" help:"List your file marks"`
}
