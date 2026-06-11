package scaffold

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jhl-labs/docs-cli/internal/schema"
)

func TestBuildDefaultsOutputDirAndUsesCommit(t *testing.T) {
	files := Build(schema.Standard(), Options{Date: "2026-06-12", Commit: "deadbee"})
	if len(files) == 0 {
		t.Fatal("no files")
	}
	// Empty OutputDir must default to "docs".
	if !strings.HasPrefix(files[0].Path, "docs"+string(filepath.Separator)) {
		t.Errorf("first path %q does not start with docs/", files[0].Path)
	}
	// The provided commit must appear in a document's frontmatter.
	var sawCommit bool
	for _, f := range files {
		if strings.Contains(f.Content, "source_commit: deadbee") {
			sawCommit = true
			break
		}
	}
	if !sawCommit {
		t.Error("provided commit not written into frontmatter")
	}
}

func TestWriteCreatesNestedDirs(t *testing.T) {
	dir := t.TempDir()
	files := []File{{Path: filepath.Join(dir, "a", "b", "c.md"), Content: "x"}}
	written, _, err := Write(files, false)
	if err != nil || len(written) != 1 {
		t.Fatalf("write nested: written=%v err=%v", written, err)
	}
}
