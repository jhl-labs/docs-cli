// Package cli wires the docs-cli subcommands together. Argument parsing is
// hand-rolled (no external dependency) to match the house CLI style.
package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jhl-labs/docs-cli/internal/agent"
	"github.com/jhl-labs/docs-cli/internal/project"
	"github.com/jhl-labs/docs-cli/internal/render"
	"github.com/jhl-labs/docs-cli/internal/scaffold"
	"github.com/jhl-labs/docs-cli/internal/schema"
	"github.com/jhl-labs/docs-cli/internal/skill"
	"github.com/jhl-labs/docs-cli/internal/validate"
)

// Exit codes, aligned with the house DevSecOps CLIs.
const (
	ExitOK          = 0
	ExitValidation  = 1
	ExitUsage       = 2
	ExitAgentFailed = 3
	ExitEnvironment = 4
)

// Run dispatches a command. It returns a process exit code.
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printHelp(stdout)
		return ExitOK
	}
	switch args[0] {
	case "-h", "--help", "help":
		printHelp(stdout)
		return ExitOK
	case "-v", "--version", "version":
		fmt.Fprintf(stdout, "docs-cli %s\n", versionString())
		return ExitOK
	case "--generate-skill":
		return runSkill(args[1:], stdout, stderr)
	case "init":
		return runInit(args[1:], stdout, stderr)
	case "generate":
		return runGenerate(args[1:], stdout, stderr)
	case "render":
		return runRender(args[1:], stdout, stderr)
	case "validate":
		return runValidate(args[1:], stdout, stderr)
	case "doctor":
		return runDoctor(args[1:], stdout, stderr)
	case "skill":
		return runSkill(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printHelp(stderr)
		return ExitUsage
	}
}

// ---- init ----

func runInit(args []string, stdout, stderr io.Writer) int {
	var (
		outputDir = "docs"
		lang      = "auto"
		force     bool
		target    = "."
	)
	rest, err := parseFlags(args, map[string]*string{
		"--output-dir": &outputDir,
		"--lang":       &lang,
	}, map[string]*bool{
		"--force": &force,
	})
	if err != nil {
		return usageError(stderr, err)
	}
	if helpRequested(rest) {
		printInitHelp(stdout)
		return ExitOK
	}
	if p, ok := firstPositional(rest); ok {
		target = p
	}

	resolvedLang := resolveLang(lang, target)
	s := schema.Standard()
	files := scaffold.Build(s, scaffold.Options{
		OutputDir: outputDir,
		Lang:      resolvedLang,
		Date:      today(),
		Commit:    gitCommit(target),
		Force:     force,
	})
	written, skipped, err := scaffold.Write(files, force)
	if err != nil {
		fmt.Fprintf(stderr, "scaffold: %v\n", err)
		return ExitEnvironment
	}
	for _, p := range written {
		fmt.Fprintf(stdout, "created %s\n", p)
	}
	if len(skipped) > 0 {
		fmt.Fprintf(stdout, "\nskipped %d existing file(s) (use --force to overwrite)\n", len(skipped))
	}
	fmt.Fprintf(stdout, "\nscaffolded standardized docs (schema v%s) into %s/ — primary language: %s\n", s.Version, outputDir, resolvedLang)
	fmt.Fprintf(stdout, "next: docs-cli generate %s --agent <claude|codex|gemini|opencode>\n", target)
	return ExitOK
}

// ---- generate ----

