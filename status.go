package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// runStatus reports how each top-level repo entry relates to its target
// without touching the filesystem, so it can be run as often as wanted to
// see what a `link` call would do first.
func runStatus(repoDir, targetDir string) error {
	repoDir, err := filepath.Abs(repoDir)
	if err != nil {
		return fmt.Errorf("resolving repo dir: %w", err)
	}

	entries, err := os.ReadDir(repoDir)
	if err != nil {
		return fmt.Errorf("reading repo dir: %w", err)
	}

	drift := 0
	for _, entry := range entries {
		name := entry.Name()
		if ignoreEntries[name] {
			continue
		}

		src := filepath.Join(repoDir, name)
		dst := filepath.Join(targetDir, name)

		state, detail, err := entryStatus(src, dst)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dotlink: %s: %v\n", name, err)
			continue
		}
		if state != "linked" {
			drift++
		}
		if detail != "" {
			fmt.Printf("%-9s %s (%s)\n", state, dst, detail)
		} else {
			fmt.Printf("%-9s %s\n", state, dst)
		}
	}

	if drift == 0 {
		fmt.Println("no drift")
	}
	return nil
}

// entryStatus classifies a single dst against the src it would be linked
// to, mirroring the cases linkOne handles but reporting instead of acting.
func entryStatus(src, dst string) (state, detail string, err error) {
	info, err := os.Lstat(dst)
	if os.IsNotExist(err) {
		return "missing", "", nil
	}
	if err != nil {
		return "", "", err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		current, err := os.Readlink(dst)
		if err != nil {
			return "", "", err
		}
		if current == src {
			return "linked", "", nil
		}
		return "conflict", "linked elsewhere -> " + current, nil
	}

	if info.IsDir() {
		return "conflict", "directory in the way", nil
	}

	identical, err := filesEqual(src, dst)
	if err != nil {
		return "", "", err
	}
	if identical {
		return "identical", "same content, not yet linked", nil
	}
	return "conflict", "file differs from repo", nil
}
