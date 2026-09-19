package main

import (
	"os"
	"path/filepath"
	"testing"
)

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

func lstatMode(t *testing.T, path string) (os.FileMode, bool) {
	t.Helper()
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return 0, false
	}
	if err != nil {
		t.Fatalf("lstat %s: %v", path, err)
	}
	return info.Mode(), true
}

func TestLinkOne_CreatesSymlinkWhenMissing(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.conf")
	dst := filepath.Join(dir, "dst.conf")
	mustWriteFile(t, src, "content")

	if err := linkOne(src, dst, false, false); err != nil {
		t.Fatalf("linkOne: %v", err)
	}

	target, err := os.Readlink(dst)
	if err != nil {
		t.Fatalf("dst is not a symlink: %v", err)
	}
	if target != src {
		t.Fatalf("dst links to %s, want %s", target, src)
	}
}

func TestLinkOne_DryRunLeavesMissingDstUntouched(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.conf")
	dst := filepath.Join(dir, "dst.conf")
	mustWriteFile(t, src, "content")

	if err := linkOne(src, dst, true, false); err != nil {
		t.Fatalf("linkOne: %v", err)
	}

	if _, exists := lstatMode(t, dst); exists {
		t.Fatalf("dry-run created %s", dst)
	}
}

func TestLinkOne_SymlinkAlreadyCorrectIsNoop(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.conf")
	dst := filepath.Join(dir, "dst.conf")
	mustWriteFile(t, src, "content")
	if err := os.Symlink(src, dst); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	if err := linkOne(src, dst, false, false); err != nil {
		t.Fatalf("linkOne: %v", err)
	}

	target, err := os.Readlink(dst)
	if err != nil {
		t.Fatalf("dst is not a symlink: %v", err)
	}
	if target != src {
		t.Fatalf("dst links to %s, want %s", target, src)
	}
}

func TestLinkOne_SymlinkConflictWithoutForceErrors(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.conf")
	other := filepath.Join(dir, "other.conf")
	dst := filepath.Join(dir, "dst.conf")
	mustWriteFile(t, src, "content")
	mustWriteFile(t, other, "elsewhere")
	if err := os.Symlink(other, dst); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	if err := linkOne(src, dst, false, false); err == nil {
		t.Fatal("expected error for conflicting symlink without -force")
	}

	target, err := os.Readlink(dst)
	if err != nil {
		t.Fatalf("dst is not a symlink: %v", err)
	}
	if target != other {
		t.Fatalf("dst was changed to %s, want untouched at %s", target, other)
	}
}

func TestLinkOne_SymlinkConflictWithForceRelinks(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.conf")
	other := filepath.Join(dir, "other.conf")
	dst := filepath.Join(dir, "dst.conf")
	mustWriteFile(t, src, "content")
	mustWriteFile(t, other, "elsewhere")
	if err := os.Symlink(other, dst); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	if err := linkOne(src, dst, false, true); err != nil {
		t.Fatalf("linkOne: %v", err)
	}

	target, err := os.Readlink(dst)
	if err != nil {
		t.Fatalf("dst is not a symlink: %v", err)
	}
	if target != src {
		t.Fatalf("dst links to %s, want %s", target, src)
	}
}

func TestLinkOne_SymlinkConflictWithForceDryRunLeavesUntouched(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.conf")
	other := filepath.Join(dir, "other.conf")
	dst := filepath.Join(dir, "dst.conf")
	mustWriteFile(t, src, "content")
	mustWriteFile(t, other, "elsewhere")
	if err := os.Symlink(other, dst); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	if err := linkOne(src, dst, true, true); err != nil {
		t.Fatalf("linkOne: %v", err)
	}

	target, err := os.Readlink(dst)
	if err != nil {
		t.Fatalf("dst is not a symlink: %v", err)
	}
	if target != other {
		t.Fatalf("dry-run changed dst to %s, want untouched at %s", target, other)
	}
}

func TestLinkOne_DirectoryConflictBacksUpAndLinks(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.conf")
	dst := filepath.Join(dir, "dst.conf")
	mustWriteFile(t, src, "content")
	if err := os.Mkdir(dst, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	mustWriteFile(t, filepath.Join(dst, "inside"), "kept")

	if err := linkOne(src, dst, false, false); err != nil {
		t.Fatalf("linkOne: %v", err)
	}

	target, err := os.Readlink(dst)
	if err != nil {
		t.Fatalf("dst is not a symlink: %v", err)
	}
	if target != src {
		t.Fatalf("dst links to %s, want %s", target, src)
	}

	matches, err := filepath.Glob(dst + ".dotlink-bak-*")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("got %d backups, want 1", len(matches))
	}
	got, err := os.ReadFile(filepath.Join(matches[0], "inside"))
	if err != nil {
		t.Fatalf("reading backed up file: %v", err)
	}
	if string(got) != "kept" {
		t.Fatalf("backed up content = %q, want %q", got, "kept")
	}
}

func TestLinkOne_IdenticalFileReplacesWithoutBackup(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.conf")
	dst := filepath.Join(dir, "dst.conf")
	mustWriteFile(t, src, "same content")
	mustWriteFile(t, dst, "same content")

	if err := linkOne(src, dst, false, false); err != nil {
		t.Fatalf("linkOne: %v", err)
	}

	target, err := os.Readlink(dst)
	if err != nil {
		t.Fatalf("dst is not a symlink: %v", err)
	}
	if target != src {
		t.Fatalf("dst links to %s, want %s", target, src)
	}

	matches, err := filepath.Glob(dst + ".dotlink-bak-*")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("got %d backups for identical content, want 0", len(matches))
	}
}

func TestLinkOne_DifferingFileBacksUpAndLinks(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.conf")
	dst := filepath.Join(dir, "dst.conf")
	mustWriteFile(t, src, "new content")
	mustWriteFile(t, dst, "old content")

	if err := linkOne(src, dst, false, false); err != nil {
		t.Fatalf("linkOne: %v", err)
	}

	target, err := os.Readlink(dst)
	if err != nil {
		t.Fatalf("dst is not a symlink: %v", err)
	}
	if target != src {
		t.Fatalf("dst links to %s, want %s", target, src)
	}

	matches, err := filepath.Glob(dst + ".dotlink-bak-*")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("got %d backups, want 1", len(matches))
	}
	got, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("reading backup: %v", err)
	}
	if string(got) != "old content" {
		t.Fatalf("backup content = %q, want %q", got, "old content")
	}
}

func TestLinkOne_DifferingFileDryRunLeavesUntouched(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.conf")
	dst := filepath.Join(dir, "dst.conf")
	mustWriteFile(t, src, "new content")
	mustWriteFile(t, dst, "old content")

	if err := linkOne(src, dst, true, false); err != nil {
		t.Fatalf("linkOne: %v", err)
	}

	mode, exists := lstatMode(t, dst)
	if !exists {
		t.Fatal("dry-run removed dst")
	}
	if mode&os.ModeSymlink != 0 {
		t.Fatal("dry-run turned dst into a symlink")
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("reading dst: %v", err)
	}
	if string(got) != "old content" {
		t.Fatalf("dst content = %q, want %q", got, "old content")
	}

	matches, err := filepath.Glob(dst + ".dotlink-bak-*")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("dry-run created %d backups, want 0", len(matches))
	}
}
