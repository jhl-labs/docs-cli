package schema

import "testing"

func TestSectionByID(t *testing.T) {
	s := Standard()
	if sec, ok := s.SectionByID("architecture"); !ok || sec.ID != "architecture" {
		t.Errorf("SectionByID(architecture) = %+v, %v", sec, ok)
	}
	if _, ok := s.SectionByID("nope"); ok {
		t.Error("SectionByID(nope) should not be found")
	}
}

func TestDocNotFound(t *testing.T) {
	s := Standard()
	if _, ok := s.Doc("does-not-exist"); ok {
		t.Error("Doc(does-not-exist) should not be found")
	}
	if _, ok := s.Doc("overview"); !ok {
		t.Error("Doc(overview) should be found")
	}
}

func TestDocsInSectionEmpty(t *testing.T) {
	if got := Standard().DocsInSection("nope"); got != nil {
		t.Errorf("DocsInSection(nope) = %v, want nil", got)
	}
}
