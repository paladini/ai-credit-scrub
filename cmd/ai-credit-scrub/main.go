package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/paladini/ai-credit-scrub/internal/scrub"
)

const usage = `ai-credit-scrub removes explicit AI-agent credit text before publication.

Usage:
  ai-credit-scrub clean [--config FILE] [--in-place] [FILE ...]
  ai-credit-scrub scan|check [--config FILE] [--format text|json] [FILE ...]
  ai-credit-scrub install [--git] [--adapter NAME]
  ai-credit-scrub pr create --title TEXT (--body TEXT | --body-file FILE) [gh flags...]
`

func main() {
	if len(os.Args) < 2 {
		die(usage)
	}
	switch os.Args[1] {
	case "clean":
		clean(os.Args[2:])
	case "scan", "check":
		scan(os.Args[1], os.Args[2:])
	case "install":
		install(os.Args[2:])
	case "pr":
		pr(os.Args[2:])
	case "--help", "help", "-h":
		fmt.Print(usage)
	default:
		die(usage)
	}
}

func clean(args []string) {
	fs := flag.NewFlagSet("clean", flag.ExitOnError)
	config := fs.String("config", findConfig(), "config file")
	inPlace := fs.Bool("in-place", false, "rewrite files")
	fs.Parse(args)
	cfg, err := scrub.Load(*config)
	must(err)
	files := fs.Args()
	if len(files) == 0 {
		data, err := io.ReadAll(os.Stdin)
		must(err)
		out, matches, err := scrub.Clean(string(data), cfg)
		must(err)
		fmt.Print(out)
		report(matches)
		return
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		must(err)
		out, matches, err := scrub.Clean(string(data), cfg)
		must(err)
		if *inPlace {
			must(os.WriteFile(file, []byte(out), 0o644))
		} else {
			fmt.Print(out)
		}
		report(matches)
	}
}

func scan(command string, args []string) {
	fs := flag.NewFlagSet(command, flag.ExitOnError)
	config := fs.String("config", findConfig(), "config file")
	format := fs.String("format", "text", "text or json")
	fs.Parse(args)
	cfg, err := scrub.Load(*config)
	must(err)
	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}
	found := false
	for _, file := range files {
		var data []byte
		if file == "-" {
			data, err = io.ReadAll(os.Stdin)
		} else {
			data, err = os.ReadFile(file)
		}
		must(err)
		matches, err := scrub.Scan(string(data), cfg)
		must(err)
		if len(matches) > 0 {
			found = true
		}
		if *format == "json" {
			must(json.NewEncoder(os.Stdout).Encode(map[string]any{"file": file, "matches": matches}))
		} else {
			for _, m := range matches {
				fmt.Printf("%s: %s: %s\n", file, m.Rule, strings.TrimSpace(m.Text))
			}
		}
	}
	if command == "check" && found {
		os.Exit(1)
	}
}

func install(args []string) {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	gitHook := fs.Bool("git", true, "install commit-msg hook")
	adapter := fs.String("adapter", "", "codex, claude, windsurf, copilot, or cursor")
	fs.Parse(args)
	if *gitHook {
		root := gitRoot()
		executable, err := os.Executable()
		must(err)
		must(installHook(root, "commit-msg", commitMsgHook(executable)))
		must(installHook(root, "pre-push", prePushHook(executable)))
		fmt.Println("installed Git commit-msg rewriter and pre-push guard; reinstall after moving this binary")
	}
	if *adapter != "" {
		root := gitRoot()
		path, body, err := adapterFile(*adapter)
		must(err)
		target := filepath.Join(root, path)
		if _, err := os.Stat(target); err == nil {
			die(target + " already exists; merge the generated adapter manually to preserve existing configuration")
		}
		must(os.MkdirAll(filepath.Dir(target), 0o755))
		must(os.WriteFile(target, []byte(body), 0o644))
		fmt.Println("installed", *adapter, "adapter:", path)
	}
}

