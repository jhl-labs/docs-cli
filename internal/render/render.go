// Package render converts the standardized Markdown documents into the output
// formats docs-cli supports: HTML (for a docs site) and XML (a structural
// representation for a web service to ingest). The Markdown subset handled here
// is exactly what the scaffold templates and generated docs use: frontmatter,
// ATX headings, fenced code, blockquotes, ordered/unordered lists, pipe tables,
// horizontal rules, and paragraphs with inline code/bold/italic/links.
package render

import (
	"encoding/xml"
	"fmt"
	"html"
	"strings"

	"github.com/jhl-labs/docs-cli/internal/mddoc"
)

// HTML renders a standardized document to a full HTML page. The frontmatter
// title becomes the <title> and <h1>.
func HTML(raw []byte) (string, error) {
	doc, _, err := mddoc.Parse(raw)
	if err != nil {
		return "", err
	}
	body := stripFrontmatter(string(raw))
	inner := markdownToHTML(body)

	title := doc.Frontmatter.Title
	if title == "" {
		title = "Document"
	}
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"ko\">\n<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	b.WriteString(fmt.Sprintf("<title>%s</title>\n", html.EscapeString(title)))
	if doc.Frontmatter.DocID != "" {
		b.WriteString(fmt.Sprintf("<meta name=\"doc-id\" content=\"%s\">\n", html.EscapeString(doc.Frontmatter.DocID)))
	}
	if doc.Frontmatter.Section != "" {
		b.WriteString(fmt.Sprintf("<meta name=\"doc-section\" content=\"%s\">\n", html.EscapeString(doc.Frontmatter.Section)))
	}
	b.WriteString(styleBlock)
	b.WriteString("</head>\n<body>\n<main class=\"doc\">\n")
	b.WriteString(inner)
	b.WriteString("</main>\n</body>\n</html>\n")
	return b.String(), nil
}

// xmlDoc is the structural XML representation of a standardized document.
type xmlDoc struct {
	XMLName       xml.Name     `xml:"document"`
	DocID         string       `xml:"doc-id,attr"`
	Section       string       `xml:"section,attr"`
	Order         int          `xml:"order,attr"`
	Status        string       `xml:"status,attr"`
	SchemaVersion string       `xml:"schema-version,attr"`
	Title         string       `xml:"title"`
	Audience      []string     `xml:"audience>role"`
	Chapters      []xmlChapter `xml:"chapter"`
}

type xmlChapter struct {
	ID       string `xml:"id,attr"`
	Heading  string `xml:"heading"`
	Markdown string `xml:"markdown"`
}

// XML renders a standardized document to its structural XML form. Chapters are
// the H2 sections; their bodies are preserved as Markdown.
func XML(raw []byte) (string, error) {
	doc, _, err := mddoc.Parse(raw)
	if err != nil {
		return "", err
	}
	fm := doc.Frontmatter
	out := xmlDoc{
		DocID:         fm.DocID,
		Section:       fm.Section,
		Order:         fm.Order,
		Status:        fm.Status,
		SchemaVersion: fm.SchemaVersion,
		Title:         fm.Title,
		Audience:      fm.Audience,
	}
	for _, h := range doc.Headings {
		if h.Level != 2 {
			continue
		}
		out.Chapters = append(out.Chapters, xmlChapter{
			ID:       slug(h.Text),
			Heading:  strings.TrimSpace(h.Text),
			Markdown: strings.TrimRight(strings.Join(h.Body, "\n"), "\n"),
		})
	}
	encoded, err := xml.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(encoded) + "\n", nil
}

func stripFrontmatter(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return text
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.Join(lines[i+1:], "\n")
		}
	}
	return text
}

// markdownToHTML converts the supported Markdown subset to HTML.
func markdownToHTML(md string) string {
	md = strings.ReplaceAll(md, "\r\n", "\n")
	lines := strings.Split(md, "\n")
	var b strings.Builder
	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		switch {
		case trimmed == "":
			i++
		case strings.HasPrefix(trimmed, "```"):
			i = renderFence(&b, lines, i)
		case isHeading(line):
			renderHeading(&b, line)
			i++
		case trimmed == "---" || trimmed == "***" || trimmed == "___":
			b.WriteString("<hr>\n")
			i++
		case strings.HasPrefix(trimmed, ">"):
			i = renderBlockquote(&b, lines, i)
		case isTableRow(line) && i+1 < len(lines) && isTableDivider(lines[i+1]):
			i = renderTable(&b, lines, i)
		case isUnorderedItem(trimmed):
			i = renderList(&b, lines, i, false)
		case isOrderedItem(trimmed):
			i = renderList(&b, lines, i, true)
		default:
			i = renderParagraph(&b, lines, i)
		}
	}
	return b.String()
}

