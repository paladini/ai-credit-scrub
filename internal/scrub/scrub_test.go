package scrub

import "testing"

func TestCleanBuiltInsAndPreservesNormalMentions(t *testing.T) {
	in := "Add docs\n\nGenerated with Claude Code\nCo-authored-by: Claude <noreply@anthropic.com>\n"
	out, matches, err := Clean(in, Config{Version: 1})
	if err != nil || out != "Add docs\n" || len(matches) != 2 {
		t.Fatalf("out=%q matches=%v err=%v", out, matches, err)
	}
	plain := "Codex reviewed Cursor output; Co-authored-by: Ada Lovelace <ada@example.com>\n"
	out, matches, err = Clean(plain, Config{Version: 1})
	if err != nil || out != plain || len(matches) != 0 {
		t.Fatalf("false positive: %q %#v %v", out, matches, err)
	}
}

func TestAllProviderFooters(t *testing.T) {
	for _, provider := range []string{"Codex", "Claude Code", "Cursor", "Windsurf", "GitHub Copilot"} {
		out, matches, err := Clean("Useful text\nGenerated with "+provider+"\n", Config{Version: 1})
		if err != nil || out != "Useful text\n" || len(matches) != 1 {
			t.Fatalf("%s: %q %#v %v", provider, out, matches, err)
		}
	}
}

func TestCustomRulesRequireReview(t *testing.T) {
	_, _, err := Clean("internal credit", Config{Version: 1, Literals: []string{"internal credit"}})
	if err == nil {
		t.Fatal("expected reviewed config error")
	}
	out, _, err := Clean("internal credit", Config{Version: 1, Reviewed: true, Literals: []string{"internal credit"}})
	if err != nil || out != "" {
		t.Fatalf("%q %v", out, err)
	}
}
