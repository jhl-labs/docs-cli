package version

import "testing"

func TestString(t *testing.T) {
	defer func(v, c string) { Version, Commit = v, c }(Version, Commit)

	Version, Commit = "v1.2.3", "abc1234"
	if got := String(); got != "v1.2.3 (abc1234)" {
		t.Errorf("String() = %q", got)
	}

	Version, Commit = "v1.2.3", "unknown"
	if got := String(); got != "v1.2.3" {
		t.Errorf("String() with unknown commit = %q", got)
	}

	Version, Commit = "dev", ""
	if got := String(); got != "dev" {
		t.Errorf("String() with empty commit = %q", got)
	}
}
