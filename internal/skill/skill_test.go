package skill

import (
	"strings"
	"testing"

	"github.com/jhl-labs/docs-cli/internal/schema"
)

func TestGenerateHasFrontmatterAndDocs(t *testing.T) {
	out := Generate(schema.Standard(), Options{})
	if !strings.HasPrefix(out, "---\nname: standardizing-project-docs\n") {
		t.Errorf("missing or wrong frontmatter head:\n%s", out[:80])
	}
	if !strings.Contains(out, "description: Use when") {
		t.Error("description must start with 'Use when'")
	}
	// Every schema doc id should appear in the document table.
	for _, d := range schema.Standard().Docs {
		if !strings.Contains(out, "`"+d.ID+"`") {
			t.Errorf("skill missing doc id %q", d.ID)
		}
	}
}

func TestGenerateRespectsBinaryName(t *testing.T) {
	out := Generate(schema.Standard(), Options{Binary: "mytool"})
	if !strings.Contains(out, "mytool init") {
		t.Error("custom binary name not used in examples")
	}
}
