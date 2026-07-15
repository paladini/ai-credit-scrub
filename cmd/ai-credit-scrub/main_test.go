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