func runGenerate(args []string, stdout, stderr io.Writer) int {
	var (
		outputDir   = "docs"
		promptsDir  = ".docs-cli/prompts"
		agentName   = "none"
		lang        = "auto"
		target      = "."
		dryRun      bool
		printPrompt bool
	)
	var docIDs multiFlag
	rest, err := parseFlagsWithMulti(args, map[string]*string{
		"--output-dir":  &outputDir,
		"--prompts-dir": &promptsDir,
		"--agent":       &agentName,
		"--lang":        &lang,
	}, map[string]*bool{
		"--dry-run":      &dryRun,
		"--print-prompt": &printPrompt,
	}, map[string]*multiFlag{
		"--doc": &docIDs,
	})
	if err != nil {
		return usageError(stderr, err)
	}
	if helpRequested(rest) {
		printGenerateHelp(stdout)
		return ExitOK
	}
	if p, ok := firstPositional(rest); ok {
		target = p
	}

	s := schema.Standard()
	docs, err := selectDocs(s, docIDs)
	if err != nil {
		return usageError(stderr, err)
	}
	resolvedLang := resolveLang(lang, target)

	useAgent := strings.ToLower(strings.TrimSpace(agentName)) != "none" && agentName != ""
	var ag agent.Agent
	if useAgent {
		var ok bool
		ag, ok = agent.Lookup(agentName)
		if !ok {
			return usageError(stderr, fmt.Errorf("unknown agent %q (known: %s)", agentName, strings.Join(agent.Known(), ", ")))
		}
		if !dryRun && !agent.Available(agentName) {
			fmt.Fprintf(stderr, "agent %q not found on PATH; rerun with --dry-run to emit prompts only\n", agentName)
			return ExitEnvironment
		}
	}

	if err := os.MkdirAll(promptsDir, 0o755); err != nil {
		fmt.Fprintf(stderr, "create prompts dir: %v\n", err)
		return ExitEnvironment
	}

	promptOpts := agent.PromptOptions{
		ProjectDir:    target,
		Language:      resolvedLang,
		OutputDir:     outputDir,
		SchemaVersion: s.Version,
	}

	runner := agentRunner
	generated := 0
	for _, doc := range docs {
		prompt := agent.BuildPrompt(doc, promptOpts)
		promptPath := filepath.Join(promptsDir, doc.ID+".prompt.md")
		if err := os.WriteFile(promptPath, []byte(prompt), 0o644); err != nil {
			fmt.Fprintf(stderr, "write prompt %s: %v\n", promptPath, err)
			return ExitEnvironment
		}
		fmt.Fprintf(stdout, "prompt %s\n", promptPath)
		if printPrompt {
			fmt.Fprintf(stdout, "\n--- prompt for %s ---\n%s\n", doc.ID, prompt)
		}

		if !useAgent || dryRun {
			continue
		}

		bin, cmdArgs := ag.Command(prompt)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		out, runErr := runner.Run(ctx, target, bin, cmdArgs)
		cancel()
		if runErr != nil {
			fmt.Fprintf(stderr, "agent %s failed for %s: %v\n", agentName, doc.ID, runErr)
			return ExitAgentFailed
		}
		docPath := filepath.Join(outputDir, doc.FileName())
		if err := os.MkdirAll(filepath.Dir(docPath), 0o755); err != nil {
			fmt.Fprintf(stderr, "create dir for %s: %v\n", docPath, err)
			return ExitEnvironment
		}
		if err := os.WriteFile(docPath, []byte(ensureTrailingNewline(out)), 0o644); err != nil {
			fmt.Fprintf(stderr, "write %s: %v\n", docPath, err)
			return ExitEnvironment
		}
		fmt.Fprintf(stdout, "generated %s (agent: %s)\n", docPath, agentName)
		generated++
	}

	if useAgent && !dryRun {
		fmt.Fprintf(stdout, "\nfilled %d document(s); run `docs-cli validate %s` before committing\n", generated, outputDir)
	} else {
		fmt.Fprintf(stdout, "\nwrote %d prompt(s) to %s; hand them to an agent and write results to %s/\n", len(docs), promptsDir, outputDir)
	}
	return ExitOK
}

// agentRunner is overridable in tests.
var agentRunner agent.Runner = agent.ExecRunner{}

func selectDocs(s schema.Schema, ids []string) ([]schema.Document, error) {
	if len(ids) == 0 {
		return s.Docs, nil
	}
	var out []schema.Document
	for _, id := range ids {
		doc, ok := s.Doc(id)
		if !ok {
			return nil, fmt.Errorf("unknown doc id %q", id)
		}
		out = append(out, doc)
	}
	return out, nil
}

