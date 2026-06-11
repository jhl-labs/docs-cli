package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jhl-labs/docs-cli/internal/schema"
)

// writeOverview overwrites docs/overview.md in a scaffolded tree with the given
// frontmatter body, keeping the required chapters so only the targeted defect
// is exercised.
func writeOverview(t *testing.T, docsDir, frontmatter string) {
	t.Helper()
	body := frontmatter + "\n# 개요\n\n## TL;DR\n\nx\n\n## 문제 정의\n\nx\n\n## 사용자와 이해관계자\n\nx\n\n## 핵심 가치와 비-목표\n\nx\n\n## 기술 스택 요약\n\nx\n"
	if err := os.WriteFile(filepath.Join(docsDir, "overview.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func findingsFor(r Report, doc string) []Finding {
	var out []Finding
	for _, f := range r.Findings {
		if f.Doc == doc {
			out = append(out, f)
		}
	}
	return out
}

func TestValidateNoFrontmatter(t *testing.T) {
	dir := scaffolded(t)
	if err := os.WriteFile(filepath.Join(dir, "overview.md"), []byte("# 개요\n\n## TL;DR\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := findingsFor(Run(schema.Standard(), dir), "overview"); !hasMessage(got, "프론트매터") {
		t.Errorf("expected frontmatter error, got %v", got)
	}
}

func TestValidateParseError(t *testing.T) {
	dir := scaffolded(t)
	if err := os.WriteFile(filepath.Join(dir, "overview.md"), []byte("---\ndoc_id: overview\n# never closed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := findingsFor(Run(schema.Standard(), dir), "overview"); !hasMessage(got, "파싱 실패") {
		t.Errorf("expected parse error, got %v", got)
	}
}

func TestValidateSectionAndIDMismatch(t *testing.T) {
	dir := scaffolded(t)
	writeOverview(t, dir, "---\ndoc_id: wrong\ntitle: 개요\nsection: nope\norder: 1\naudience: [x]\nstatus: draft\nschema_version: 1\ngenerated_by: t\nsource_commit: x\nupdated: 2026-06-12\n---")
	got := findingsFor(Run(schema.Standard(), dir), "overview")
	if !hasMessage(got, "doc_id 불일치") || !hasMessage(got, "section 불일치") {
		t.Errorf("expected doc_id/section mismatch, got %v", got)
	}
}

func TestValidateSchemaVersionWarn(t *testing.T) {
	dir := scaffolded(t)
	writeOverview(t, dir, "---\ndoc_id: overview\ntitle: 개요\nsection: overview\norder: 1\naudience: [x]\nstatus: draft\nschema_version: 0\ngenerated_by: t\nsource_commit: x\nupdated: 2026-06-12\n---")
	got := findingsFor(Run(schema.Standard(), dir), "overview")
	if !hasSeverity(got, Warn) {
		t.Errorf("expected schema_version warning, got %v", got)
	}
}

func TestValidateInvalidStatusAndEmptyTitle(t *testing.T) {
	dir := scaffolded(t)
	writeOverview(t, dir, "---\ndoc_id: overview\ntitle: \nsection: overview\norder: 1\naudience: [x]\nstatus: bogus\nschema_version: 1\ngenerated_by: t\nsource_commit: x\nupdated: 2026-06-12\n---")
	got := findingsFor(Run(schema.Standard(), dir), "overview")
	if !hasMessage(got, "유효하지 않") || !hasMessage(got, "title") {
		t.Errorf("expected invalid status + empty title, got %v", got)
	}
}

func TestValidateOptionalDocMissingIsWarn(t *testing.T) {
	dir := scaffolded(t)
	// data-model is optional; removing it should warn, not error.
	if err := os.Remove(filepath.Join(dir, "data-model.md")); err != nil {
		t.Fatal(err)
	}
	got := findingsFor(Run(schema.Standard(), dir), "data-model")
	if len(got) != 1 || got[0].Severity != Warn {
		t.Errorf("expected single warning for optional doc, got %v", got)
	}
}

func TestReportCountersAndFormat(t *testing.T) {
	var r Report
	r.add(Error, "a", "boom")
	r.add(Warn, "b", "careful")
	if r.Errors() != 1 || r.Warnings() != 1 {
		t.Fatalf("counts: errors=%d warnings=%d", r.Errors(), r.Warnings())
	}
	out := FormatText(r)
	if !strings.Contains(out, "[error] a: boom") || !strings.Contains(out, "[warn] b: careful") {
		t.Errorf("FormatText missing findings:\n%s", out)
	}
	if !strings.Contains(out, "1 error(s), 1 warning(s)") {
		t.Errorf("FormatText missing summary:\n%s", out)
	}
}

func hasMessage(fs []Finding, sub string) bool {
	for _, f := range fs {
		if strings.Contains(f.Message, sub) {
			return true
		}
	}
	return false
}

func hasSeverity(fs []Finding, sev Severity) bool {
	for _, f := range fs {
		if f.Severity == sev {
			return true
		}
	}
	return false
}
