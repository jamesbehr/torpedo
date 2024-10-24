package cmd

import (
	"github.com/jamesbehr/torpedo/core"
	"github.com/jamesbehr/torpedo/editor"
)

func editJumplist(jl *core.Jumplist) ([]string, error) {
	f, err := editor.CreateTemp("torpedo.jumplist*")
	if err != nil {
		return nil, err
	}

	defer f.Close()

	if err := jl.Serialize(f); err != nil {
		return nil, err
	}

	if err := f.Edit(); err != nil {
		return nil, err
	}

	loaded, err := core.ParseJumplist(f)
	if err != nil {
		return nil, err
	}

	if err := f.Close(); err != nil {
		return nil, err
	}

	return loaded, nil
}