// ---- render ----

func runRender(args []string, stdout, stderr io.Writer) int {
	var (
		inputDir  = "docs"
		outputDir = ""
		target    = ""
	)
	var formats multiFlag
	rest, err := parseFlagsWithMulti(args, map[string]*string{
		"--input-dir":  &inputDir,
		"--output-dir": &outputDir,
	}, map[string]*bool{}, map[string]*multiFlag{
		"--format": &formats,
	})
	if err != nil {
		return usageError(stderr, err)
	}
	if helpRequested(rest) {
		printRenderHelp(stdout)
		return ExitOK
	}
	if p, ok := firstPositional(rest); ok {
		target = p
		// When a project path is given, default input to <path>/docs.
		if inputDir == "docs" {
			inputDir = filepath.Join(target, "docs")
		}
	}
	if len(formats) == 0 {
		formats = multiFlag{"html"}
	}

	mdFiles, err := collectMarkdown(inputDir)
	if err != nil {
		fmt.Fprintf(stderr, "scan %s: %v\n", inputDir, err)
		return ExitEnvironment
	}
	if len(mdFiles) == 0 {
		fmt.Fprintf(stderr, "no Markdown files under %s\n", inputDir)
		return ExitUsage
	}

	count := 0
	for _, format := range formats {
		ext, subdir, err := formatTarget(format)
		if err != nil {
			return usageError(stderr, err)
		}
		base := outputDir
		if base == "" {
			base = filepath.Join(inputDir, subdir)
		}
		for _, mdPath := range mdFiles {
			raw, err := os.ReadFile(mdPath)
			if err != nil {
				fmt.Fprintf(stderr, "read %s: %v\n", mdPath, err)
				return ExitEnvironment
			}
			var rendered string
			switch format {
			case "html":
				rendered, err = render.HTML(raw)
			case "xml":
				rendered, err = render.XML(raw)
			}
			if err != nil {
				fmt.Fprintf(stderr, "render %s: %v\n", mdPath, err)
				return ExitUsage
			}
			rel, _ := filepath.Rel(inputDir, mdPath)
			outPath := filepath.Join(base, strings.TrimSuffix(rel, ".md")+ext)
			if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
				fmt.Fprintf(stderr, "create dir for %s: %v\n", outPath, err)
				return ExitEnvironment
			}
			if err := os.WriteFile(outPath, []byte(rendered), 0o644); err != nil {
				fmt.Fprintf(stderr, "write %s: %v\n", outPath, err)
				return ExitEnvironment
			}
			fmt.Fprintf(stdout, "rendered %s\n", outPath)
			count++
		}
	}
	fmt.Fprintf(stdout, "\nrendered %d file(s) across %d format(s)\n", count, len(formats))
	return ExitOK
}

func formatTarget(format string) (ext, subdir string, err error) {
	switch strings.ToLower(format) {
	case "html":
		return ".html", "_site", nil
	case "xml":
		return ".xml", "_xml", nil
	default:
		return "", "", fmt.Errorf("unsupported format %q (html|xml)", format)
	}
}

func collectMarkdown(dir string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if path != dir && (name == "_site" || name == "_xml") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".md") {
			out = append(out, path)
		}
		return nil
	})
	return out, err
}

// ---- validate ----

func runValidate(args []string, stdout, stderr io.Writer) int {
	var (
		inputDir = "docs"
		strict   bool
		target   = ""
	)
	rest, err := parseFlags(args, map[string]*string{
		"--input-dir": &inputDir,
	}, map[string]*bool{
		"--strict": &strict,
	})
	if err != nil {
		return usageError(stderr, err)
	}
	if helpRequested(rest) {
		printValidateHelp(stdout)
		return ExitOK
	}
	if p, ok := firstPositional(rest); ok {
		target = p
		if inputDir == "docs" {
			inputDir = filepath.Join(target, "docs")
		}
	}

	report := validate.Run(schema.Standard(), inputDir)
	fmt.Fprint(stdout, validate.FormatText(report))
	if report.Errors() > 0 {
		return ExitValidation
	}
	if strict && report.Warnings() > 0 {
		fmt.Fprintln(stderr, "strict mode: warnings present")
		return ExitValidation
	}
	return ExitOK
}

