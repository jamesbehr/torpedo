# Torpedo
Inspired by ThePrimeagen's Harpoon and tmux-sessionizer

## Introduction
Torpedo manages via *projects*, which are special directories containing a
`.torpedo` subdirectory and some other files (e.g. source code).
Each project is assigned a Tmux session and Torpedo provides functions to
easily switch between them.

If you are inside a directory, you can turn it into a project by running the
following command.

    $ torpedo init

This creates a `.torpedo` directory, containing a number of configuration
files.

You can also create template `.torpedo` directories.

    $ torpedo init --template go

It's recommended to add the `.torpedo` directory to your global `gitignore`
which is by default located at `~/.config/git/ignore`.
This prevents it being tracked by Git, and gives you the freedom to add
whatever you want to the directory.

You can use Torpedo to search for projects a list of search paths, and then
pick one to jump to using a fuzzy finder.

    $ torpedo pick --paths ~/personal,~/work
    $ torpedo pick --paths ~/personal --paths ~/work # you can pass the flag more than once

This requires `fzf` and `tmux` to be installed.
After picking a project, Torpedo will switch to the corresponding `tmux`
session or create it if it does not exist.

You can re-run this command at any time to pick a project. This can be
bound to a key in your `~/.tmux.conf` for easy access.

    bind-key C-f display-popup -E -E "torpedo pick"

Sometimes, you want to be able to jump to a project faster than is possible
using `fzf`. In this case, you can use marks.
Marks associate a project with a key, and allow you to jump to the project by
key.
If you are inside a project, you can associate that project with the mark `foo`.

    $ torpedo marks set foo

It is then possible to jump to the mark by key, instead of using the fuzzy finder.

    $ torpedo marks jump foo

You can also list the current marks

    $ torpedo marks list

If you don't want the mark anymore, you can delete it

    $ torpedo marks del foo

It can be quite useful to bind these commands to hotkeys in your `.tmux.conf`.

    bind-key C-h run-shell "torpedo marks jump 0"
    bind-key C-j run-shell "torpedo marks jump 1"
    bind-key C-k run-shell "torpedo marks jump 2"
    bind-key C-l run-shell "torpedo marks jump 3"

    bind-key C-s switch-client -T setmark
    bind-key -T setmark C-h run-shell "torpedo marks set 0"
    bind-key -T setmark C-j run-shell "torpedo marks set 1"
    bind-key -T setmark C-k run-shell "torpedo marks set 2"
    bind-key -T setmark C-l run-shell "torpedo marks set 3"

This allows you to, for example, press `<prefix> C-s C-h` to set a mark under
key `0`, then jump to the same mark using `<prefix> C-h`.
It also binds `j`, `k`, `l` to to the same under keys `1`, `2`, and `3`
respectively for Harpoon-style project navigation.

## File marks
Torpedo tracks file marks in a similar way to project marks.
Each project has its own set of file marks.

If you are inside a project, you can add a file mark under key `foo` to the
file `README.md` on line 12, column 1.

    $ torpedo file-marks set foo README.md 12 1

The file must be a file inside the project directories or its subdirectories.

Just like project marks, you can list the available marks

    $ torpedo file-marks list

If you don't want the file mark anymore, you can remove it.

    $ torpedo file-marks rm foo

Unlike project marks, file marks don't provide a handy way to jump to them.
This is intended to be used from inside your text editor via a plugin.
It is possible to list the marks as JSON objects with the mark key, absolute
path, line and column with the following command.

    $ torpedo file-marks list --format json --fields mark,path,line,col

A plugin can use this output to jump to a mark by key.

## Commands
Commands are programs that are configured in your `.torpedo/config.json`,
similar to `npm`.

For example, you can have the following config

    {
        "commands": {
            "test": "go test"
        }
    }

Then you can run a command by name.

    $ torpedo run test

The commands are run using your `$SHELL` via the -c flag.
You can also pass additional arguments to the script.
These will be appended to the command.

    $ torpedo run test ./...

You can pass flags through without parsing them as well after `--`.

    $ torpedo run test -- --help

## Layouts
Torpedo can restore Tmux layouts using a setup in `.torpedo/config.json`.

For example given the following config

    {
        "windows": [
            {
                "name": "editor",
                "active": true,
                "panes": [
                    {
                        "cmd": ["nvim"]
                    }
                ]
            },
            {
                "name": "shell"
            },
            {
                "name": "manual",
                "layout": "tiled",
                "panes": [
                    {
                        "cmd": ["man", "tmux"]
                    },
                    {
                        "cmd": ["man", "nvim"]
                    }
                ]
            }
        ]
    }

When a new session is started for the project, it will create 3 windows.
- The first one, called `editor`, is the active window and runs `nvim`
- The second one, is an empty window called `shell`
- The third one has two panes in the `tiled` layout.
    - One runs `man tmux`
    - One runs `man nvim`

Windows is an array of window objects which have the following fields.
All the fields are optional

| Field    | Description                            |
|----------|----------------------------------------|
| `name`   | The name of the window.                |
| `layout` | A tmux layout string.                  |
| `panes`  | An array of panes in the window.       |
| `active` | Whether the window is selected or not. |

The Tmux layout string can be one of the predefined layout strings (see `man tmux`):
`even-horizontal`, `even-vertical`, `main-horizontal`,
`main-horizontal-mirrored`, `main-vertical`, `main-vertical`, or `tiled`.
Or it can be a layout string as printed by `list-windows`

    $ tmux list-windows
    1: nvim- (1 panes) [210x56] [layout c4dd,210x56,0,0,0] @0
    2: zsh* (1 panes) [210x56] [layout c4df,210x56,0,0,2] @2 (active)

The layout string `c4dd,210x56,0,0,0` defines a single pane 210 colums by 56
rows.

Panes are an array of pane objects with the following fields.
All the fields are optional

| Field    | Description                                       |
|----------|---------------------------------------------------|
| `pwd`    | The working directory relative to the project.    |
| `env`    | A list of environment variables.                  |
| `cmd`    | The command to execute, as an array of arguments. |
| `active` | Whether the pane is selected or not.              |

`env` is a list of environment variables as `KEY=VALUE` pairs.

If `active` is set multiple times in a given array, the last window or pane in
that array with `active` set becomes the active one.
If no window or pane is marked active, the last one in the array becomes
active.

If `pwd` is not a subdirectory of the project, it is ignored.
