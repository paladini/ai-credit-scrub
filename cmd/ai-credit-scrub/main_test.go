package main

import (
	"strings"
	"testing"
)

func TestAdapterFilesDescribeAllProviders(t *testing.T) {
	for _, provider := range []string{"codex", "claude", "windsurf", "copilot", "cursor"} {
		path, body, err := adapterFile(provider)
		if err != nil || path == "" || !strings.Contains(body, "ai-credit-scrub pr create") {
			t.Fatalf("%s: path=%q body=%q err=%v", provider, path, body, err)
		}
	}
}

func TestGitHookBodiesPreserveExistingHooksAndGuardPushes(t *testing.T) {
	commit := commitMsgHook("/tmp/ai-credit-scrub")
	if !strings.Contains(commit, "commit-msg.ai-credit-scrub-original") || !strings.Contains(commit, "clean --in-place") {
		t.Fatalf("unexpected commit hook: %s", commit)
	}
	push := prePushHook("/tmp/ai-credit-scrub")
	if !strings.Contains(push, "pre-push.ai-credit-scrub-original") || !strings.Contains(push, "git log --format=%B") || !strings.Contains(push, "check") {
		t.Fatalf("unexpected pre-push hook: %s", push)
	}
}
