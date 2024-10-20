package scan

import (
	"errors"
	"path/filepath"

	"github.com/jamesbehr/torpedo/filesystem"
	"github.com/jamesbehr/torpedo/root"
	"github.com/jamesbehr/torpedo/workspace"
)

func TraverseParents(dir string) ([]filesystem.FS, error) {
	result := []filesystem.FS{}

	for {
		result = append(result, filesystem.DirFS(dir))

		next := filepath.Dir(dir)
		if next == dir {
			break
		}

		dir = next
	}

	return result, nil
}

type Info struct {
	Root      *root.Service
	Workspace *workspace.Service
}

func Discover(entries []filesystem.FS) (*Info, error) {
	var rootFs, workspaceFs filesystem.FS

	for _, fs := range entries {
		children, err := fs.ReadDir(".")
		if err != nil {
			return nil, err
		}

		for _, child := range children {
			if child.Name() == root.ConfigName && child.Type().IsRegular() {
				rootFs = fs
			}
		}

		for _, child := range children {
			if child.Name() == workspace.ConfigDir && child.IsDir() {
				workspaceFs = fs
			}
		}
	}

	result := Info{}

	if rootFs != nil {
		rt, err := root.New(rootFs)
		if err != nil {
			return nil, err
		}

		result.Root = rt
	}

	if workspaceFs != nil {
		ws, err := workspace.New(workspaceFs)
		if err != nil {
			return nil, err
		}

		result.Workspace = ws
	}

	return &result, nil
}

func (i *Info) ValidateIsRoot() error {
	if i.Root == nil {
		return errors.New("scan: cannot find root")
	}

	return nil
}
