package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func initProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	chdir(t, dir)
	if code, _, errOut := run(t, "init", "."); code != ExitOK {
		t.Fatalf("init failed: code=%d err=%s", code, errOut)
	}
	return dir
}

func TestSubcommandHelp(t *testing.T) {
	for _, cmd := range []string{"init", "generate", "validate", "render", "doctor", "skill"} {
		code, out, _ := run(t, cmd, "-h")
		if code != ExitOK || strings.TrimSpace(out) == "" {
			t.Errorf("%s -h: code=%d out empty=%v", cmd, code, out == "")
		}
	}
	// Top-level --version and -v.
	if code, out, _ := run(t, "-v"); code != ExitOK || !strings.Contains(out, "docs-cli") {
		t.Errorf("-v: code=%d", code)
	}
}

func TestDoctor(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	// No docs yet → reports "not found" but still OK.
	code, out, _ := run(t, "doctor")
	if code != ExitOK {
		t.Fatalf("doctor (no docs): code=%d", code)
	}
	if !strings.Contains(out, "schema: v") {
		t.Errorf("doctor missing schema line:\n%s", out)
	}

	// After init, doctor validates the docs and reports counts.
	if code, _, _ := run(t, "init", "."); code != ExitOK {
		t.Fatal("init failed")
	}
	code, out, _ = run(t, "doctor")
	if code != ExitOK || !strings.Contains(out, "error(s)") {
		t.Errorf("doctor (with docs): code=%d out=%s", code, out)
	}
}

func TestSkillToFile(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	out := filepath.Join("nested", "skill.md")
	if code, _, errOut := run(t, "skill", "--output", out, "--binary", "mytool"); code != ExitOK {
		t.Fatalf("skill: code=%d err=%s", code, errOut)
	}
	data, err := os.ReadFile(filepath.Join(dir, out))
	if err != nil {
		t.Fatalf("skill file not written: %v", err)
	}
	if !strings.Contains(string(data), "mytool init") {
		t.Error("skill did not honor --binary")
	}
}

func TestValidateMissingRequiredExitsOne(t *testing.T) {
	dir := initProject(t)
	if err := os.Remove(filepath.Join(dir, "docs", "overview.md")); err != nil {
		t.Fatal(err)
	}
	if code, _, _ := run(t, "validate", "."); code != ExitValidation {
		t.Errorf("validate with missing required doc: code=%d, want %d", code, ExitValidation)
	}
}

func TestValidateStrictWarningsExitsOne(t *testing.T) {
	dir := initProject(t)
	// data-model is optional; removing it yields a warning, which --strict fails.
	if err := os.Remove(filepath.Join(dir, "docs", "data-model.md")); err != nil {
		t.Fatal(err)
	}
	if code, _, _ := run(t, "validate", ".", "--strict"); code != ExitValidation {
		t.Errorf("strict validate with warning: code=%d, want %d", code, ExitValidation)
	}
	// Without --strict, a warning alone passes.
	if code, _, _ := run(t, "validate", "."); code != ExitOK {
		t.Errorf("non-strict validate with warning: code=%d, want OK", code)
	}
}

