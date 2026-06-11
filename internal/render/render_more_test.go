package render

import (
	"strings"
	"testing"
)

func mdHTML(t *testing.T, md string) string {
	t.Helper()
	out, err := HTML([]byte("---\ndoc_id: x\ntitle: T\n---\n\n# T\n\n" + md))
	if err != nil {
		t.Fatalf("HTML: %v", err)
	}
	return out
}

func TestRenderBlockquote(t *testing.T) {
	out := mdHTML(t, "> quoted line one\n> quoted line two\n")
	if !strings.Contains(out, "<blockquote>") || !strings.Contains(out, "quoted line one") {
		t.Errorf("blockquote not rendered:\n%s", out)
	}
}

func TestRenderOrderedList(t *testing.T) {
	out := mdHTML(t, "1. first\n2. second\n3. third\n")
	if !strings.Contains(out, "<ol>") || strings.Count(out, "<li>") != 3 {
		t.Errorf("ordered list wrong:\n%s", out)
	}
	if !strings.Contains(out, "<li>first</li>") {
		t.Errorf("ordered item text wrong:\n%s", out)
	}
}

func TestRenderPlainParagraph(t *testing.T) {
	out := mdHTML(t, "이것은 평범한 문단입니다. 두 번째 문장도 있습니다.\n")
	if !strings.Contains(out, "<p>이것은 평범한 문단입니다") {
		t.Errorf("paragraph not rendered:\n%s", out)
	}
}

func TestRenderHorizontalRule(t *testing.T) {
	out := mdHTML(t, "before\n\n---\n\nafter\n")
	if !strings.Contains(out, "<hr>") {
		t.Errorf("hr not rendered:\n%s", out)
	}
}

func TestRenderLinkEdgeCases(t *testing.T) {
	// A bracket with no following parenthesis is left as text.
	out := mdHTML(t, "see [label] and [open](http://x) end\n")
	if !strings.Contains(out, `<a href="http://x">open</a>`) {
		t.Errorf("valid link missing:\n%s", out)
	}
	if strings.Contains(out, "<a href=\"\">label</a>") {
		t.Errorf("non-link bracket should not become anchor:\n%s", out)
	}
}

func TestRenderPipeRowsWithoutDividerAreNotTable(t *testing.T) {
	// Two pipe rows where the second is not a divider must not become a table.
	out := mdHTML(t, "| a | b |\n| c | d |\n")
	if strings.Contains(out, "<table>") {
		t.Errorf("should not render a table without a divider row:\n%s", out)
	}
}

func TestHTMLWithoutFrontmatter(t *testing.T) {
	out, err := HTML([]byte("# Plain\n\nbody text\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<title>Document</title>") {
		t.Errorf("expected default title:\n%s", out[:120])
	}
	if !strings.Contains(out, "<h1") {
		t.Error("expected h1 from body")
	}
}

func TestXMLParseError(t *testing.T) {
	if _, err := XML([]byte("---\ndoc_id: x\nnever closed\n")); err == nil {
		t.Fatal("expected parse error for unterminated frontmatter")
	}
}

func TestHTMLParseError(t *testing.T) {
	if _, err := HTML([]byte("---\ndoc_id: x\nnever closed\n")); err == nil {
		t.Fatal("expected parse error for unterminated frontmatter")
	}
}
