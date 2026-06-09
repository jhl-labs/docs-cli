package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jhl-labs/docs-cli/internal/mddoc"
	"github.com/jhl-labs/docs-cli/internal/schema"
)

func TestBuildProducesEveryDocument(t *testing.T) {
	s := schema.Standard()
	files := Build(s, Options{OutputDir: "docs", Date: "2026-06-09"})

	got := map[string]string{}
	for _, f := range files {
		got[f.Path] = f.Content
	}
	if _, ok := got[filepath.Join("docs", "README.md")]; !ok {
		t.Fatal("missing index README.md")
	}
	for _, d := range s.Docs {
		path := filepath.Join("docs", d.FileName())
		content, ok := got[path]
		if !ok {
			t.Errorf("missing %s", path)
			continue
		}
		// Every scaffolded document must round-trip with valid frontmatter
		// and contain each schema chapter heading.
		doc, hasFM, err := mddoc.Parse([]byte(content))
		if err != nil || !hasFM {
			t.Errorf("%s: bad frontmatter (err=%v, hasFM=%v)", path, err, hasFM)
			continue
		}
		if doc.Frontmatter.DocID != d.ID {
			t.Errorf("%s: doc_id = %q", path, doc.Frontmatter.DocID)
		}
		h2 := map[string]bool{}
		for _, h := range doc.H2() {
			h2[h] = true
		}
		for _, ch := range d.Chapters {
			if !h2[ch.Heading] {
				t.Errorf("%s: missing chapter %q", path, ch.Heading)
			}
		}
	}
}

func TestWriteSkipsExistingUnlessForce(t *testing.T) {
	dir := t.TempDir()
	files := []File{{Path: filepath.Join(dir, "a.md"), Content: "first"}}

	written, skipped, err := Write(files, false)
	if err != nil || len(written) != 1 || len(skipped) != 0 {
		t.Fatalf("first write: written=%v skipped=%v err=%v", written, skipped, err)
	}

	files[0].Content = "second"
	written, skipped, err = Write(files, false)
	if err != nil || len(written) != 0 || len(skipped) != 1 {
		t.Fatalf("skip write: written=%v skipped=%v err=%v", written, skipped, err)
	}
	data, _ := os.ReadFile(files[0].Path)
	if string(data) != "first" {
		t.Errorf("content overwritten without force: %q", data)
	}

	written, _, err = Write(files, true)
	if err != nil || len(written) != 1 {
		t.Fatalf("force write: written=%v err=%v", written, err)
	}
	data, _ = os.ReadFile(files[0].Path)
	if string(data) != "second" {
		t.Errorf("force did not overwrite: %q", data)
	}
}

func TestIndexMentionsSchemaVersion(t *testing.T) {
	s := schema.Standard()
	files := Build(s, Options{OutputDir: "docs", Date: "2026-06-09"})
	if !strings.Contains(files[0].Content, "스키마 v"+s.Version) {
		t.Errorf("index missing schema version marker")
	}
}
