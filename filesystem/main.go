package filesystem

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type File interface {
	fs.File
	io.Writer
}

type FS interface {
	Path(string) (string, error)
	Sub(string) (FS, error)
	ReadFile(string) ([]byte, error)
	WriteFile(string, []byte, fs.FileMode) error
	MkdirAll(string, fs.FileMode) error
	Create(string) (File, error)
	Open(string) (File, error)
}

type DirFS string

func (fsys DirFS) Path(name string) (string, error) {
	if !fs.ValidPath(name) {
		return "", &os.PathError{Op: "path", Path: name, Err: os.ErrInvalid}
	}

	return filepath.Join(string(fsys), name), nil
}

func (fsys DirFS) Sub(name string) (FS, error) {
	if !fs.ValidPath(name) {
		return nil, &os.PathError{Op: "sub", Path: name, Err: os.ErrInvalid}
	}

	return DirFS(filepath.Join(string(fsys), name)), nil
}

func (fsys DirFS) MkdirAll(name string, perm fs.FileMode) error {
	if !fs.ValidPath(name) {
		return &os.PathError{Op: "mkdir", Path: name, Err: os.ErrInvalid}
	}

	if err := os.MkdirAll(filepath.Join(string(fsys), name), perm); err != nil {
		if e, ok := err.(*os.PathError); ok {
			e.Path = name
		}

		return err
	}

	return nil
}

func (fsys DirFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	if !fs.ValidPath(name) {
		return &os.PathError{Op: "writefile", Path: name, Err: os.ErrInvalid}
	}

	if err := os.WriteFile(filepath.Join(string(fsys), name), data, perm); err != nil {
		if e, ok := err.(*os.PathError); ok {
			e.Path = name
		}

		return err
	}

	return nil
}

func (fsys DirFS) ReadFile(name string) ([]byte, error) {
	if !fs.ValidPath(name) {
		return nil, &os.PathError{Op: "readfile", Path: name, Err: os.ErrInvalid}
	}

	b, err := os.ReadFile(filepath.Join(string(fsys), name))
	if err != nil {
		if e, ok := err.(*os.PathError); ok {
			e.Path = name
		}

		return nil, err
	}

	return b, nil
}

func (fsys DirFS) Create(name string) (File, error) {
	if !fs.ValidPath(name) {
		return nil, &os.PathError{Op: "create", Path: name, Err: os.ErrInvalid}
	}

	f, err := os.Create(filepath.Join(string(fsys), name))
	if err != nil {
		if e, ok := err.(*os.PathError); ok {
			e.Path = name
		}

		return nil, err
	}

	return f, nil
}

func (fsys DirFS) Open(name string) (File, error) {
	if !fs.ValidPath(name) {
		return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrInvalid}
	}

	f, err := os.Open(filepath.Join(string(fsys), name))
	if err != nil {
		if e, ok := err.(*os.PathError); ok {
			e.Path = name
		}

		return nil, err
	}

	return f, nil
}
