package render

import (
	"strings"
	"testing"
)

const sample = `---
doc_id: overview
title: 프로젝트 개요
section: overview
order: 1
audience: [newcomer]
status: generated
schema_version: 1
generated_by: test
source_commit: abc
updated: 2026-06-09
---

# 프로젝트 개요

## TL;DR

- **bold** item with ` + "`code`" + `
- a [link](https://example.com)

## 표

| A | B |
| --- | --- |
| 1 | 2 |

` + "```go\nfmt.Println(\"hi\")\n```" + `
`

func TestHTMLRendersStructure(t *testing.T) {
	out, err := HTML([]byte(sample))
	if err != nil {
		t.Fatalf("HTML: %v", err)
	}
	checks := []string{
		"<title>프로젝트 개요</title>",
		"<meta name=\"doc-id\" content=\"overview\">",
		"<h1",
		"<h2",
		"<strong>bold</strong>",
		"<code>code</code>",
		`<a href="https://example.com">link</a>`,
		"<table>",
		"<th>A</th>",
		"<td>1</td>",
		`<pre><code class="language-go">`,
	}
	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Errorf("HTML missing %q", c)
		}
	}
	// Frontmatter must not leak into the body.
	if strings.Contains(out, "doc_id: overview") {
		t.Error("frontmatter leaked into HTML body")
	}
}

func TestHTMLEscapesCodeContent(t *testing.T) {
	out, err := HTML([]byte("# x\n\n```\n<script>alert(1)</script>\n```\n"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "<script>alert(1)</script>") {
		t.Error("code content not escaped")
	}
	if !strings.Contains(out, "&lt;script&gt;") {
		t.Error("expected escaped code")
	}
}

func TestXMLStructure(t *testing.T) {
	out, err := XML([]byte(sample))
	if err != nil {
		t.Fatalf("XML: %v", err)
	}
	checks := []string{
		`doc-id="overview"`,
		`section="overview"`,
		`schema-version="1"`,
		"<title>프로젝트 개요</title>",
		"<role>newcomer</role>",
		`<chapter id="tl-dr">`,
		"<heading>TL;DR</heading>",
	}
	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Errorf("XML missing %q\n---\n%s", c, out)
		}
	}
}

func TestSlugKeepsUnicode(t *testing.T) {
	if got := slug("문제 정의"); got != "문제-정의" {
		t.Errorf("slug = %q", got)
	}
	if got := slug("TL;DR"); got != "tl-dr" {
		t.Errorf("slug = %q", got)
	}
}
