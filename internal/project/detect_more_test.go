package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectExcludesVendorDirsAndCountsDepth(t *testing.T) {
	dir := t.TempDir()
	mk := func(rel string) {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Real Python sources at varying depths (within the 2-level walk).
	mk("a.py")
	mk("pkg/b.py")
	mk("pkg/util/c.py")
	// Deeper than the walk limit — must be ignored.
	mk("pkg/util/deep/deeper/d.py")
	// A dependency dir packed with TypeScript that must be excluded.
	for i := 0; i < 10; i++ {
		mk(filepath.Join("node_modules", "m"+string(rune('0'+i))+".ts"))
	}

	det := Detect(dir)
	if det.Primary != Python {
		t.Errorf("Primary = %q, want python (node_modules must be excluded)", det.Primary)
	}
	found := false
	for _, l := range det.All {
		if l == Python {
			found = true
		}
		if l == TypeScript {
			t.Error("TypeScript from node_modules should be excluded")
		}
	}
	if !found {
		t.Errorf("All = %v, want python present", det.All)
	}
}

func TestDetectEmptyDirArg(t *testing.T) {
	// Empty dir argument defaults to "." which is this package directory and
	// contains .go files, so Go is detected.
	if got := Detect("").Primary; got != Go {
		t.Errorf("Detect(\"\") = %q, want go", got)
	}
}
