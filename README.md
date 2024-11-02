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
using `fzf`. In this case, you can use marks. If you are inside a project, you
can add/remove it to/from your list of marks.

    $ torpedo marks add
    $ torpedo marks rm

Its possible to rearrange the list of marks in your chosen text editor.

    $ torpedo marks edit

Once you have some projects marked, you can then jump to one of the items by
way of its index on the list of marks.

    $ torpedo marks jump 0

This is quite useful when bound to a hotkey in your `~/.tmux.conf`.

    bind-key C-h run-shell "torpedo marks jump 0"
    bind-key C-j run-shell "torpedo marks jump 1"
    bind-key C-k run-shell "torpedo marks jump 2"
    bind-key C-l run-shell "torpedo marks jump 3"

## File marks
TODO

## Commands
Commands are programs stored in the `.torpedo/cmd` directory of your project
that can be run by name via the `run` command. Commands that do not have the
executable bit set are ignored.

For example, you can put a script at `.torpedo/cmd/test`, set the executable
bit and run it with the following command.

    $ torpedo run test
