package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMakeAbsolutePathFromRoot_AbsoluteUnchanged(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	abs := filepath.Join(cwd, "some", "path")
	got := makeAbsolutePathFromRoot(abs, cwd)
	if got != abs {
		t.Fatalf("expected absolute path unchanged, got %s want %s", got, abs)
	}
}

func TestMakeAbsolutePathFromRoot_RelativeResolved(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	rel := filepath.Join(".", "mount", "resources", "media")
	got := makeAbsolutePathFromRoot(rel, cwd)
	want := filepath.Join(cwd, rel)
	if got != want {
		t.Fatalf("expected relative path resolved, got %s want %s", got, want)
	}
}

func TestMakeAbsoluteConnectionStringFromRoot_NonFileUnchanged(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	cs := "memory:db"
	if got := makeAbsoluteConnectionStringFromRoot(cs, cwd); got != cs {
		t.Fatalf("expected non-file conn string unchanged, got %s want %s", got, cs)
	}
}

func TestMakeAbsoluteConnectionStringFromRoot_AbsoluteFileUnchanged(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	absFile := "file:" + filepath.Join(cwd, "mount", "database", "db.sqlite")
	if got := makeAbsoluteConnectionStringFromRoot(absFile, cwd); got != absFile {
		t.Fatalf("expected absolute file conn string unchanged, got %s want %s", got, absFile)
	}
}

func TestMakeAbsoluteConnectionStringFromRoot_RelativeFileResolved(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	relFile := filepath.Join(".", "mount", "database", "db.sqlite")
	got := makeAbsoluteConnectionStringFromRoot("file:"+relFile, cwd)
	want := "file:" + filepath.Join(cwd, relFile)
	if got != want {
		t.Fatalf("expected relative file conn string resolved, got %s want %s", got, want)
	}
}

func TestExtractPathFromConnectionString_NonFileEmpty(t *testing.T) {
	if got := extractPathFromConnectionString("memory:db"); got != "" {
		t.Fatalf("expected empty for non-file conn string, got %s", got)
	}
}

func TestExtractPathFromConnectionString_FileStripsPrefix(t *testing.T) {
	want := filepath.Join(".", "mount", "database", "db.sqlite")
	cs := "file:" + want
	if got := extractPathFromConnectionString(cs); got != want {
		t.Fatalf("expected extracted file path, got %s want %s", got, want)
	}
}

func TestCreateDirectoriesFromPaths(t *testing.T) {
	tmp := t.TempDir()
	paths := []string{
		filepath.Join(tmp, "a", "b"),
		filepath.Join(tmp, "c"),
	}
	if err := createDirectoriesFromPaths(paths); err != nil {
		t.Fatalf("createDirectoriesFromPaths returned error: %v", err)
	}
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			t.Fatalf("expected directory to exist: %s err: %v", p, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected a directory at %s", p)
		}
	}
}