// Package scaffold turns the standardized schema into a concrete docs/ tree:
// an index, one Markdown file per document (frontmatter + chapter stubs), and
// the adr/ and diagrams/ collection directories with starter files.
package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jhl-labs/docs-cli/internal/mddoc"
	"github.com/jhl-labs/docs-cli/internal/schema"
)

// Options configure generation.
type Options struct {
	// OutputDir is the docs root, e.g. "docs".
	OutputDir string
	// Lang is the detected primary language (used in frontmatter notes).
	Lang string
	// Date is an injected timestamp (ISO date); keeps generation deterministic.
	Date string
	// Commit is the source commit recorded in frontmatter, if known.
	Commit string
	// Force overwrites existing files when true.
	Force bool
}

// File is a generated file: a path relative to OutputDir's parent (it already
// includes OutputDir) and its content.
type File struct {
	Path    string
	Content string
}

// Build returns every file the standardized tree contains, without touching
// disk. The result is deterministic for a given schema and Options.
func Build(s schema.Schema, opts Options) []File {
	if opts.OutputDir == "" {
		opts.OutputDir = "docs"
	}
	files := []File{
		{Path: filepath.Join(opts.OutputDir, "README.md"), Content: buildIndex(s, opts)},
	}
	for _, doc := range s.Docs {
		files = append(files, File{
			Path:    filepath.Join(opts.OutputDir, doc.FileName()),
			Content: buildDocument(s, doc, opts),
		})
	}
	// Starter files for the collection directories so the folder tree is concrete.
	files = append(files,
		File{Path: filepath.Join(opts.OutputDir, "adr", "ADR-0001-record-architecture-decisions.md"), Content: starterADR(opts)},
		File{Path: filepath.Join(opts.OutputDir, "diagrams", "context.md"), Content: starterDiagram(opts)},
	)
	return files
}

// Write persists the files to disk. It creates parent directories as needed and
// skips files that already exist unless opts.Force is set. It returns the paths
// written and the paths skipped.
func Write(files []File, force bool) (written, skipped []string, err error) {
	for _, f := range files {
		if !force {
			if _, statErr := os.Stat(f.Path); statErr == nil {
				skipped = append(skipped, f.Path)
				continue
			}
		}
		if mkErr := os.MkdirAll(filepath.Dir(f.Path), 0o755); mkErr != nil {
			return written, skipped, fmt.Errorf("create dir for %s: %w", f.Path, mkErr)
		}
		if writeErr := os.WriteFile(f.Path, []byte(f.Content), 0o644); writeErr != nil {
			return written, skipped, fmt.Errorf("write %s: %w", f.Path, writeErr)
		}
		written = append(written, f.Path)
	}
	return written, skipped, nil
}

func buildDocument(s schema.Schema, doc schema.Document, opts Options) string {
	fm := mddoc.Frontmatter{
		DocID:         doc.ID,
		Title:         doc.Title,
		Section:       doc.Section,
		Order:         doc.Order,
		Audience:      doc.Audience,
		Status:        schema.StatusDraft,
		SchemaVersion: s.Version,
		GeneratedBy:   "docs-cli template",
		SourceCommit:  defaultStr(opts.Commit, "unknown"),
		Updated:       opts.Date,
	}

	var b strings.Builder
	b.WriteString(fm.Marshal())
	b.WriteString("\n# ")
	b.WriteString(doc.Title)
	b.WriteString("\n\n> ")
	b.WriteString(doc.Purpose)
	b.WriteString("\n>\n> _이 문서는 `docs-cli`가 표준 스키마 v")
	b.WriteString(s.Version)
	b.WriteString("로 스캐폴딩한 초안입니다. 각 챕터의 안내를 따라 채우거나 `docs-cli generate`로 에이전트가 채우게 하세요._\n")

	for _, ch := range doc.Chapters {
		b.WriteString("\n## ")
		b.WriteString(ch.Heading)
		b.WriteString("\n\n> **작성 안내:** ")
		b.WriteString(ch.Guidance)
		b.WriteString("\n\n_TODO(docs-cli): 내용을 작성하세요._\n")
	}
	return b.String()
}

