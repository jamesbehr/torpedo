package cmd

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"path/filepath"
)

type FileMarksRemoveCmd struct {
	Directory string `default:"."`
	File      string `arg:""`
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

	return marks.Remove(rel)
}

type FileMarksEditCmd struct {
	Directory string `default:"."`
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

	return marks.Edit()
}

type FileMarksAddCmd struct {
	Directory string `default:"."`
	File      string `arg:""`
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

	return marks.Add(rel)
}

type FileMarksListCmd struct {
	Directory string `default:"."`
	Format    string `default:"{{.Index}}: {{.Mark}}"`
}
type fileMark struct {
	Index int
	Mark  string
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
		data := fileMark{
			Index: i,
			Mark:  mark,
			// TODO: Path, line numbers
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

	if _, err := fmt.Fprintln(ctx.Stdout, mark); err != nil {
		return err
	}

	return nil
}

type FileMarksJumpCmd struct {
	Directory string `default:"."`
	Index     int    `arg:""`
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

	log.Println(mark)

	// TODO:
	// Find the pane that is running your editor
	//   tmux list-panes -st test -f '#{==:#{pane_current_command},nvim}' -F '#{pane_id}'
	// Then using the returned pane id, you can send commands to Vim
	//   tmux send-keys -t '%2' Escape ':e +<line> <filename>' Enter

	return nil
}

type FileMarksCmd struct {
	Add  FileMarksAddCmd    `cmd:"" help:"Add a file to your marks"`
	Rm   FileMarksRemoveCmd `cmd:"" help:"Remove a file from your marks"`
	Edit FileMarksEditCmd   `cmd:"" help:"Edit your file marks in your editor"`
	List FileMarksListCmd   `cmd:"" help:"List your file marks"`
	Get  FileMarksGetCmd    `cmd:"" help:"Get a marked file by index"`
	Jump FileMarksJumpCmd   `cmd:"" help:"Jump to a marked file by index"`
}
