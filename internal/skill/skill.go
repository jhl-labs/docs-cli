// Package skill generates a SKILL.md describing how an AI agent should use
// docs-cli to scaffold, fill, validate, and render the standardized docs set.
// The output follows the superpowers skill convention: YAML frontmatter with
// name + description (triggering conditions only), then a concise body.
package skill

import (
	"fmt"
	"strings"

	"github.com/jhl-labs/docs-cli/internal/schema"
)

// Options parameterize skill generation.
type Options struct {
	// Name is the skill slug (letters, numbers, hyphens).
	Name string
	// Binary is the CLI invocation name shown in examples.
	Binary string
}

// Generate returns a SKILL.md body for the given schema and options.
func Generate(s schema.Schema, opts Options) string {
	name := opts.Name
	if name == "" {
		name = "standardizing-project-docs"
	}
	bin := opts.Binary
	if bin == "" {
		bin = "docs-cli"
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: " + name + "\n")
	b.WriteString("description: Use when a project needs a standardized docs/ folder, when reverse-engineering an existing codebase into architecture docs, or when generating docs that a web service will ingest\n")
	b.WriteString("---\n\n")

	b.WriteString("# Standardizing Project Docs\n\n")
	b.WriteString("## Overview\n\n")
	b.WriteString("`" + bin + "` produces a **fixed, web-portable documentation set** from one schema. ")
	b.WriteString("Every document carries the same frontmatter keys and the same ordered chapters, so the output is uniform across projects and ingestible by a downstream service.\n\n")
	b.WriteString("**Core principle:** The schema is the single source of truth. Never invent documents or chapters — scaffold from the schema, fill the chapters, then validate.\n\n")

	b.WriteString("## When to Use\n\n")
	b.WriteString("- A repository has no docs/, or has ad-hoc docs that need standardizing.\n")
	b.WriteString("- You must reverse-engineer an unfamiliar codebase into architecture docs.\n")
	b.WriteString("- Generated docs must be consumed by a web service (consistent structure required).\n")
	b.WriteString("- Supported project languages: Python, TypeScript, Go, Rust, C#, Java.\n\n")
	b.WriteString("**Do NOT use for:** one-off READMEs, or free-form design notes that need no structure.\n\n")

	b.WriteString("## The Iron Law\n\n")
	b.WriteString("```\nNO DOCUMENT MARKED reviewed WITHOUT PASSING `" + bin + " validate`\n```\n\n")
	b.WriteString("Filled a doc? Run validate. Errors > 0? Fix before committing. ")
	b.WriteString("Leaving `TODO(docs-cli)` placeholders in a `reviewed` doc is a validation failure, not a style nit.\n\n")

	b.WriteString("## Workflow\n\n")
	b.WriteString("```bash\n")
	b.WriteString(bin + " init .                 # 1. scaffold the standardized tree into docs/\n")
	b.WriteString(bin + " generate . --agent claude   # 2. let an agent fill each chapter from the code\n")
	b.WriteString(bin + " validate .             # 3. gate: required docs, frontmatter, chapters\n")
	b.WriteString(bin + " render . --format html --format xml   # 4. emit web-portable artifacts\n")
	b.WriteString("```\n\n")
	b.WriteString("Without an agent, `generate --dry-run` writes one prompt file per document under `.docs-cli/prompts/`; hand those to any agent and write the result back to the matching `docs/<id>.md`.\n\n")

	b.WriteString("## Quick Reference\n\n")
	b.WriteString("| Command | Purpose |\n| --- | --- |\n")
	b.WriteString("| `" + bin + " init [path]` | Scaffold docs/ from the schema |\n")
	b.WriteString("| `" + bin + " generate [path]` | Build prompts / run an agent to fill docs |\n")
	b.WriteString("| `" + bin + " validate [path]` | Check conformance to the schema |\n")
	b.WriteString("| `" + bin + " render [path]` | Convert docs to HTML/XML |\n")
	b.WriteString("| `" + bin + " doctor` | Check environment and available agents |\n")
	b.WriteString("| `" + bin + " skill` | Regenerate this skill |\n\n")

	b.WriteString("## The Standardized Document Set (schema v" + s.Version + ")\n\n")
	b.WriteString("Fill every chapter; do not add or drop documents.\n\n")
	b.WriteString("| Document | Section | Chapters |\n| --- | --- | --- |\n")
	for _, doc := range s.Docs {
		headings := make([]string, 0, len(doc.Chapters))
		for _, ch := range doc.Chapters {
			headings = append(headings, ch.Heading)
		}
		req := ""
		if doc.Required {
			req = " *(required)*"
		}
		b.WriteString(fmt.Sprintf("| `%s`%s | %s | %s |\n", doc.ID, req, doc.Section, strings.Join(headings, " · ")))
	}
	b.WriteString("\n")

	b.WriteString("## Frontmatter Contract\n\n")
	b.WriteString("Every document begins with this block (values filled, keys unchanged):\n\n")
	b.WriteString("```yaml\n---\ndoc_id: <id>\ntitle: <title>\nsection: <section>\norder: <n>\naudience: [<roles>]\nstatus: draft | generated | reviewed\nschema_version: " + s.Version + "\ngenerated_by: <template | agent name>\nsource_commit: <git-sha>\nupdated: <YYYY-MM-DD>\n---\n```\n\n")

	b.WriteString("## Common Mistakes\n\n")
	b.WriteString("| Mistake | Fix |\n| --- | --- |\n")
	b.WriteString("| Renaming or reordering chapters | Keep the schema's headings exactly; validate enforces them |\n")
	b.WriteString("| Leaving `TODO(docs-cli)` in a reviewed doc | Fill the chapter, or keep status at `generated` |\n")
	b.WriteString("| Inventing extra documents | The web service only knows schema doc_ids; add via a schema change |\n")
	b.WriteString("| Marking `reviewed` before running validate | Run `" + bin + " validate` first; 0 errors required |\n\n")

	b.WriteString("## Red Flags - STOP\n\n")
	b.WriteString("- Writing prose without reading the actual source first\n")
	b.WriteString("- Committing docs with validate errors\n")
	b.WriteString("- \"The chapter doesn't apply\" → write why it's N/A, don't delete the heading\n")
	return b.String()
}
