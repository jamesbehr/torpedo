package proto

import (
	"fmt"
	"os"
	"os/exec"
)

type Client struct {
	// BinaryPath is the path (absolute or relative to $PATH) for the Tmux binary.
	// If it is empty, it defaults to "tmux"
	BinaryPath string

	// Session is the target session to execute the commands in.
	// This overrides the TMUX environment variable when executing the command.
	// If empty, it uses the value of the TMUX variable from the process's environment
	Session string
}

type Command interface {
	Args() []string
}

func (c *Client) cmd(args ...string) *exec.Cmd {
	binaryPath := "tmux"
	if c.BinaryPath != "" {
		binaryPath = c.BinaryPath
	}

	cmd := exec.Command(binaryPath, args...)
	cmd.Stdin = os.Stdin // allows tmux to detect a terminal

	if c.Session != "" {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("TMUX=%s", c.Session))
	}

	return cmd
}

func (c *Client) Run(command Command) error {
	cmd := c.cmd(command.Args()...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tmux: error running tmux: %w", err)
	}

	return nil
}

func IsExitError(err error) bool {
	if _, ok := err.(*exec.ExitError); ok {
		return true
	}

	return false
}

type NewSession struct {
	Name           *string
	StartDirectory *string
	Environment    []string
	Detached       bool
}

func (opts *NewSession) Args() []string {
	args := []string{"new-session"}

	if opts.Name != nil {
		args = append(args, "-s", *opts.Name)
	}

	if opts.StartDirectory != nil {
		args = append(args, "-c", *opts.StartDirectory)
	}

	for _, env := range opts.Environment {
		args = append(args, "-e", env)
	}

	if opts.Detached {
		args = append(args, "-d")
	}

	return args
}

type HasSession string

func (opts HasSession) Args() []string {
	return []string{"has-session", "-t", string(opts)}
}

type AttachSession string

func (opts AttachSession) Args() []string {
	return []string{"attach-session", "-t", string(opts)}
}

type SwitchClient struct {
	Session *string
}

func (opts *SwitchClient) Args() []string {
	args := []string{"switch-client"}

	if opts.Session != nil {
		args = append(args, "-t", *opts.Session)
	}

	return args
}
