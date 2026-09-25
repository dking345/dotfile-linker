package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func runLink(repoDir, targetDir, only string, dryRun, force bool) error {
	repoDir, err := filepath.Abs(repoDir)
	if err != nil {
		return fmt.Errorf("resolving repo dir: %w", err)
	}

	ignore, err := loadIgnore(repoDir)
	if err != nil {
		return err
	}

	if only != "" {
		if ignore[only] {
			return fmt.Errorf("%s is ignored (see %s)", only, ignoreFileName)
		}
		src := filepath.Join(repoDir, only)
		if _, err := os.Lstat(src); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("%s: no such entry in %s", only, repoDir)
			}
			return err
		}
		return linkOne(src, filepath.Join(targetDir, only), dryRun, force)
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

		if err := linkOne(src, dst, dryRun, force); err != nil {
			fmt.Fprintf(os.Stderr, "dotlink: %s: %v\n", name, err)
		}
	}
	return nil
}

func linkOne(src, dst string, dryRun, force bool) error {
	info, err := os.Lstat(dst)
	if os.IsNotExist(err) {
		return createLink(src, dst, dryRun)
	}
	if err != nil {
		return err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		current, err := os.Readlink(dst)
		if err != nil {
			return err
		}
		if current == src {
			fmt.Printf("ok      %s\n", dst)
			return nil
		}
		if !force {
			return fmt.Errorf("%s already links to %s (use -force to replace)", dst, current)
		}
		if dryRun {
			fmt.Printf("relink  %s -> %s (was %s)\n", dst, src, current)
			return nil
		}
		if err := os.Remove(dst); err != nil {
			return err
		}
		return createLink(src, dst, dryRun)
	}

	if info.IsDir() {
		// directories are swapped aside whole, no content comparison needed
		return backupAndLink(src, dst, dryRun, true)
	}

	identical, err := filesEqual(src, dst)
	if err != nil {
		return err
	}
	if identical {
		if dryRun {
			fmt.Printf("replace %s -> %s (identical content, no backup needed)\n", dst, src)
			return nil
		}
		if err := os.Remove(dst); err != nil {
			return err
		}
		return createLink(src, dst, dryRun)
	}

	return backupAndLink(src, dst, dryRun, false)
}

func backupAndLink(src, dst string, dryRun, isDir bool) error {
	backup := dst + ".dotlink-bak-" + time.Now().Format("20060102-150405")
	if dryRun {
		kind := "file"
		if isDir {
			kind = "directory"
		}
		fmt.Printf("backup  %s -> %s (existing %s differs)\n", dst, backup, kind)
		fmt.Printf("link    %s -> %s\n", dst, src)
		return nil
	}
	if err := os.Rename(dst, backup); err != nil {
		return fmt.Errorf("backing up existing entry: %w", err)
	}
	return createLink(src, dst, dryRun)
}

func createLink(src, dst string, dryRun bool) error {
	if dryRun {
		fmt.Printf("link    %s -> %s\n", dst, src)
		return nil
	}
	if err := os.Symlink(src, dst); err != nil {
		return fmt.Errorf("creating symlink: %w", err)
	}
	fmt.Printf("linked  %s -> %s\n", dst, src)
	return nil
}