func renderFence(b *strings.Builder, lines []string, start int) int {
	open := strings.TrimSpace(lines[start])
	lang := strings.TrimSpace(strings.TrimPrefix(open, "```"))
	var code []string
	i := start + 1
	for i < len(lines) && strings.TrimSpace(lines[i]) != "```" {
		code = append(code, lines[i])
		i++
	}
	if lang != "" {
		b.WriteString(fmt.Sprintf("<pre><code class=\"language-%s\">", html.EscapeString(lang)))
	} else {
		b.WriteString("<pre><code>")
	}
	b.WriteString(html.EscapeString(strings.Join(code, "\n")))
	b.WriteString("</code></pre>\n")
	if i < len(lines) {
		i++ // consume closing fence
	}
	return i
}

func renderHeading(b *strings.Builder, line string) {
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	text := strings.TrimSpace(line[level:])
	b.WriteString(fmt.Sprintf("<h%d id=\"%s\">%s</h%d>\n", level, slug(text), inline(text), level))
}

func renderBlockquote(b *strings.Builder, lines []string, start int) int {
	var content []string
	i := start
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(trimmed, ">") {
			break
		}
		content = append(content, strings.TrimSpace(strings.TrimPrefix(trimmed, ">")))
		i++
	}
	b.WriteString("<blockquote>\n")
	b.WriteString(markdownToHTML(strings.Join(content, "\n")))
	b.WriteString("</blockquote>\n")
	return i
}

func renderList(b *strings.Builder, lines []string, start int, ordered bool) int {
	tag := "ul"
	if ordered {
		tag = "ol"
	}
	b.WriteString("<" + tag + ">\n")
	i := start
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if ordered && !isOrderedItem(trimmed) {
			break
		}
		if !ordered && !isUnorderedItem(trimmed) {
			break
		}
		b.WriteString("<li>" + inline(listItemText(trimmed, ordered)) + "</li>\n")
		i++
	}
	b.WriteString("</" + tag + ">\n")
	return i
}

func renderTable(b *strings.Builder, lines []string, start int) int {
	header := splitRow(lines[start])
	i := start + 2 // skip header + divider
	b.WriteString("<table>\n<thead>\n<tr>")
	for _, cell := range header {
		b.WriteString("<th>" + inline(cell) + "</th>")
	}
	b.WriteString("</tr>\n</thead>\n<tbody>\n")
	for i < len(lines) && isTableRow(lines[i]) {
		b.WriteString("<tr>")
		for _, cell := range splitRow(lines[i]) {
			b.WriteString("<td>" + inline(cell) + "</td>")
		}
		b.WriteString("</tr>\n")
		i++
	}
	b.WriteString("</tbody>\n</table>\n")
	return i
}

func renderParagraph(b *strings.Builder, lines []string, start int) int {
	var content []string
	i := start
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || isHeading(lines[i]) || strings.HasPrefix(trimmed, "```") ||
			strings.HasPrefix(trimmed, ">") || isUnorderedItem(trimmed) || isOrderedItem(trimmed) ||
			isTableRow(lines[i]) {
			break
		}
		content = append(content, trimmed)
		i++
	}
	if len(content) > 0 {
		b.WriteString("<p>" + inline(strings.Join(content, " ")) + "</p>\n")
	}
	if i == start {
		i++ // ensure progress
	}
	return i
}

// inline converts inline Markdown to HTML, escaping along the way and protecting
// inline code spans from emphasis/link processing.
func inline(s string) string {
	var b strings.Builder
	var codes []string
	// Protect inline code spans with placeholders.
	protected := replaceCodeSpans(s, func(code string) string {
		codes = append(codes, code)
		return fmt.Sprintf("\x00%d\x00", len(codes)-1)
	})
	escaped := html.EscapeString(protected)
	escaped = applyLinks(escaped)
	escaped = applyEmphasis(escaped)
	// Restore code spans.
	for idx, code := range codes {
		placeholder := fmt.Sprintf("\x00%d\x00", idx)
		escaped = strings.ReplaceAll(escaped, placeholder, "<code>"+html.EscapeString(code)+"</code>")
	}
	b.WriteString(escaped)
	return b.String()
}

func replaceCodeSpans(s string, repl func(string) string) string {
	var b strings.Builder
	for {
		open := strings.IndexByte(s, '`')
		if open == -1 {
			b.WriteString(s)
			break
		}
		close := strings.IndexByte(s[open+1:], '`')
		if close == -1 {
			b.WriteString(s)
			break
		}
		close += open + 1
		b.WriteString(s[:open])
		b.WriteString(repl(s[open+1 : close]))
		s = s[close+1:]
	}
	return b.String()
}

