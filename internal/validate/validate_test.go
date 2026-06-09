package validate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jhl-labs/docs-cli/internal/scaffold"
	"github.com/jhl-labs/docs-cli/internal/schema"
)

func scaffolded(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs")
	files := scaffold.Build(schema.Standard(), scaffold.Options{OutputDir: docsDir, Date: "2026-06-09"})
	if _, _, err := scaffold.Write(files, true); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	return docsDir
}

func TestFreshScaffoldHasNoErrors(t *testing.T) {
	report := Run(schema.Standard(), scaffolded(t))
	if report.Errors() != 0 {
		t.Fatalf("fresh scaffold has errors:\n%s", FormatText(report))
	}
}

func TestMissingRequiredDocIsError(t *testing.T) {
	docsDir := scaffolded(t)
	if err := os.Remove(filepath.Join(docsDir, "overview.md")); err != nil {
		t.Fatal(err)
	}
	report := Run(schema.Standard(), docsDir)
	if report.Errors() == 0 {
		t.Fatal("expected error for missing required doc")
	}
}

func TestMissingChapterIsError(t *testing.T) {
	docsDir := scaffolded(t)
	// Overwrite overview.md, dropping a required chapter heading.
	bad := "---\ndoc_id: overview\ntitle: 개요\nsection: overview\norder: 1\naudience: [newcomer]\nstatus: draft\nschema_version: 1\ngenerated_by: x\nsource_commit: x\nupdated: 2026-06-09\n---\n\n# 개요\n\n## TL;DR\n\nonly one chapter\n"
	if err := os.WriteFile(filepath.Join(docsDir, "overview.md"), []byte(bad), 0o644); err != nil {
		t.Fatal(err)
	}
	report := Run(schema.Standard(), docsDir)
	if report.Errors() == 0 {
		t.Fatal("expected error for missing chapter")
	}
}

func TestReviewedWithTODOIsError(t *testing.T) {
	docsDir := scaffolded(t)
	raw, err := os.ReadFile(filepath.Join(docsDir, "overview.md"))
	if err != nil {
		t.Fatal(err)
	}
	// Flip status to reviewed while TODO placeholders remain.
	updated := []byte(string(raw))
	updated = []byte(replaceFirst(string(updated), "status: draft", "status: reviewed"))
	if err := os.WriteFile(filepath.Join(docsDir, "overview.md"), updated, 0o644); err != nil {
		t.Fatal(err)
	}
	report := Run(schema.Standard(), docsDir)
	if report.Errors() == 0 {
		t.Fatal("expected error for reviewed doc with TODO placeholders")
	}
}

func TestMissingDocsDirIsError(t *testing.T) {
	report := Run(schema.Standard(), filepath.Join(t.TempDir(), "nope"))
	if report.Errors() == 0 {
		t.Fatal("expected error for missing docs dir")
	}
}

func replaceFirst(s, old, new string) string {
	i := indexOf(s, old)
	if i < 0 {
		return s
	}
	return s[:i] + new + s[i+len(old):]
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
