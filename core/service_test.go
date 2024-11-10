package core

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jamesbehr/torpedo/tmux"
	"github.com/stretchr/testify/require"
)

func lookPath(t *testing.T, cmd string) string {
	path, err := exec.LookPath(cmd)
	if err != nil {
		t.Fatal(t)
	}

	return path
}

func TestCreateSession(t *testing.T) {
	tmp := t.TempDir()

	tests := []struct {
		Name       string
		Config     string
		ProjectDir string
		Given      []Window
		Expected   []Window
	}{
		{
			Name:       "basic",
			Config:     "testdata/tmux/config/base.conf",
			ProjectDir: tmp,
			Given: []Window{
				{
					Name:   "foo",
					Layout: "tiled",
					Panes: []Pane{
						{
							Pwd:    "1",
							Cmd:    []string{"man", "tmux"},
							Active: false,
						},
						{
							Pwd:    "2",
							Active: false,
						},
					},
					Active: false,
				},
			},
			Expected: []Window{
				{
					Name:   "foo",
					Layout: "9295,80x24,0,0[80x11,0,0,0,80x12,0,12,1]",
					Panes: []Pane{
						{
							Pwd: filepath.Join(tmp, "1"),
							Cmd: []string{
								lookPath(t, "man"),
								"tmux",
							},
							Env: []string{},
						},
						{
							Pwd:    filepath.Join(tmp, "2"),
							Cmd:    []string{"sh"},
							Env:    []string{},
							Active: true,
						},
					},
					Active: true,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := tmux.Client{
				SocketPath: filepath.Join(t.TempDir(), "tmux"),
				Config:     test.Config,
			}

			defer func() {
				if err := client.Run(&tmux.KillServer{}); err != nil {
					t.Fatal(err)
				}
			}()

			if err := client.Run(&tmux.StartServer{}); err != nil {
				t.Fatal(err)
			}

			svc := Service{&client}

			if err := svc.CreateSession(t.Name(), test.ProjectDir, test.Given); err != nil {
				t.Fatal(err)
			}

			dump, err := svc.DumpSession(t.Name())
			if err != nil {
				t.Fatal(err)
			}

			require.Equal(t, dump, test.Expected)
		})
	}
}
