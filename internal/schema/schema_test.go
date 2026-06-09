package schema

import "testing"

func TestStandardIsWellFormed(t *testing.T) {
	s := Standard()
	if s.Version != SchemaVersion {
		t.Fatalf("version = %q, want %q", s.Version, SchemaVersion)
	}
	if len(s.Docs) == 0 {
		t.Fatal("no documents")
	}

	sectionIDs := map[string]bool{}
	for _, sec := range s.Sections {
		sectionIDs[sec.ID] = true
	}

	seen := map[string]bool{}
	for _, d := range s.Docs {
		if d.ID == "" {
			t.Errorf("document with empty id")
		}
		if seen[d.ID] {
			t.Errorf("duplicate doc id %q", d.ID)
		}
		seen[d.ID] = true
		if !sectionIDs[d.Section] {
			t.Errorf("doc %q references unknown section %q", d.ID, d.Section)
		}
		if d.Title == "" {
			t.Errorf("doc %q has empty title", d.ID)
		}
		if len(d.Chapters) == 0 {
			t.Errorf("doc %q has no chapters", d.ID)
		}
		for _, ch := range d.Chapters {
			if ch.Heading == "" || ch.Guidance == "" {
				t.Errorf("doc %q has an incomplete chapter", d.ID)
			}
		}
	}
}

func TestFileName(t *testing.T) {
	cases := map[string]string{
		"overview": "overview.md",
		"adr":      "adr/README.md",
	}
	s := Standard()
	for id, want := range cases {
		doc, ok := s.Doc(id)
		if !ok {
			t.Fatalf("doc %q not found", id)
		}
		if got := doc.FileName(); got != want {
			t.Errorf("%s FileName = %q, want %q", id, got, want)
		}
	}
}

func TestDocsInSectionOrdered(t *testing.T) {
	s := Standard()
	docs := s.DocsInSection("overview")
	if len(docs) < 2 {
		t.Fatalf("expected at least 2 overview docs, got %d", len(docs))
	}
}
