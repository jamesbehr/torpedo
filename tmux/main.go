package tmux

import (
	"errors"
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
	SocketPath string
	Config     string
}

type Command interface {
	Args() []string
}

func (c *Client) cmd(args ...string) *exec.Cmd {
	binaryPath := "tmux"
	if c.BinaryPath != "" {
		binaryPath = c.BinaryPath
	}

	combinedArgs := []string{}

	if c.SocketPath != "" {
		combinedArgs = append(combinedArgs, "-S", c.SocketPath)
	}

	if c.Config != "" {
		combinedArgs = append(combinedArgs, "-f", c.Config)
	}

	combinedArgs = append(combinedArgs, args...)

	cmd := exec.Command(binaryPath, combinedArgs...)
	cmd.Stdin = os.Stdin // allows tmux to detect a terminal

	return cmd
}

func (c *Client) Success(command Command) (bool, error) {
	cmd := c.cmd(command.Args()...)

	if err := cmd.Run(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return false, nil
		}

		return false, fmt.Errorf("tmux: error running tmux: %w", err)
	}

	return true, nil
}

func (c *Client) Output(command Command) ([]byte, error) {
	cmd := c.cmd(command.Args()...)

	result, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("tmux: error running tmux: %w", err)
	}

	return result, nil
}

func (c *Client) Run(command Command) error {
	cmd := c.cmd(command.Args()...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tmux: error running tmux: %w", err)
	}

	return nil
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

type HasSession struct {
	SessionName string
}

func (opts *HasSession) Args() []string {
	args := []string{"has-session"}

	if opts.SessionName != "" {
		args = append(args, "-t", opts.SessionName)
	}

	return args
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
	TargetWindow   string
	WindowName     string
	StartDirectory string
	Environment    []string
	Command        []string
	Detached       bool
}

func (opts *NewWindow) Args() []string {
	args := []string{"new-window"}
	if opts.WindowName != "" {
		args = append(args, "-n", opts.WindowName)
	}

	if opts.Detached {
		args = append(args, "-d")
	}

	if opts.StartDirectory != "" {
		args = append(args, "-c", opts.StartDirectory)
	}

	if opts.TargetWindow != "" {
		args = append(args, "-t", opts.TargetWindow)
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
	TargetSession string
	Format        string
}

func (opts *ListWindows) Args() []string {
	args := []string{"list-windows"}
	if opts.Format != "" {
		args = append(args, "-F", opts.Format)
	}

	if opts.TargetSession != "" {
		args = append(args, "-t", opts.TargetSession)
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
	Target  string
	Filter  string
	Format  string
}

func (opts *ListPanes) Args() []string {
	args := []string{"list-panes"}

	if opts.Session {
		args = append(args, "-s")
	}

	if opts.Target != "" {
		args = append(args, "-t", opts.Target)
	}

	if opts.Format != "" {
		args = append(args, "-F", opts.Format)
	}

	if opts.Filter != "" {
		args = append(args, "-f", opts.Filter)
	}

	return args
}

type ShowOptions struct {
	OnlyValue bool
	Global    bool
	Name      string
}

func (opts *ShowOptions) Args() []string {
	args := []string{"show-options"}

	if opts.OnlyValue {
		args = append(args, "-v")
	}

	if opts.Global {
		args = append(args, "-g")
	}

	if opts.Name != "" {
		args = append(args, opts.Name)
	}

	return args
}

type StartServer struct{}

func (opts *StartServer) Args() []string {
	return []string{"start-server"}
}

type KillServer struct{}

func (opts *KillServer) Args() []string {
	return []string{"kill-server"}
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
