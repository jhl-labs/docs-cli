package cli

import (
	"fmt"
	"strings"

	"github.com/jhl-labs/docs-cli/internal/version"
)

// multiFlag is a repeatable string flag.
type multiFlag []string

func versionString() string { return version.String() }

// parseFlags parses args against the given string and bool flag maps. It
// supports both "--flag value" and "--flag=value" forms. Unconsumed tokens
// (positionals and -h/--help) are returned in rest.
func parseFlags(args []string, strs map[string]*string, bools map[string]*bool) (rest []string, err error) {
	return parseFlagsWithMulti(args, strs, bools, nil)
}

// parseFlagsWithMulti is parseFlags plus repeatable flags.
func parseFlagsWithMulti(args []string, strs map[string]*string, bools map[string]*bool, multis map[string]*multiFlag) (rest []string, err error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--" {
			rest = append(rest, args[i+1:]...)
			break
		}
		if arg == "-h" || arg == "--help" {
			rest = append(rest, arg)
			continue
		}

		// --flag=value form.
		if name, value, ok := strings.Cut(arg, "="); ok && strings.HasPrefix(name, "--") {
			if p, ok := strs[name]; ok {
				*p = value
				continue
			}
			if p, ok := multis[name]; ok {
				*p = append(*p, value)
				continue
			}
			if _, ok := bools[name]; ok {
				return rest, fmt.Errorf("flag %s does not take a value", name)
			}
			return rest, fmt.Errorf("unknown flag %q", name)
		}

		if p, ok := bools[arg]; ok {
			*p = true
			continue
		}
		if p, ok := strs[arg]; ok {
			value, e := requireValue(args, i+1, arg)
			if e != nil {
				return rest, e
			}
			*p = value
			i++
			continue
		}
		if p, ok := multis[arg]; ok {
			value, e := requireValue(args, i+1, arg)
			if e != nil {
				return rest, e
			}
			*p = append(*p, value)
			i++
			continue
		}

		if strings.HasPrefix(arg, "-") {
			return rest, fmt.Errorf("unknown flag %q", arg)
		}
		rest = append(rest, arg)
	}
	return rest, nil
}

func requireValue(args []string, idx int, flag string) (string, error) {
	if idx >= len(args) {
		return "", fmt.Errorf("flag %s requires a value", flag)
	}
	return args[idx], nil
}

func helpRequested(rest []string) bool {
	for _, r := range rest {
		if r == "-h" || r == "--help" {
			return true
		}
	}
	return false
}

func firstPositional(rest []string) (string, bool) {
	for _, r := range rest {
		if r == "-h" || r == "--help" {
			continue
		}
		return r, true
	}
	return "", false
}