func TestRenderOutputDirAndXML(t *testing.T) {
	initProject(t)
	outDir := t.TempDir()
	code, _, errOut := run(t, "render", ".", "--format", "xml", "--output-dir", outDir)
	if code != ExitOK {
		t.Fatalf("render xml: code=%d err=%s", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(outDir, "overview.xml")); err != nil {
		t.Errorf("xml not written to output-dir: %v", err)
	}
}

func TestRenderUnsupportedFormat(t *testing.T) {
	initProject(t)
	if code, _, _ := run(t, "render", ".", "--format", "pdf"); code != ExitUsage {
		t.Errorf("render pdf: code=%d, want usage", code)
	}
}

func TestRenderNoMarkdown(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	if code, _, _ := run(t, "render", ".", "--input-dir", dir); code != ExitUsage {
		t.Errorf("render empty dir: code=%d, want usage", code)
	}
}

func TestGenerateUsageErrors(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	if code, _, _ := run(t, "generate", ".", "--agent", "bogus", "--dry-run"); code != ExitUsage {
		t.Errorf("unknown agent: code=%d, want usage", code)
	}
	if code, _, _ := run(t, "generate", ".", "--doc", "nope", "--dry-run"); code != ExitUsage {
		t.Errorf("unknown doc: code=%d, want usage", code)
	}
}

func TestGenerateAgentNotAvailable(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	t.Setenv("PATH", "") // no agent resolvable
	if code, _, _ := run(t, "generate", ".", "--agent", "claude", "--doc", "overview"); code != ExitEnvironment {
		t.Errorf("unavailable agent: code=%d, want environment", code)
	}
}

func TestGeneratePrintPrompt(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	code, out, _ := run(t, "generate", ".", "--dry-run", "--print-prompt", "--doc", "overview")
	if code != ExitOK {
		t.Fatalf("generate print-prompt: code=%d", code)
	}
	if !strings.Contains(out, "doc_id: overview") {
		t.Errorf("printed prompt missing frontmatter contract:\n%s", out)
	}
}

// fakeRunner is defined in cli_test.go; here we verify the trailing-newline path.
func TestGenerateAppendsTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	fakeBin := filepath.Join(dir, "bin")
	if err := os.MkdirAll(fakeBin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fakeBin, "claude"), []byte("#!/bin/sh\necho ok\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))

	prev := agentRunner
	agentRunner = fakeRunner{content: "---\ndoc_id: overview\ntitle: x\n---\n\n# x\n\n## TL;DR\n\nno trailing newline"}
	t.Cleanup(func() { agentRunner = prev })

	if code, _, errOut := run(t, "generate", ".", "--agent", "claude", "--doc", "overview"); code != ExitOK {
		t.Fatalf("generate: code=%d err=%s", code, errOut)
	}
	data, err := os.ReadFile(filepath.Join(dir, "docs", "overview.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(data), "\n") {
		t.Error("generated doc should end with a newline")
	}
}

func TestFlagParsingForms(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	// --flag=value (string) and repeatable --doc=value.
	if code, _, _ := run(t, "generate", ".", "--output-dir=docs2", "--doc=overview", "--dry-run"); code != ExitOK {
		t.Errorf("--flag=value forms: code=%d", code)
	}
	// bool flag given a value is an error.
	if code, _, _ := run(t, "init", "--force=yes"); code != ExitUsage {
		t.Errorf("bool with value: code=%d, want usage", code)
	}
	// unknown flag.
	if code, _, _ := run(t, "init", "--nope"); code != ExitUsage {
		t.Errorf("unknown flag: code=%d, want usage", code)
	}
	// string flag missing its value.
	if code, _, _ := run(t, "init", "--output-dir"); code != ExitUsage {
		t.Errorf("missing value: code=%d, want usage", code)
	}
}

func TestInitExplicitLangAndSkipExisting(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	// Explicit --lang skips detection (covers the non-auto branch).
	code, out, _ := run(t, "init", ".", "--lang", "rust")
	if code != ExitOK || !strings.Contains(out, "rust") {
		t.Fatalf("init --lang rust: code=%d out=%s", code, out)
	}

	// A second init without --force skips existing files.
	code, out, _ = run(t, "init", ".")
	if code != ExitOK || !strings.Contains(out, "skipped") {
		t.Errorf("init twice should skip existing: code=%d out=%s", code, out)
	}
	_ = filepath.Join(dir, "docs")
}

func TestDoubleDashTerminator(t *testing.T) {
	dir := initProject(t)
	_ = dir
	// "--" forces the following token to be treated as a positional path.
	if code, _, _ := run(t, "validate", "--", "."); code != ExitOK {
		t.Errorf(`validate -- . : code=%d, want OK`, code)
	}
}
