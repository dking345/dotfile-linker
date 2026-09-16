package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const ignoreFileName = ".dotlinkignore"

// builtinIgnore lists top-level entries that are about the repo itself,
// not something that should end up symlinked into $HOME.
var builtinIgnore = map[string]bool{
	".git":       true,
	".gitignore": true,
	"README.md":  true,
	"LICENSE":    true,
}

// loadIgnore returns the set of top-level entry names to skip for repoDir:
// the built-in defaults plus one name per line from repoDir/.dotlinkignore,
// if that file exists. Blank lines and lines starting with # are ignored.
// A missing ignore file is not an error; every repo works without one.
func loadIgnore(repoDir string) (map[string]bool, error) {
	ignore := make(map[string]bool, len(builtinIgnore))
	for name := range builtinIgnore {
		ignore[name] = true
	}

	f, err := os.Open(filepath.Join(repoDir, ignoreFileName))
	if os.IsNotExist(err) {
		return ignore, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", ignoreFileName, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ignore[line] = true
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", ignoreFileName, err)
	}

	return ignore, nil
}
