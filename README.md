# Torpedo
Inspired by ThePrimeagen's Harpoon and tmux-sessionizer

## Workspaces
The tool is built around the idea *workspaces* that contain *projects*.

To begin creating workspaces and project, you must first create a workspace
root. A workspace root contains a `.torpedo.kdl` configuration file that tracks
all the workspaces.

    $ mkdir ~/workspaces
    $ cd ~/workspaces
    $ torpedo init

Once inside a workspace root, you can create a workspaces, e.g. called `foo`.

    $ torpedo workspace new foo

Once you have created a workspace, you can select it with the fuzzy finder.

    $ torpedo workspace pick

After picking a workspace, `torpedo` will switch to the corresponding `tmux`
session or create it if it does not exist.

You can re-run this command at any time to pick a workspace. This can be
bound to a key in your `~/.tmux.conf` for easy access.

    bind-key C-w display-popup -E -E "torpedo workspace pick"

You can also manage a workspace jumplist. This is a numbered list of workspaces
that allow for quick jumping. You can edit the workspace jumplist inside your
editor.

    $ torpedo workspace jumplist edit

This will print out all the available workspaces, but commented out. You can
uncomment and reorder them to edit the jumplist. The first uncommented line is
numbered 0, with the remaining items being numbered sequentially.
You can jump to workspace by its jumplist number.

    $ torpedo workspace jumplist goto 0

You can also see the workspace path without switching.

    $ torpedo workspace jumplist get 0

## Projects
Once you're inside a workspace, you are able to start adding projects to it.
You can create a new project directory `foo`.

    $ torpedo project new foo

Sometimes the directory already exists, for example if you `git clone` it. You
can add an existing directory as a project instead.

    $ git clone ssh://git@github.com/jamesbehr/torpedo
    $ torpedo project add torpedo

Torpedo remembers all your projects inside your workspace's config file.
Projects have their own Tmux sessions that you can jump to with the fuzzy finder.

    $ torpedo project pick

If you know the project's name you can jump to it directly. This can be useful
for scripting, such as when you have a project that is common to many
workspaces (e.g. `notes`).

    $ torpedo project goto notes

Projects have their own jumplists as well that you can edit in the same way as
workspaces.

    $ torpedo project jumplist edit
    $ torpedo project jumplist goto 0

You can automate this all inside your `~/.tmux.conf`.

    bind-key C-f display-popup -E -E "torpedo project pick"
    bind-key C-n run-shell "torpedo project goto notes"
    bind-key C-h run-shell "torpedo project jumplist goto 0"
    bind-key C-j run-shell "torpedo project jumplist goto 1"
    bind-key C-k run-shell "torpedo project jumplist goto 2"
    bind-key C-l run-shell "torpedo project jumplist goto 3"

## Files
Like Harpoon, we have file jumplists for each project as well. Torpedo doesn't
manage a Tmux session for a file in the jumplist, but you can edit it and get
the filepath just like you would for a workspace or project. You must be inside
a project for this jumplist to work.

    $ torpedo files jumplist edit
    $ torpedo files jumplist get 0

You could use this in a script, e.g. to open the first jumplist item in `vim`.

    $ vim "$(torpedo files jumplist get 0)"

This could also be integrated inside your editor for fast Harpoon-style jumping
as well.
