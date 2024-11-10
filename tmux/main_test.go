package tmux

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommands(t *testing.T) {
	tests := []struct {
		Command  Command
		Expected []string
	}{
		// new-session
		{
			Command:  &NewSession{},
			Expected: []string{"new-session"},
		},
		{
			Command:  &NewSession{SessionName: "foo"},
			Expected: []string{"new-session", "-s", "foo"},
		},
		{
			Command:  &NewSession{StartDirectory: "foo"},
			Expected: []string{"new-session", "-c", "foo"},
		},
		{
			Command:  &NewSession{WindowName: "foo"},
			Expected: []string{"new-session", "-n", "foo"},
		},
		{
			Command:  &NewSession{Detached: true},
			Expected: []string{"new-session", "-d"},
		},
		{
			Command:  &NewSession{Environment: []string{"FOO=1", "BAR=2"}},
			Expected: []string{"new-session", "-e", "FOO=1", "-e", "BAR=2"},
		},
	}

	for _, test := range tests {
		require.Equal(t, test.Expected, test.Command.Args())
	}
}
