package mddoc

import (
	"strings"
	"testing"
)

func TestParseRoundTrip(t *testing.T) {
	fm := Frontmatter{
		DocID:         "overview",
		Title:         "프로젝트 개요",
		Section:       "overview",
		Order:         1,
		Audience:      []string{"newcomer", "decision-maker"},
		Status:        "draft",
		SchemaVersion: "1",
		GeneratedBy:   "docs-cli template",
		SourceCommit:  "abc1234",
		Updated:       "2026-06-09",
	}
	raw := fm.Marshal() + "\n# 프로젝트 개요\n\n## TL;DR\n\n- one\n\n## 문제 정의\n\nbody\n"

	doc, hasFM, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !hasFM {
		t.Fatal("expected frontmatter")
	}
	if doc.Frontmatter.DocID != "overview" || doc.Frontmatter.Order != 1 {
		t.Errorf("frontmatter mismatch: %+v", doc.Frontmatter)
	}
	if len(doc.Frontmatter.Audience) != 2 {
		t.Errorf("audience = %v", doc.Frontmatter.Audience)
	}
	h2 := doc.H2()
	if len(h2) != 2 || h2[0] != "TL;DR" || h2[1] != "문제 정의" {
		t.Errorf("H2 = %v", h2)
	}
}

func TestParseNoFrontmatter(t *testing.T) {
	doc, hasFM, err := Parse([]byte("# Title\n\ntext\n"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if hasFM {
		t.Error("did not expect frontmatter")
	}
	if len(doc.Headings) != 1 || doc.Headings[0].Level != 1 {
		t.Errorf("headings = %+v", doc.Headings)
	}
}

func TestParseUnterminatedFrontmatter(t *testing.T) {
	_, _, err := Parse([]byte("---\ntitle: x\n# no close\n"))
	if err == nil {
		t.Fatal("expected error for unterminated frontmatter")
	}
}

func TestMarshalStableOrder(t *testing.T) {
	out := Frontmatter{DocID: "x", Order: 2}.Marshal()
	if !strings.HasPrefix(out, "---\ndoc_id: x\n") {
		t.Errorf("unexpected marshal head: %q", out)
	}
	if !strings.Contains(out, "order: 2\n") {
		t.Errorf("missing order: %q", out)
	}
}
