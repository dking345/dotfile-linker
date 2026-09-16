package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// runUnlink reverses what link did: for each top-level repo entry that is
// currently a symlink pointing into repoDir, it removes the symlink and, if
// a backup exists from an earlier link run, restores the newest one in its
// place. Entries that aren't symlinks owned by this repo are left alone.
func runUnlink(repoDir, targetDir string, dryRun bool) error {
	repoDir, err := filepath.Abs(repoDir)
	if err != nil {
		return fmt.Errorf("resolving repo dir: %w", err)
	}

	ignore, err := loadIgnore(repoDir)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(repoDir)
	if err != nil {
		return fmt.Errorf("reading repo dir: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if ignore[name] {
			continue
		}

		src := filepath.Join(repoDir, name)
		dst := filepath.Join(targetDir, name)

		if err := unlinkOne(src, dst, dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "dotlink: %s: %v\n", name, err)
		}
	}
	return nil
}

func unlinkOne(src, dst string, dryRun bool) error {
	info, err := os.Lstat(dst)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return nil
	}

	current, err := os.Readlink(dst)
	if err != nil {
		return err
	}
	if current != src {
		return nil
	}

	backup, err := newestBackup(dst)
	if err != nil {
		return err
	}

	if dryRun {
		if backup != "" {
			fmt.Printf("restore %s -> %s (from %s)\n", dst, src, backup)
		} else {
			fmt.Printf("remove  %s -> %s\n", dst, src)
		}
		return nil
	}

	if err := os.Remove(dst); err != nil {
		return err
	}
	if backup == "" {
		fmt.Printf("removed %s -> %s\n", dst, src)
		return nil
	}
	if err := os.Rename(backup, dst); err != nil {
		return fmt.Errorf("restoring backup: %w", err)
	}
	fmt.Printf("restored %s (from %s)\n", dst, backup)
	return nil
}

// newestBackup returns the most recently created backup for dst, or "" if
// none exists. Backup names embed a sortable timestamp, so the newest one is
// the last after a lexicographic sort.
func newestBackup(dst string) (string, error) {
	matches, err := filepath.Glob(dst + ".dotlink-bak-*")
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", nil
	}
	sort.Strings(matches)
	return matches[len(matches)-1], nil
}
