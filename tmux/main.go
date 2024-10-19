package tmux

import (
	"errors"
	"os"
	"os/exec"

	"github.com/jamesbehr/torpedo/tmux/proto"
)

type Client struct {
	proto.Client
}

type Session struct {
	Name        string
	Path        string
	Environment []string
}

func (c *Client) AttachOrCreateSession(params *Session) error {
	if err := c.Run(proto.HasSession(params.Name)); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			cmd := &proto.NewSession{
				Detached:       true,
				Name:           &params.Name,
				StartDirectory: &params.Path,
				Environment:    params.Environment,
			}

			if err := c.Run(cmd); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if _, ok := os.LookupEnv("TMUX"); ok {
		return c.Run(&proto.SwitchClient{
			Session: &params.Name,
		})
	} else {
		return c.Run(proto.AttachSession(params.Name))
	}
}