func applyLinks(s string) string {
	var b strings.Builder
	for {
		lb := strings.IndexByte(s, '[')
		if lb == -1 {
			b.WriteString(s)
			break
		}
		rb := strings.IndexByte(s[lb:], ']')
		if rb == -1 || lb+rb+1 >= len(s) || s[lb+rb+1] != '(' {
			b.WriteString(s[:lb+1])
			s = s[lb+1:]
			continue
		}
		rb += lb
		rp := strings.IndexByte(s[rb:], ')')
		if rp == -1 {
			b.WriteString(s[:lb+1])
			s = s[lb+1:]
			continue
		}
		rp += rb
		text := s[lb+1 : rb]
		url := s[rb+2 : rp]
		b.WriteString(s[:lb])
		b.WriteString(fmt.Sprintf("<a href=\"%s\">%s</a>", url, text))
		s = s[rp+1:]
	}
	return b.String()
}

func applyEmphasis(s string) string {
	s = wrap(s, "**", "strong")
	s = wrap(s, "__", "strong")
	s = wrap(s, "*", "em")
	return s
}

func wrap(s, marker, tag string) string {
	var b strings.Builder
	open := "<" + tag + ">"
	closeT := "</" + tag + ">"
	for {
		first := strings.Index(s, marker)
		if first == -1 {
			b.WriteString(s)
			break
		}
		second := strings.Index(s[first+len(marker):], marker)
		if second == -1 {
			b.WriteString(s)
			break
		}
		second += first + len(marker)
		b.WriteString(s[:first])
		b.WriteString(open)
		b.WriteString(s[first+len(marker) : second])
		b.WriteString(closeT)
		s = s[second+len(marker):]
	}
	return b.String()
}

func isHeading(line string) bool {
	i := 0
	for i < len(line) && line[i] == '#' {
		i++
	}
	return i > 0 && i <= 6 && i < len(line) && line[i] == ' '
}

func isUnorderedItem(trimmed string) bool {
	return strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ")
}

func isOrderedItem(trimmed string) bool {
	i := 0
	for i < len(trimmed) && trimmed[i] >= '0' && trimmed[i] <= '9' {
		i++
	}
	return i > 0 && i < len(trimmed) && trimmed[i] == '.' && i+1 < len(trimmed) && trimmed[i+1] == ' '
}

func listItemText(trimmed string, ordered bool) string {
	if ordered {
		dot := strings.IndexByte(trimmed, '.')
		return strings.TrimSpace(trimmed[dot+1:])
	}
	return strings.TrimSpace(trimmed[2:])
}

func isTableRow(line string) bool {
	t := strings.TrimSpace(line)
	return strings.HasPrefix(t, "|") && strings.Count(t, "|") >= 2
}

func isTableDivider(line string) bool {
	if !isTableRow(line) {
		return false
	}
	for _, cell := range splitRow(line) {
		c := strings.TrimSpace(cell)
		if c == "" {
			continue
		}
		for _, r := range c {
			if r != '-' && r != ':' {
				return false
			}
		}
	}
	return true
}

func splitRow(line string) []string {
	t := strings.TrimSpace(line)
	t = strings.TrimPrefix(t, "|")
	t = strings.TrimSuffix(t, "|")
	parts := strings.Split(t, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func slug(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	var b strings.Builder
	prevDash := false
	for _, r := range text {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		case r >= 0x80: // keep unicode letters (e.g. Korean) as-is
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

const styleBlock = `<style>
:root { color-scheme: light dark; }
body { margin: 0; font-family: system-ui, -apple-system, "Segoe UI", sans-serif; line-height: 1.6; }
main.doc { max-width: 52rem; margin: 0 auto; padding: 2rem 1.25rem 4rem; }
h1, h2, h3 { line-height: 1.25; }
h1 { border-bottom: 2px solid currentColor; padding-bottom: .3rem; }
h2 { margin-top: 2.5rem; border-bottom: 1px solid rgba(127,127,127,.4); padding-bottom: .2rem; }
code { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; background: rgba(127,127,127,.18); padding: .1rem .35rem; border-radius: .25rem; }
pre { background: rgba(127,127,127,.12); padding: 1rem; border-radius: .5rem; overflow-x: auto; }
pre code { background: none; padding: 0; }
blockquote { margin: 1rem 0; padding: .25rem 1rem; border-left: 4px solid rgba(127,127,127,.5); opacity: .9; }
table { border-collapse: collapse; width: 100%; margin: 1rem 0; }
th, td { border: 1px solid rgba(127,127,127,.4); padding: .5rem .75rem; text-align: left; }
a { color: #2563eb; }
</style>
`
