package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "link":
		linkCmd := flag.NewFlagSet("link", flag.ExitOnError)
		dryRun := linkCmd.Bool("dry-run", false, "show what would happen without touching the filesystem")
		force := linkCmd.Bool("force", false, "replace symlinks that point somewhere else")
		only := linkCmd.String("only", "", "link only this top-level entry instead of the whole repo")
		linkCmd.Parse(os.Args[2:])

		args := linkCmd.Args()
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "usage: dotlink link <repo-dir> [target-dir]")
			os.Exit(1)
		}

		repoDir := args[0]
		targetDir := ""
		if len(args) > 1 {
			targetDir = args[1]
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Fprintln(os.Stderr, "dotlink: could not determine home directory:", err)
				os.Exit(1)
			}
			targetDir = home
		}

		if err := runLink(repoDir, targetDir, *only, *dryRun, *force); err != nil {
			fmt.Fprintln(os.Stderr, "dotlink:", err)
			os.Exit(1)
		}
	case "unlink":
		unlinkCmd := flag.NewFlagSet("unlink", flag.ExitOnError)
		dryRun := unlinkCmd.Bool("dry-run", false, "show what would happen without touching the filesystem")
		unlinkCmd.Parse(os.Args[2:])

		args := unlinkCmd.Args()
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "usage: dotlink unlink <repo-dir> [target-dir]")
			os.Exit(1)
		}

		repoDir := args[0]
		targetDir := ""
		if len(args) > 1 {
			targetDir = args[1]
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Fprintln(os.Stderr, "dotlink: could not determine home directory:", err)
				os.Exit(1)
			}
			targetDir = home
		}

		if err := runUnlink(repoDir, targetDir, *dryRun); err != nil {
			fmt.Fprintln(os.Stderr, "dotlink:", err)
			os.Exit(1)
		}
	case "status":
		statusCmd := flag.NewFlagSet("status", flag.ExitOnError)
		statusCmd.Parse(os.Args[2:])

		args := statusCmd.Args()
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "usage: dotlink status <repo-dir> [target-dir]")
			os.Exit(1)
		}

		repoDir := args[0]
		targetDir := ""
		if len(args) > 1 {
			targetDir = args[1]
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Fprintln(os.Stderr, "dotlink: could not determine home directory:", err)
				os.Exit(1)
			}
			targetDir = home
		}

		if err := runStatus(repoDir, targetDir); err != nil {
			fmt.Fprintln(os.Stderr, "dotlink:", err)
			os.Exit(1)
		}
	case "-h", "--help", "help":
		usage()
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `dotlink - symlink a dotfiles repo into your home directory

usage:
  dotlink link <repo-dir> [target-dir]     symlink top-level entries of repo-dir into target-dir (default: $HOME)
  dotlink unlink <repo-dir> [target-dir]   remove symlinks created by link, restoring the newest backup if one exists
  dotlink status <repo-dir> [target-dir]   report drift between repo-dir and target-dir without changing anything

flags for link:
  -dry-run   print what would happen without changing anything
  -force     replace existing symlinks that point somewhere else
  -only      link only this top-level entry instead of the whole repo

flags for unlink:
  -dry-run   print what would happen without changing anything`)
}
