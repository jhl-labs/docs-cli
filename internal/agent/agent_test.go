package agent

import (
	"strings"
	"testing"

	"github.com/jhl-labs/docs-cli/internal/schema"
)

func TestKnownAndLookup(t *testing.T) {
	known := Known()
	if len(known) != 4 {
		t.Fatalf("known = %v", known)
	}
	for _, name := range known {
		if _, ok := Lookup(name); !ok {
			t.Errorf("Lookup(%q) failed", name)
		}
	}
	if _, ok := Lookup("nope"); ok {
		t.Error("Lookup of unknown agent succeeded")
	}
}

func TestCommandAppendsPrompt(t *testing.T) {
	a, _ := Lookup("claude")
	bin, args := a.Command("hello")
	if bin != "claude" {
		t.Errorf("bin = %q", bin)
	}
	if len(args) != 2 || args[0] != "-p" || args[1] != "hello" {
		t.Errorf("args = %v", args)
	}

	c, _ := Lookup("codex")
	_, cargs := c.Command("p")
	if cargs[0] != "exec" || cargs[len(cargs)-1] != "p" {
		t.Errorf("codex args = %v", cargs)
	}
}

func TestBuildPromptContainsChaptersAndFrontmatter(t *testing.T) {
	doc, _ := schema.Standard().Doc("architecture")
	prompt := BuildPrompt(doc, PromptOptions{ProjectDir: "/repo", Language: "go", OutputDir: "docs", SchemaVersion: "1"})

	if !strings.Contains(prompt, "doc_id: architecture") {
		t.Error("prompt missing frontmatter doc_id")
	}
	if !strings.Contains(prompt, "docs/architecture.md") {
		t.Error("prompt missing output path")
	}
	if !strings.Contains(prompt, "주 구현 언어: go") {
		t.Error("prompt missing language")
	}
	for _, ch := range doc.Chapters {
		if !strings.Contains(prompt, ch.Heading) {
			t.Errorf("prompt missing chapter %q", ch.Heading)
		}
	}
}
