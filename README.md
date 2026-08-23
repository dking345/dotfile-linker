# dotfile-linker

I keep my dotfiles in a git repo and every time I set up a new machine I end
up either hand-copying files into place or writing a throwaway shell script
that half-works. `dotlink` is that script, done once, properly.

It takes a dotfiles repo and symlinks its top-level entries into your home
directory (or wherever you point it), so `~/.vimrc` becomes a symlink to
`~/dotfiles/.vimrc` instead of a copy. Edit either one and the other sees it
immediately, and `git status` in the repo tells you what actually changed.

## What it does

Given a repo laid out like:

```
dotfiles/
  .vimrc
  .gitconfig
  .config/
    nvim/
      init.vim
```

running

```
dotlink link ./dotfiles
```

creates `~/.vimrc`, `~/.gitconfig`, and `~/.config` as symlinks pointing back
into `./dotfiles`. Note that `.config` is linked whole, as one directory, not
walked file by file — if you need finer-grained control over one directory,
split it out at the top level of your repo instead.

If a target already exists:

- **already the right symlink** — left alone, reported as `ok`.
- **a symlink pointing somewhere else** — left alone unless you pass `-force`.
- **a real file or directory** — backed up by renaming it to
  `<name>.dotlink-bak-<timestamp>` before the symlink is created, unless its
  content is byte-for-byte identical to the source, in which case it's
  just replaced (no point keeping a backup of a file with nothing in it
  worth keeping).

That comparison streams both files through SHA-256 instead of reading them
into memory, so pointing this at a home directory with a multi-gigabyte
`.bash_history` or similar won't spike memory usage.

## Usage

```
dotlink link <repo-dir> [target-dir]

  -dry-run   print what would happen without changing anything
  -force     replace existing symlinks that point somewhere else
```

`target-dir` defaults to `$HOME`. Files named `.git`, `.gitignore`,
`README.md`, and `LICENSE` at the top of the repo are skipped automatically,
since those describe the repo, not your home directory.

Example, checking what a link run would do before committing to it:

```
$ dotlink link ~/dotfiles -dry-run
ok      /home/dking345/.gitconfig
backup  /home/dking345/.vimrc -> /home/dking345/.vimrc.dotlink-bak-20260824-101530 (existing file differs)
link    /home/dking345/.vimrc -> /home/dking345/dotfiles/.vimrc
link    /home/dking345/.config -> /home/dking345/dotfiles/.config
```

## Building

Standard library only, no external dependencies.

```
go build -o dotlink .
```

## Status

Early. `link` is the only command. See the roadmap for what's missing.