// ---- doctor ----

func runDoctor(args []string, stdout, stderr io.Writer) int {
	var inputDir = "docs"
	rest, err := parseFlags(args, map[string]*string{
		"--input-dir": &inputDir,
	}, map[string]*bool{})
	if err != nil {
		return usageError(stderr, err)
	}
	if helpRequested(rest) {
		printDoctorHelp(stdout)
		return ExitOK
	}

	s := schema.Standard()
	fmt.Fprintf(stdout, "docs-cli %s\n", versionString())
	fmt.Fprintf(stdout, "schema: v%s (%d documents)\n", s.Version, len(s.Docs))

	failures := 0
	if err := checkWritable("."); err != nil {
		failures++
		fmt.Fprintf(stdout, "[fail] working dir: %v\n", err)
	} else {
		fmt.Fprintln(stdout, "[ok] working dir writable")
	}

	anyAgent := false
	for _, name := range agent.Known() {
		if agent.Available(name) {
			fmt.Fprintf(stdout, "[ok] agent %s available\n", name)
			anyAgent = true
		} else {
			fmt.Fprintf(stdout, "[warn] agent %s not found on PATH\n", name)
		}
	}
	if !anyAgent {
		fmt.Fprintln(stdout, "[warn] no AI agent found; use `generate --dry-run` to emit prompts")
	}

	if _, err := os.Stat(inputDir); err == nil {
		report := validate.Run(s, inputDir)
		fmt.Fprintf(stdout, "[info] %s: %d error(s), %d warning(s)\n", inputDir, report.Errors(), report.Warnings())
	} else {
		fmt.Fprintf(stdout, "[info] %s not found; run `docs-cli init` to scaffold\n", inputDir)
	}

	if failures > 0 {
		return ExitEnvironment
	}
	return ExitOK
}

// ---- skill ----

func runSkill(args []string, stdout, stderr io.Writer) int {
	var (
		output = "skill.md"
		name   = "standardizing-project-docs"
		binary = "docs-cli"
	)
	rest, err := parseFlags(args, map[string]*string{
		"--output": &output,
		"--name":   &name,
		"--binary": &binary,
	}, map[string]*bool{})
	if err != nil {
		return usageError(stderr, err)
	}
	if helpRequested(rest) {
		printSkillHelp(stdout)
		return ExitOK
	}

	content := skill.Generate(schema.Standard(), skill.Options{Name: name, Binary: binary})
	if output == "-" || output == "" {
		fmt.Fprint(stdout, content)
		return ExitOK
	}
	if dir := filepath.Dir(output); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(stderr, "create dir for %s: %v\n", output, err)
			return ExitEnvironment
		}
	}
	if err := os.WriteFile(output, []byte(content), 0o644); err != nil {
		fmt.Fprintf(stderr, "write %s: %v\n", output, err)
		return ExitEnvironment
	}
	fmt.Fprintf(stdout, "wrote %s\n", output)
	return ExitOK
}

// ---- shared helpers ----

func resolveLang(lang, target string) string {
	if lang != "" && lang != "auto" {
		return lang
	}
	return string(project.Detect(target).Primary)
}

func ensureTrailingNewline(s string) string {
	if s == "" || strings.HasSuffix(s, "\n") {
		return s
	}
	return s + "\n"
}

func checkWritable(dir string) error {
	probe := filepath.Join(dir, ".docs-cli-write-probe")
	if err := os.WriteFile(probe, []byte("ok"), 0o644); err != nil {
		return err
	}
	return os.Remove(probe)
}

func today() string {
	return time.Now().UTC().Format("2006-01-02")
}

func gitCommit(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func usageError(stderr io.Writer, err error) int {
	fmt.Fprintln(stderr, err)
	return ExitUsage
}
