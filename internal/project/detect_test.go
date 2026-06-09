package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectByMarker(t *testing.T) {
	cases := map[string]Language{
		"go.mod":         Go,
		"Cargo.toml":     Rust,
		"pyproject.toml": Python,
		"tsconfig.json":  TypeScript,
		"pom.xml":        Java,
	}
	for marker, want := range cases {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, marker), "x")
		if got := Detect(dir).Primary; got != want {
			t.Errorf("marker %s: got %q, want %q", marker, got, want)
		}
	}
}

func TestDetectCSharpByProject(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "App.csproj"), "<Project/>")
	if got := Detect(dir).Primary; got != CSharp {
		t.Errorf("got %q, want csharp", got)
	}
}

func TestDetectByExtensionFallback(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.rs"), "fn main(){}")
	writeFile(t, filepath.Join(dir, "b.rs"), "fn x(){}")
	writeFile(t, filepath.Join(dir, "c.py"), "print()")
	if got := Detect(dir).Primary; got != Rust {
		t.Errorf("got %q, want rust (more .rs files)", got)
	}
}

func TestDetectUnknown(t *testing.T) {
	if got := Detect(t.TempDir()).Primary; got != Unknown {
		t.Errorf("got %q, want unknown", got)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
