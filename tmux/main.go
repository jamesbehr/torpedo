package tmux

import (
	"fmt"
	"os"
	"os/exec"
)

func InSession() bool {
	_, ok := os.LookupEnv("TMUX")
	return ok
}

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

func (c *Client) Output(command Command) (string, error) {
	cmd := c.cmd(command.Args()...)

	result, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tmux: error running tmux: %w", err)
	}

	return string(result), nil
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
	WindowName     string
	SessionName    string
	StartDirectory string
	Environment    []string
	Detached       bool
	Command        []string
}

func (opts *NewSession) Args() []string {
	args := []string{"new-session"}
	if opts.SessionName != "" {
		args = append(args, "-s", opts.SessionName)
	}

	if opts.WindowName != "" {
		args = append(args, "-n", opts.WindowName)
	}

	if opts.StartDirectory != "" {
		args = append(args, "-c", opts.StartDirectory)
	}

	for _, env := range opts.Environment {
		args = append(args, "-e", env)
	}

	if opts.Detached {
		args = append(args, "-d")
	}

	args = append(args, opts.Command...)

	return args
}

type HasSession string

func (opts HasSession) Args() []string {
	return []string{"has-session", "-t", string(opts)}
}

type AttachSession struct {
	SessionName string
}

func (opts AttachSession) Args() []string {
	args := []string{"attach-session"}

	if opts.SessionName != "" {
		args = append(args, "-t", opts.SessionName)
	}

	return args
}

type SwitchClient struct {
	SessionName string
}

func (opts *SwitchClient) Args() []string {
	args := []string{"switch-client"}

	if opts.SessionName != "" {
		args = append(args, "-t", opts.SessionName)
	}

	return args
}

type NewWindow struct {
	WindowName     string
	StartDirectory string
	Environment    []string
	Command        []string
}

func (opts *NewWindow) Args() []string {
	args := []string{"new-window"}
	if opts.WindowName != "" {
		args = append(args, "-n", opts.WindowName)
	}

	if opts.StartDirectory != "" {
		args = append(args, "-c", opts.StartDirectory)
	}

	for _, env := range opts.Environment {
		args = append(args, "-e", env)
	}

	args = append(args, opts.Command...)

	return args
}

type SplitWindow struct {
	StartDirectory string
	Environment    []string
	Command        []string
}

func (opts *SplitWindow) Args() []string {
	args := []string{"split-window"}
	if opts.StartDirectory != "" {
		args = append(args, "-c", opts.StartDirectory)
	}

	for _, env := range opts.Environment {
		args = append(args, "-e", env)
	}

	args = append(args, opts.Command...)

	return args
}

type SelectLayout struct {
	LayoutName string
}

func (opts *SelectLayout) Args() []string {
	args := []string{"select-layout"}
	if opts.LayoutName != "" {
		args = append(args, opts.LayoutName)
	}

	return args
}

type SelectPane struct {
	TargetPane string
}

func (opts *SelectPane) Args() []string {
	args := []string{"select-pane"}
	if opts.TargetPane != "" {
		args = append(args, "-t", opts.TargetPane)
	}

	return args
}

type SelectWindow struct {
	TargetWindow string
}

func (opts *SelectWindow) Args() []string {
	args := []string{"select-window"}
	if opts.TargetWindow != "" {
		args = append(args, "-t", opts.TargetWindow)
	}

	return args
}

type ListWindows struct {
	Format string
}

func (opts *ListWindows) Args() []string {
	args := []string{"list-windows"}
	if opts.Format != "" {
		args = append(args, "-F", opts.Format)
	}

	return args
}

type SendKeys struct {
	TargetPane string
	Keys       []string
}

func (opts *SendKeys) Args() []string {
	args := []string{"send-keys"}

	if opts.TargetPane != "" {
		args = append(args, "-t", opts.TargetPane)
	}

	args = append(args, opts.Keys...)

	return args
}

type ListPanes struct {
	Session bool
	Filter  string
	Format  string
}

func (opts *ListPanes) Args() []string {
	args := []string{"list-panes"}

	if opts.Session {
		args = append(args, "-s")
	}

	if opts.Format != "" {
		args = append(args, "-F", opts.Format)
	}

	if opts.Filter != "" {
		args = append(args, "-f", opts.Filter)
	}

	return args
}

type Multi []Command

func (m Multi) Args() []string {
	args := []string{}

	for i, cmd := range m {
		if i > 0 {
			args = append(args, ";")
		}

		args = append(args, cmd.Args()...)
	}

	return args
}