func buildIndex(s schema.Schema, opts Options) string {
	var b strings.Builder
	b.WriteString("# 프로젝트 문서 (docs)\n\n")
	b.WriteString("> 이 폴더는 `docs-cli`가 정의한 **표준 문서 스키마 v")
	b.WriteString(s.Version)
	b.WriteString("**를 따릅니다. 모든 문서는 동일한 프론트매터 키와 챕터 구조를 가지므로, ")
	b.WriteString("웹 서비스로 그대로 이식할 수 있습니다.\n\n")
	if opts.Lang != "" && opts.Lang != "unknown" {
		b.WriteString(fmt.Sprintf("> 감지된 주 언어: `%s`\n\n", opts.Lang))
	}

	b.WriteString("## 문서 지도\n\n")
	sections := append([]schema.Section(nil), s.Sections...)
	sort.Slice(sections, func(i, j int) bool { return sections[i].Order < sections[j].Order })
	for _, sec := range sections {
		b.WriteString(fmt.Sprintf("### %d. %s\n\n", sec.Order, sec.Title))
		b.WriteString(fmt.Sprintf("> %s\n\n", sec.Blurb))
		docs := s.DocsInSection(sec.ID)
		sort.Slice(docs, func(i, j int) bool { return docs[i].Order < docs[j].Order })
		b.WriteString("| 문서 | 목적 | 필수 |\n| --- | --- | --- |\n")
		for _, d := range docs {
			req := ""
			if d.Required {
				req = "✅"
			}
			b.WriteString(fmt.Sprintf("| [%s](./%s) | %s | %s |\n", d.Title, d.FileName(), d.Purpose, req))
		}
		b.WriteString("\n")
	}

	b.WriteString("## 문서 상태\n\n")
	b.WriteString("> `docs-cli validate`로 정합성을, `docs-cli render`로 HTML/XML 산출물을 만드세요.\n\n")
	b.WriteString("| 문서 | 상태 |\n| --- | --- |\n")
	for _, d := range s.Docs {
		b.WriteString(fmt.Sprintf("| %s | %s |\n", d.Title, schema.StatusDraft))
	}
	b.WriteString("\n---\n\n")
	b.WriteString(fmt.Sprintf("> 생성: `docs-cli` · 스키마 v%s · %s\n", s.Version, opts.Date))
	return b.String()
}

func starterADR(opts Options) string {
	return strings.Join([]string{
		"# ADR-0001: 아키텍처 결정 기록을 사용한다",
		"",
		"- **상태:** 수락 (Accepted)",
		"- **날짜:** " + opts.Date,
		"",
		"## 맥락 (Context)",
		"",
		"이 프로젝트의 주요 설계 결정을 추적 가능한 형태로 남길 필요가 있다.",
		"",
		"## 결정 (Decision)",
		"",
		"모든 구조적 결정은 `docs/adr/ADR-NNNN-<slug>.md` 파일로 기록하고,",
		"[adr/README.md](./README.md) 인덱스에 한 줄을 추가한다.",
		"",
		"## 결과 (Consequences)",
		"",
		"- 결정의 맥락이 보존되어 신규 기여자가 \"왜\"를 추적할 수 있다.",
		"- 결정을 바꿀 때는 기존 ADR을 폐기(Superseded)로 표시하고 새 ADR을 추가한다.",
		"",
	}, "\n")
}

func starterDiagram(opts Options) string {
	return strings.Join([]string{
		"# 시스템 컨텍스트 다이어그램",
		"",
		"> 외부 행위자·시스템과 이 프로젝트의 경계를 보여준다.",
		"",
		"```mermaid",
		"graph LR",
		"    user[사용자] --> system[이 시스템]",
		"    system --> dep[외부 의존성]",
		"```",
		"",
		"_TODO(docs-cli): 실제 컨텍스트로 교체하세요._",
		"",
	}, "\n")
}

func defaultStr(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
