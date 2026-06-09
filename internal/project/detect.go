// Package project inspects a target directory to infer the primary
// implementation language. docs-cli uses this to tailor reverse-engineering
// guidance for AI agents. The test platform supports Python, TypeScript, Go,
// Rust, C#, and Java; those are detected first.
package project

import (
	"os"
	"path/filepath"
	"sort"
)

// Language is a detected implementation language.
type Language string

const (
	Go         Language = "go"
	Python     Language = "python"
	TypeScript Language = "typescript"
	Rust       Language = "rust"
	CSharp     Language = "csharp"
	Java       Language = "java"
	Unknown    Language = "unknown"
)

// Supported lists the languages the test platform targets, in priority order.
var Supported = []Language{Go, Python, TypeScript, Rust, CSharp, Java}

// marker maps a language to filenames whose presence strongly signals it.
var markers = map[Language][]string{
	Go:         {"go.mod"},
	Python:     {"pyproject.toml", "setup.py", "requirements.txt", "Pipfile"},
	TypeScript: {"tsconfig.json", "package.json"},
	Rust:       {"Cargo.toml"},
	CSharp:     {"global.json"},
	Java:       {"pom.xml", "build.gradle", "build.gradle.kts", "settings.gradle"},
}

// extByLang maps a language to source-file extensions, used as a fallback when
// no manifest marker is found.
var extByLang = map[Language][]string{
	Go:         {".go"},
	Python:     {".py"},
	TypeScript: {".ts", ".tsx"},
	Rust:       {".rs"},
	CSharp:     {".cs"},
	Java:       {".java"},
}

// Detection is the outcome of inspecting a directory.
type Detection struct {
	// Primary is the highest-confidence language, or Unknown.
	Primary Language
	// All lists every language for which evidence was found, most evidence first.
	All []Language
}

// Detect inspects dir and infers the implementation language(s). It first looks
// for manifest markers in the directory root, then falls back to counting source
// files one level deep. It never returns an error for a readable directory; an
// empty or unreadable directory yields Unknown.
func Detect(dir string) Detection {
	if dir == "" {
		dir = "."
	}

	// 1) Manifest markers in the root carry the most weight.
	for _, lang := range Supported {
		for _, marker := range markers[lang] {
			if fileExists(filepath.Join(dir, marker)) {
				// A C# project is identified by .csproj/.sln anywhere near root.
				return Detection{Primary: lang, All: markerLanguages(dir)}
			}
		}
	}
	if hasGlob(dir, "*.csproj") || hasGlob(dir, "*.sln") {
		return Detection{Primary: CSharp, All: []Language{CSharp}}
	}

	// 2) Fall back to source-file extension counts.
	counts := extensionCounts(dir)
	if len(counts) == 0 {
		return Detection{Primary: Unknown}
	}
	type pair struct {
		lang  Language
		count int
	}
	pairs := make([]pair, 0, len(counts))
	for lang, c := range counts {
		pairs = append(pairs, pair{lang, c})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count != pairs[j].count {
			return pairs[i].count > pairs[j].count
		}
		return pairs[i].lang < pairs[j].lang
	})
	all := make([]Language, 0, len(pairs))
	for _, p := range pairs {
		all = append(all, p.lang)
	}
	return Detection{Primary: pairs[0].lang, All: all}
}

func markerLanguages(dir string) []Language {
	var out []Language
	for _, lang := range Supported {
		for _, marker := range markers[lang] {
			if fileExists(filepath.Join(dir, marker)) {
				out = append(out, lang)
				break
			}
		}
	}
	if hasGlob(dir, "*.csproj") || hasGlob(dir, "*.sln") {
		out = append(out, CSharp)
	}
	return dedupe(out)
}

func extensionCounts(dir string) map[Language]int {
	counts := map[Language]int{}
	extLang := map[string]Language{}
	for lang, exts := range extByLang {
		for _, ext := range exts {
			extLang[ext] = lang
		}
	}
	// Walk at most two levels deep to stay fast on large repos.
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if path != dir && (name == ".git" || name == "node_modules" || name == "vendor" || name == "target" || name == "dist" || name == "bin") {
				return filepath.SkipDir
			}
			if depth(dir, path) > 2 {
				return filepath.SkipDir
			}
			return nil
		}
		if lang, ok := extLang[filepath.Ext(path)]; ok {
			counts[lang]++
		}
		return nil
	})
	return counts
}

// depth returns how many directory levels path is below root.
func depth(root, path string) int {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." {
		return 0
	}
	n := 1
	for _, r := range rel {
		if r == filepath.Separator {
			n++
		}
	}
	return n
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func hasGlob(dir, pattern string) bool {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	return err == nil && len(matches) > 0
}

func dedupe(langs []Language) []Language {
	seen := map[Language]struct{}{}
	var out []Language
	for _, l := range langs {
		if _, ok := seen[l]; ok {
			continue
		}
		seen[l] = struct{}{}
		out = append(out, l)
	}
	return out
}