func installHook(root, name, body string) error {
	hook := filepath.Join(root, ".git", "hooks", name)
	backup := hook + ".ai-credit-scrub-original"
	if err := os.MkdirAll(filepath.Dir(hook), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(hook); err == nil {
		if _, backupErr := os.Stat(backup); os.IsNotExist(backupErr) {
			if err := os.Rename(hook, backup); err != nil {
				return fmt.Errorf("preserve existing %s hook: %w", name, err)
			}
		}
	}
	return os.WriteFile(hook, []byte(body), 0o755)
}

func hookPreamble(name string) string {
	return "#!/bin/sh\nHOOK_DIR=$(CDPATH= cd -- \"$(dirname -- \"$0\")\" && pwd)\nORIGINAL=\"$HOOK_DIR/" + name + ".ai-credit-scrub-original\"\nif [ -x \"$ORIGINAL\" ]; then \"$ORIGINAL\" \"$@\" || exit $?; fi\n"
}

func commitMsgHook(executable string) string {
	return hookPreamble("commit-msg") + "\"" + executable + "\" clean --in-place \"$1\"\n"
}

func prePushHook(executable string) string {
	return hookPreamble("pre-push") + `while read local_ref local_sha remote_ref remote_sha; do
  if [ "$local_sha" = "0000000000000000000000000000000000000000" ]; then continue; fi
  if [ "$remote_sha" = "0000000000000000000000000000000000000000" ]; then range="$local_sha"; else range="$remote_sha..$local_sha"; fi
  git log --format=%B "$range" | "` + executable + `" check || {
    echo "ai-credit-scrub: push blocked; amend the affected commit or remove its explicit credit text." >&2
    exit 1
  }
done
`
}

func pr(args []string) {
	if len(args) == 0 || args[0] != "create" {
		die("usage: ai-credit-scrub pr create --title TEXT (--body TEXT | --body-file FILE) [gh flags...]")
	}
	args = args[1:]
	titleAt, bodyAt, bodyFileAt := -1, -1, -1
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--title":
			titleAt = i + 1
		case "--body":
			bodyAt = i + 1
		case "--body-file":
			bodyFileAt = i + 1
		}
	}
	if titleAt < 0 || titleAt >= len(args) || (bodyAt < 0 && bodyFileAt < 0) {
		die("pr create requires --title and --body or --body-file so content can be scrubbed")
	}
	cfg, err := scrub.Load(findConfig())
	must(err)
	cleanValue := func(value string) string {
		out, matches, err := scrub.Clean(value, cfg)
		must(err)
		report(matches)
		return out
	}
	args[titleAt] = cleanValue(args[titleAt])
	if bodyAt >= 0 && bodyAt < len(args) {
		args[bodyAt] = cleanValue(args[bodyAt])
	}
	if bodyFileAt >= 0 && bodyFileAt < len(args) {
		data, err := os.ReadFile(args[bodyFileAt])
		must(err)
		out := cleanValue(string(data))
		tmp, err := os.CreateTemp("", "ai-credit-scrub-pr-*.md")
		must(err)
		defer os.Remove(tmp.Name())
		_, err = tmp.WriteString(out)
		must(err)
		must(tmp.Close())
		args[bodyFileAt] = tmp.Name()
	}
	cmd := exec.Command("gh", append([]string{"pr", "create"}, args...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	must(cmd.Run())
}

func adapterFile(name string) (string, string, error) {
	message := "Use ai-credit-scrub pr create for pull requests and allow the Git commit-msg hook to sanitize commit messages."
	switch name {
	case "codex":
		return ".codex/hooks.json", `{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"echo '{\"systemMessage\":\"` + message + `\"}'"}]}]}}`, nil
	case "claude":
		return ".claude/settings.json", `{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"echo '` + message + `'"}]}]}}`, nil
	case "windsurf":
		return ".windsurf/hooks.json", `{"hooks":{"pre_run_command":[{"command":"echo '` + message + `'","show_output":true}]}}`, nil
	case "copilot":
		return ".github/hooks/ai-credit-scrub.json", `{"hooks":{"sessionStart":[{"type":"command","bash":"echo '` + message + `'","powershell":"Write-Output '` + message + `'","cwd":"."}]}}`, nil
	case "cursor":
		return ".cursor/rules/ai-credit-scrub.mdc", "---\ndescription: Publication hygiene\nalwaysApply: true\n---\n" + message + "\n", nil
	default:
		return "", "", errors.New("unknown adapter; choose codex, claude, windsurf, copilot, or cursor")
	}
}

func findConfig() string {
	if _, err := os.Stat(".ai-credit-scrub.yml"); err == nil {
		return ".ai-credit-scrub.yml"
	}
	return ""
}
func gitRoot() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	must(err)
	return strings.TrimSpace(string(out))
}
func report(matches []scrub.Match) {
	if len(matches) > 0 {
		fmt.Fprintf(os.Stderr, "ai-credit-scrub: removed %d explicit credit block(s)\n", len(matches))
	}
}
func must(err error) {
	if err != nil {
		die(err.Error())
	}
}
func die(message string) { fmt.Fprintln(os.Stderr, "ai-credit-scrub:", message); os.Exit(2) }
