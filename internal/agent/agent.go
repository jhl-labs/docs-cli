// Package agent drives external AI coding agents (Claude Code, Codex, Gemini
// CLI, opencode) to reverse-engineer a project and fill standardized documents.
// It builds a per-document prompt from the schema and runs the selected agent in
// non-interactive mode. The runner is an interface so the orchestration is
// testable without invoking a real agent.
package agent

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/jhl-labs/docs-cli/internal/schema"
)

// Agent describes how to invoke a non-interactive AI CLI.
type Agent struct {
	// Name is the selector used on the command line.
	Name string
	// Bin is the executable to look up on PATH.
	Bin string
	// args are the fixed arguments that precede the prompt.
	args []string
}

// Command returns the binary and the full argument list (fixed args + prompt).
func (a Agent) Command(prompt string) (string, []string) {
	return a.Bin, append(append([]string{}, a.args...), prompt)
}

var registry = map[string]Agent{
	"claude":   {Name: "claude", Bin: "claude", args: []string{"-p"}},
	"codex":    {Name: "codex", Bin: "codex", args: []string{"exec"}},
	"gemini":   {Name: "gemini", Bin: "gemini", args: []string{"-p"}},
	"opencode": {Name: "opencode", Bin: "opencode", args: []string{"run"}},
}

// Known returns the supported agent selectors, sorted.
func Known() []string {
	out := make([]string, 0, len(registry))
	for name := range registry {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// Lookup returns the agent for a selector.
func Lookup(name string) (Agent, bool) {
	a, ok := registry[strings.ToLower(strings.TrimSpace(name))]
	return a, ok
}

// Available reports whether the agent's binary is on PATH.
func Available(name string) bool {
	a, ok := Lookup(name)
	if !ok {
		return false
	}
	_, err := exec.LookPath(a.Bin)
	return err == nil
}

// Runner executes a command with a working directory and returns combined
// stdout. Implementations may capture stderr separately.
type Runner interface {
	Run(ctx context.Context, dir, bin string, args []string) (string, error)
}

// ExecRunner runs commands with os/exec.
type ExecRunner struct{}

// Run implements Runner.
func (ExecRunner) Run(ctx context.Context, dir, bin string, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if ok := asExitError(err, &exitErr); ok && len(exitErr.Stderr) > 0 {
			return string(out), fmt.Errorf("%w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return string(out), err
	}
	return string(out), nil
}

func asExitError(err error, target **exec.ExitError) bool {
	if e, ok := err.(*exec.ExitError); ok {
		*target = e
		return true
	}
	return false
}

// PromptOptions parameterize prompt construction.
type PromptOptions struct {
	// ProjectDir is the path the agent should analyze.
	ProjectDir string
	// Language is the detected primary language (may be empty/unknown).
	Language string
	// OutputDir is where the filled Markdown file should be written.
	OutputDir string
	// SchemaVersion is recorded in the prompt's frontmatter contract.
	SchemaVersion string
}

// BuildPrompt produces the reverse-engineering instruction for one document.
// The prompt tells the agent exactly which chapters to write and what
// frontmatter to emit, so the output stays conformant to the schema.
func BuildPrompt(doc schema.Document, opts PromptOptions) string {
	var b strings.Builder
	b.WriteString("당신은 시니어 소프트웨어 아키텍트입니다. 아래 프로젝트를 리버스 엔지니어링하여 ")
	b.WriteString("표준 문서 한 편을 작성합니다.\n\n")
	b.WriteString(fmt.Sprintf("분석 대상 경로: %s\n", defaultStr(opts.ProjectDir, ".")))
	if opts.Language != "" && opts.Language != "unknown" {
		b.WriteString(fmt.Sprintf("주 구현 언어: %s\n", opts.Language))
	}
	b.WriteString(fmt.Sprintf("작성할 문서: %s — %s\n", doc.Title, doc.Purpose))
	b.WriteString(fmt.Sprintf("저장 위치: %s/%s\n\n", defaultStr(opts.OutputDir, "docs"), doc.FileName()))

	b.WriteString("## 작성 규칙 (반드시 준수)\n\n")
	b.WriteString("1. 실제 소스 코드·설정·CI를 읽고 사실에 근거해 작성합니다. 추측은 명시적으로 표시합니다.\n")
	b.WriteString("2. 파일 맨 위에 아래 프론트매터를 그대로 둡니다(값만 채움):\n\n")
	b.WriteString("```yaml\n---\n")
	b.WriteString(fmt.Sprintf("doc_id: %s\n", doc.ID))
	b.WriteString(fmt.Sprintf("title: %s\n", doc.Title))
	b.WriteString(fmt.Sprintf("section: %s\n", doc.Section))
	b.WriteString(fmt.Sprintf("order: %d\n", doc.Order))
	b.WriteString(fmt.Sprintf("audience: [%s]\n", strings.Join(doc.Audience, ", ")))
	b.WriteString("status: generated\n")
	b.WriteString(fmt.Sprintf("schema_version: %s\n", defaultStr(opts.SchemaVersion, schema.SchemaVersion)))
	b.WriteString("generated_by: <agent-name>\n")
	b.WriteString("source_commit: <git-sha>\n")
	b.WriteString("updated: <YYYY-MM-DD>\n")
	b.WriteString("---\n```\n\n")
	b.WriteString(fmt.Sprintf("3. H1 제목은 `# %s` 입니다.\n", doc.Title))
	b.WriteString("4. 아래 H2 챕터를 **정확히 이 제목과 순서로** 모두 포함합니다. 빈 챕터를 남기지 마세요.\n\n")

	for i, ch := range doc.Chapters {
		b.WriteString(fmt.Sprintf("   %d. `## %s` — %s\n", i+1, ch.Heading, ch.Guidance))
	}
	b.WriteString("\n5. 다이어그램은 Mermaid 코드블록(```mermaid)으로, 표는 GitHub 표 문법으로 작성합니다.\n")
	b.WriteString("6. 한국어로 작성합니다. 출력은 완성된 Markdown 문서 본문만 포함합니다(설명 문구 금지).\n")
	return b.String()
}

func defaultStr(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
