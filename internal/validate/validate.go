// Package validate checks that a docs/ tree conforms to the standardized
// schema: required documents exist, frontmatter carries the mandatory keys with
// values consistent with the schema, the declared chapters are present, and no
// scaffold TODO placeholders remain in documents marked reviewed.
package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jhl-labs/docs-cli/internal/mddoc"
	"github.com/jhl-labs/docs-cli/internal/schema"
)

// Severity classifies a finding.
type Severity string

const (
	Error Severity = "error"
	Warn  Severity = "warn"
)

// Finding is a single validation result.
type Finding struct {
	Severity Severity
	Doc      string // doc id or file path
	Message  string
}

// Report aggregates findings.
type Report struct {
	Findings []Finding
}

// Errors returns the number of error-severity findings.
func (r Report) Errors() int {
	n := 0
	for _, f := range r.Findings {
		if f.Severity == Error {
			n++
		}
	}
	return n
}

// Warnings returns the number of warn-severity findings.
func (r Report) Warnings() int {
	n := 0
	for _, f := range r.Findings {
		if f.Severity == Warn {
			n++
		}
	}
	return n
}

func (r *Report) add(sev Severity, doc, format string, args ...any) {
	r.Findings = append(r.Findings, Finding{Severity: sev, Doc: doc, Message: fmt.Sprintf(format, args...)})
}

// Run validates the docs tree rooted at dir against the schema.
func Run(s schema.Schema, dir string) Report {
	var r Report
	if dir == "" {
		dir = "docs"
	}

	if _, err := os.Stat(dir); err != nil {
		r.add(Error, dir, "docs 디렉터리를 찾을 수 없습니다: %v", err)
		return r
	}

	indexPath := filepath.Join(dir, "README.md")
	if _, err := os.Stat(indexPath); err != nil {
		r.add(Error, "README.md", "문서 인덱스(README.md)가 없습니다 — `docs-cli init`을 먼저 실행하세요")
	}

	for _, doc := range s.Docs {
		path := filepath.Join(dir, doc.FileName())
		raw, err := os.ReadFile(path)
		if err != nil {
			sev := Warn
			if doc.Required {
				sev = Error
			}
			r.add(sev, doc.ID, "문서 파일이 없습니다: %s", doc.FileName())
			continue
		}
		validateDocument(&r, s, doc, raw)
	}

	return r
}

func validateDocument(r *Report, s schema.Schema, doc schema.Document, raw []byte) {
	parsed, hasFM, err := mddoc.Parse(raw)
	if err != nil {
		r.add(Error, doc.ID, "파싱 실패: %v", err)
		return
	}
	if !hasFM {
		r.add(Error, doc.ID, "프론트매터(---) 블록이 없습니다")
		return
	}

	fm := parsed.Frontmatter
	if fm.DocID != doc.ID {
		r.add(Error, doc.ID, "doc_id 불일치: 프론트매터=%q, 스키마=%q", fm.DocID, doc.ID)
	}
	if fm.Section != doc.Section {
		r.add(Error, doc.ID, "section 불일치: 프론트매터=%q, 스키마=%q", fm.Section, doc.Section)
	}
	if fm.SchemaVersion != s.Version {
		r.add(Warn, doc.ID, "schema_version=%q (현재 표준은 v%s)", fm.SchemaVersion, s.Version)
	}
	if !validStatus(fm.Status) {
		r.add(Error, doc.ID, "status=%q 는 유효하지 않습니다 (draft|generated|reviewed)", fm.Status)
	}
	if strings.TrimSpace(fm.Title) == "" {
		r.add(Error, doc.ID, "title 이 비어 있습니다")
	}

	// Collection index documents (adr, diagrams) only require the index chapter.
	headings := normalizeSet(parsed.H2())
	for _, ch := range doc.Chapters {
		if _, ok := headings[strings.ToLower(strings.TrimSpace(ch.Heading))]; !ok {
			r.add(Error, doc.ID, "필수 챕터 누락: %q", ch.Heading)
		}
	}

	// A reviewed document must not still carry scaffold placeholders.
	if fm.Status == schema.StatusReviewed {
		body := string(raw)
		if strings.Contains(body, "TODO(docs-cli)") {
			r.add(Error, doc.ID, "status=reviewed 인데 스캐폴드 자리표시자(TODO(docs-cli))가 남아 있습니다")
		}
	}
}

func validStatus(status string) bool {
	switch status {
	case schema.StatusDraft, schema.StatusGenerated, schema.StatusReviewed:
		return true
	default:
		return false
	}
}

func normalizeSet(items []string) map[string]struct{} {
	set := make(map[string]struct{}, len(items))
	for _, it := range items {
		set[strings.ToLower(strings.TrimSpace(it))] = struct{}{}
	}
	return set
}

// FormatText renders a report as human-readable lines, sorted by severity then doc.
func FormatText(r Report) string {
	findings := append([]Finding(nil), r.Findings...)
	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Severity != findings[j].Severity {
			return findings[i].Severity == Error
		}
		return findings[i].Doc < findings[j].Doc
	})
	var b strings.Builder
	for _, f := range findings {
		b.WriteString(fmt.Sprintf("[%s] %s: %s\n", f.Severity, f.Doc, f.Message))
	}
	b.WriteString(fmt.Sprintf("\n%d error(s), %d warning(s)\n", r.Errors(), r.Warnings()))
	return b.String()
}
