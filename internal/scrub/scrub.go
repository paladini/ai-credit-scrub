package scrub

import (
	"fmt"
	"regexp"
	"strings"
)

// Match describes one conservative removal made by the engine.
type Match struct {
	Rule string `json:"rule"`
	Text string `json:"text"`
}

type rule struct {
	name string
	re   *regexp.Regexp
}

var builtins = []rule{
	{"codex-footer", regexp.MustCompile(`(?im)^\s*(?:🤖\s*)?(?:generated|created|written) (?:with|by) (?:OpenAI )?Codex\.?\s*$\r?\n?`)},
	{"claude-footer", regexp.MustCompile(`(?im)^\s*(?:🤖\s*)?(?:generated|created|written) (?:with|by) Claude(?: Code)?\.?\s*$\r?\n?`)},
	{"cursor-footer", regexp.MustCompile(`(?im)^\s*(?:🤖\s*)?(?:generated|created|written) (?:with|by) Cursor\.?\s*$\r?\n?`)},
	{"windsurf-footer", regexp.MustCompile(`(?im)^\s*(?:🤖\s*)?(?:generated|created|written) (?:with|by) Windsurf(?: Cascade)?\.?\s*$\r?\n?`)},
	{"copilot-footer", regexp.MustCompile(`(?im)^\s*(?:🤖\s*)?(?:generated|created|written) (?:with|by) GitHub Copilot\.?\s*$\r?\n?`)},
	{"claude-trailer", regexp.MustCompile(`(?im)^\s*co-authored-by:\s*claude(?: code)?\s*<[^>]*anthropic[^>]*>\s*$\r?\n?`)},
	{"codex-trailer", regexp.MustCompile(`(?im)^\s*co-authored-by:\s*(?:openai )?codex\s*<[^>]*(?:openai|codex)[^>]*>\s*$\r?\n?`)},
	{"copilot-trailer", regexp.MustCompile(`(?im)^\s*co-authored-by:\s*(?:github )?copilot\s*<[^>]*(?:github|copilot)[^>]*>\s*$\r?\n?`)},
}

// Clean removes explicit agent-credit lines only. It does not match product names
// by themselves, so ordinary prose such as "reviewed in Cursor" is retained.
func Clean(input string, cfg Config) (string, []Match, error) {
	if err := cfg.Validate(); err != nil {
		return "", nil, err
	}
	rules := append([]rule{}, builtins...)
	for i, literal := range cfg.Literals {
		rules = append(rules, rule{fmt.Sprintf("literal-%d", i+1), regexp.MustCompile(regexp.QuoteMeta(literal))})
	}
	for i, pattern := range cfg.Regex {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return "", nil, fmt.Errorf("invalid custom regex %q: %w", pattern, err)
		}
		rules = append(rules, rule{fmt.Sprintf("regex-%d", i+1), re})
	}

	output := input
	matches := []Match{}
	for _, r := range rules {
		for _, found := range r.re.FindAllString(output, -1) {
			if !excluded(found, cfg.Exclude) {
				matches = append(matches, Match{Rule: r.name, Text: found})
			}
		}
		output = r.re.ReplaceAllStringFunc(output, func(found string) string {
			if excluded(found, cfg.Exclude) {
				return found
			}
			return ""
		})
	}
	output = strings.TrimRight(output, " \t\r\n") + trailingNewline(input)
	return output, matches, nil
}

func Scan(input string, cfg Config) ([]Match, error) {
	_, matches, err := Clean(input, cfg)
	return matches, err
}

func excluded(text string, values []string) bool {
	for _, value := range values {
		if strings.Contains(text, value) {
			return true
		}
	}
	return false
}

func trailingNewline(input string) string {
	if strings.HasSuffix(input, "\n") {
		return "\n"
	}
	return ""
}
