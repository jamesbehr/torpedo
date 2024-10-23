package editor

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type File struct {
	closed bool
	edited bool
	f      *os.File
}

func CreateTemp(pattern string) (*File, error) {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, fmt.Errorf("editor: cannot create temp file: %w", err)
	}

	return &File{false, false, f}, nil
}

func (f *File) Write(buf []byte) (int, error) {
	if f.edited {
		return 0, errors.New("editor: cannot write file that has been edited")
	}

	return f.f.Write(buf)
}

func (f *File) Read(buf []byte) (int, error) {
	if !f.edited {
		return 0, errors.New("editor: cannot read file until it has been edited")
	}

	return f.f.Read(buf)
}

func (f *File) Edit() error {
	if _, err := f.f.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("editor: cannot seek: %w", err)
	}

	if err := Open(f.f.Name()); err != nil {
		return err
	}

	f.edited = true
	return nil
}

func (f *File) Close() error {
	if f.closed {
		return nil
	}

	if err := f.f.Close(); err != nil {
		return fmt.Errorf("editor: cannot close temporary file: %w", err)
	}

	if err := os.Remove(f.f.Name()); err != nil {
		return fmt.Errorf("editor: cannot remove temporary file: %w", err)
	}

	f.closed = true
	return nil
}

func Editor() string {
	if v, ok := os.LookupEnv("TERM"); ok && v != "dumb" {
		v, ok := os.LookupEnv("VISUAL")
		if ok {
			return v
		}
	}

	v, ok := os.LookupEnv("EDITOR")
	if ok {
		return v
	}

	return "vi"
}

func Open(path string) error {
	cmd := exec.Command(Editor(), path)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor: unable to start editor: %w", err)
	}

	return nil
}
