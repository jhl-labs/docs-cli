// Package mddoc parses and serializes the standardized Markdown documents that
// docs-cli produces: a YAML-style frontmatter block followed by a Markdown body
// organized into headings. It intentionally supports only the small, fixed
// frontmatter vocabulary the schema defines, so it has no external dependencies.
package mddoc

import (
	"fmt"
	"strconv"
	"strings"
)

// Frontmatter holds the standardized metadata block. The keys are fixed so a
// downstream web service can ingest any generated document without per-project
// mapping.
type Frontmatter struct {
	DocID         string
	Title         string
	Section       string
	Order         int
	Audience      []string
	Status        string
	SchemaVersion string
	GeneratedBy   string
	SourceCommit  string
	Updated       string
	// extra preserves unknown keys in declaration order so round-tripping a
	// document never silently drops metadata.
	extra []kv
}

type kv struct {
	key   string
	value string
}

// Heading is a single Markdown heading and the body lines that follow it, up to
// the next heading of any level.
type Heading struct {
	Level int
	Text  string
	// Body holds the raw Markdown lines between this heading and the next one.
	Body []string
}

// Document is a parsed standardized Markdown document.
type Document struct {
	Frontmatter Frontmatter
	// Preamble holds body lines before the first heading (usually empty).
	Preamble []string
	Headings []Heading
}

// H2 returns the H2 heading texts in order, trimmed.
func (d Document) H2() []string {
	var out []string
	for _, h := range d.Headings {
		if h.Level == 2 {
			out = append(out, strings.TrimSpace(h.Text))
		}
	}
	return out
}

// Parse splits raw bytes into frontmatter and a heading tree. A document
// without a frontmatter fence parses with an empty Frontmatter and HasFrontmatter
// false.
func Parse(raw []byte) (Document, bool, error) {
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	lines := strings.Split(text, "\n")

	var doc Document
	hasFrontmatter := false
	bodyStart := 0

	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		end := -1
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				end = i
				break
			}
		}
		if end == -1 {
			return doc, false, fmt.Errorf("unterminated frontmatter block")
		}
		fm, err := parseFrontmatter(lines[1:end])
		if err != nil {
			return doc, false, err
		}
		doc.Frontmatter = fm
		hasFrontmatter = true
		bodyStart = end + 1
	}

	parseBody(lines[bodyStart:], &doc)
	return doc, hasFrontmatter, nil
}

func parseBody(lines []string, doc *Document) {
	var current *Heading
	for _, line := range lines {
		if level, text, ok := headingLine(line); ok {
			doc.Headings = append(doc.Headings, Heading{Level: level, Text: text})
			current = &doc.Headings[len(doc.Headings)-1]
			continue
		}
		if current == nil {
			doc.Preamble = append(doc.Preamble, line)
		} else {
			current.Body = append(current.Body, line)
		}
	}
}

// headingLine reports whether a line is an ATX heading (# .. ######) and returns
// its level and trimmed text. Lines inside fenced code blocks are not detected
// here; callers that need code-aware parsing should pre-strip fences. For the
// standardized documents this is sufficient because chapter headings never
// appear inside code fences.
func headingLine(line string) (int, string, bool) {
	trimmed := strings.TrimRight(line, " ")
	i := 0
	for i < len(trimmed) && trimmed[i] == '#' {
		i++
	}
	if i == 0 || i > 6 {
		return 0, "", false
	}
	if i >= len(trimmed) || trimmed[i] != ' ' {
		return 0, "", false
	}
	return i, strings.TrimSpace(trimmed[i+1:]), true
}

func parseFrontmatter(lines []string) (Frontmatter, error) {
	var fm Frontmatter
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		idx := strings.Index(line, ":")
		if idx == -1 {
			return fm, fmt.Errorf("invalid frontmatter line: %q", line)
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		value = strings.Trim(value, `"`)
		switch key {
		case "doc_id":
			fm.DocID = value
		case "title":
			fm.Title = value
		case "section":
			fm.Section = value
		case "order":
			n, err := strconv.Atoi(value)
			if err != nil {
				return fm, fmt.Errorf("invalid order %q: %w", value, err)
			}
			fm.Order = n
		case "audience":
			fm.Audience = parseList(value)
		case "status":
			fm.Status = value
		case "schema_version":
			fm.SchemaVersion = value
		case "generated_by":
			fm.GeneratedBy = value
		case "source_commit":
			fm.SourceCommit = value
		case "updated":
			fm.Updated = value
		default:
			fm.extra = append(fm.extra, kv{key: key, value: value})
		}
	}
	return fm, nil
}

func parseList(value string) []string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Marshal renders the frontmatter as a fenced YAML-style block, terminated with
// a trailing newline. Keys are emitted in a fixed, stable order.
func (fm Frontmatter) Marshal() string {
	var b strings.Builder
	b.WriteString("---\n")
	writeScalar(&b, "doc_id", fm.DocID)
	writeScalar(&b, "title", fm.Title)
	writeScalar(&b, "section", fm.Section)
	b.WriteString(fmt.Sprintf("order: %d\n", fm.Order))
	writeList(&b, "audience", fm.Audience)
	writeScalar(&b, "status", fm.Status)
	writeScalar(&b, "schema_version", fm.SchemaVersion)
	writeScalar(&b, "generated_by", fm.GeneratedBy)
	writeScalar(&b, "source_commit", fm.SourceCommit)
	writeScalar(&b, "updated", fm.Updated)
	for _, e := range fm.extra {
		writeScalar(&b, e.key, e.value)
	}
	b.WriteString("---\n")
	return b.String()
}

func writeScalar(b *strings.Builder, key, value string) {
	b.WriteString(key)
	b.WriteString(": ")
	b.WriteString(value)
	b.WriteString("\n")
}

func writeList(b *strings.Builder, key string, values []string) {
	b.WriteString(key)
	b.WriteString(": [")
	b.WriteString(strings.Join(values, ", "))
	b.WriteString("]\n")
}
