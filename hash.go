package main

import (
	"bytes"
	"crypto/sha256"
	"io"
	"os"
)

// filesEqual compares two files by streaming their contents through sha256
// rather than reading either one fully into memory first. Dotfiles are
// usually tiny, but this also gets pointed at things like .bash_history or
// .netrc that can grow large over years, and those shouldn't ever require
// buffering the whole file just to check whether a backup is needed.
func filesEqual(a, b string) (bool, error) {
	fa, err := os.Open(a)
	if err != nil {
		return false, err
	}
	defer fa.Close()

	fb, err := os.Open(b)
	if err != nil {
		return false, err
	}
	defer fb.Close()

	infoA, err := fa.Stat()
	if err != nil {
		return false, err
	}
	infoB, err := fb.Stat()
	if err != nil {
		return false, err
	}
	if infoA.Size() != infoB.Size() {
		return false, nil
	}

	hashA := sha256.New()
	if _, err := io.Copy(hashA, fa); err != nil {
		return false, err
	}

	hashB := sha256.New()
	if _, err := io.Copy(hashB, fb); err != nil {
		return false, err
	}

	return bytes.Equal(hashA.Sum(nil), hashB.Sum(nil)), nil
}
