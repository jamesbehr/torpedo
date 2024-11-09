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

You can now pick a project to jump to, using `fzf`. This also requires `tmux`
to be installed.

    $ torpedo pick

After picking a project, Torpedo will switch to the corresponding `tmux`
session or create it if it does not exist.

You can re-run this command at any time to pick a project. This can be
bound to a key in your `~/.tmux.conf` for easy access.

    bind-key C-f display-popup -E -E "torpedo pick"

You can also provide a path (or paths) to `pick`.

    $ torpedo pick --path ~/personal --path ~/work

The default set of paths can be configured, otherwise it will default to
`$HOME`.

    TODO

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
Commands are programs stored in the `.torpedo/cmd` directory of your project
that can be run by name via the `run` command. Commands that do not have the
executable bit set are ignored.

For example, you can put a script at `.torpedo/cmd/test`, set the executable
bit and run it with the following command.

    $ torpedo run test
