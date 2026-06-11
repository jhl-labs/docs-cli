package agent

import (
	"context"
	"strings"
	"testing"
)

func TestExecRunnerSuccess(t *testing.T) {
	out, err := ExecRunner{}.Run(context.Background(), t.TempDir(), "echo", []string{"hello world"})
	if err != nil {
		t.Fatalf("Run echo: %v", err)
	}
	if strings.TrimSpace(out) != "hello world" {
		t.Errorf("out = %q", out)
	}
}

func TestExecRunnerFailureCapturesStderr(t *testing.T) {
	_, err := ExecRunner{}.Run(context.Background(), t.TempDir(), "sh", []string{"-c", "echo boom >&2; exit 3"})
	if err == nil {
		t.Fatal("expected error from failing command")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("error should include stderr, got %v", err)
	}
}

func TestExecRunnerMissingBinary(t *testing.T) {
	_, err := ExecRunner{}.Run(context.Background(), t.TempDir(), "docs-cli-no-such-binary-xyz", nil)
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
}

func TestAvailable(t *testing.T) {
	if Available("definitely-not-an-agent") {
		t.Error("unknown agent must not be available")
	}
	// Exercises the PATH lookup path; result depends on the host, so we only
	// require that it returns without panicking for a known agent.
	_ = Available("claude")
}

func TestDefaultStr(t *testing.T) {
	if got := defaultStr("", "fallback"); got != "fallback" {
		t.Errorf("defaultStr empty = %q", got)
	}
	if got := defaultStr("  ", "fallback"); got != "fallback" {
		t.Errorf("defaultStr blank = %q", got)
	}
	if got := defaultStr("set", "fallback"); got != "set" {
		t.Errorf("defaultStr set = %q", got)
	}
}
