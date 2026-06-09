package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func run(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var out, errBuf bytes.Buffer
	code := Run(args, &out, &errBuf)
	return code, out.String(), errBuf.String()
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
}

func TestVersionAndHelp(t *testing.T) {
	code, out, _ := run(t, "version")
	if code != ExitOK || !strings.Contains(out, "docs-cli") {
		t.Errorf("version: code=%d out=%q", code, out)
	}
	code, out, _ = run(t)
	if code != ExitOK || !strings.Contains(out, "docs-cli") {
		t.Errorf("help: code=%d", code)
	}
	code, _, _ = run(t, "bogus")
	if code != ExitUsage {
		t.Errorf("unknown command: code=%d", code)
	}
}

func TestInitValidateRenderSkillFlow(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	// init
	if code, _, errOut := run(t, "init", "."); code != ExitOK {
		t.Fatalf("init failed: code=%d err=%s", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, "docs", "README.md")); err != nil {
		t.Fatalf("index not created: %v", err)
	}

	// validate (fresh scaffold => 0 errors => ExitOK)
	if code, out, _ := run(t, "validate", "."); code != ExitOK {
		t.Fatalf("validate: code=%d\n%s", code, out)
	}

	// render html + xml
	code, _, errOut := run(t, "render", ".", "--format", "html", "--format", "xml")
	if code != ExitOK {
		t.Fatalf("render: code=%d err=%s", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, "docs", "_site", "overview.html")); err != nil {
		t.Errorf("html not rendered: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "docs", "_xml", "overview.xml")); err != nil {
		t.Errorf("xml not rendered: %v", err)
	}

	// skill to stdout
	if code, out, _ := run(t, "skill", "--output", "-"); code != ExitOK || !strings.Contains(out, "name: standardizing-project-docs") {
		t.Errorf("skill: code=%d", code)
	}
}

func TestGenerateDryRunWritesPrompts(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	code, _, errOut := run(t, "generate", ".", "--dry-run", "--doc", "overview", "--doc", "architecture")
	if code != ExitOK {
		t.Fatalf("generate dry-run: code=%d err=%s", code, errOut)
	}
	for _, id := range []string{"overview", "architecture"} {
		p := filepath.Join(dir, ".docs-cli", "prompts", id+".prompt.md")
		if _, err := os.Stat(p); err != nil {
			t.Errorf("prompt %s missing: %v", id, err)
		}
	}
}

// fakeRunner returns canned document content instead of invoking a real agent.
type fakeRunner struct{ content string }

func (f fakeRunner) Run(_ context.Context, _, _ string, _ []string) (string, error) {
	return f.content, nil
}

func TestGenerateWithFakeAgentWritesDocs(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	// Swap the runner and force the agent to look "available" by selecting one
	// and using --dry-run=false; availability is checked via PATH, so we bypass
	// it by pre-seeding a fake binary on PATH.
	fakeBin := filepath.Join(dir, "bin")
	if err := os.MkdirAll(fakeBin, 0o755); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\necho ok\n"
	if err := os.WriteFile(filepath.Join(fakeBin, "claude"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))

	prev := agentRunner
	agentRunner = fakeRunner{content: cannedOverview}
	t.Cleanup(func() { agentRunner = prev })

	code, out, errOut := run(t, "generate", ".", "--agent", "claude", "--doc", "overview")
	if code != ExitOK {
		t.Fatalf("generate: code=%d\nout=%s\nerr=%s", code, out, errOut)
	}
	data, err := os.ReadFile(filepath.Join(dir, "docs", "overview.md"))
	if err != nil {
		t.Fatalf("doc not written: %v", err)
	}
	if !strings.Contains(string(data), "doc_id: overview") {
		t.Errorf("doc content unexpected:\n%s", data)
	}
}

const cannedOverview = `---
doc_id: overview
title: 프로젝트 개요
section: overview
order: 1
audience: [newcomer]
status: generated
schema_version: 1
generated_by: claude
source_commit: abc
updated: 2026-06-09
---

# 프로젝트 개요

## TL;DR

- generated
`
