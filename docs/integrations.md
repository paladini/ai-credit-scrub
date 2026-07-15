# Agent integrations

Install the Git guardrail first. It is the only integration that operates at a
universal Git boundary:

```sh
ai-credit-scrub install --git
```

Then add an optional adapter for your coding tool. Adapters are local guidance
for publication commands; they do not claim to intercept every possible remote
API call.

```sh
ai-credit-scrub install --adapter codex
ai-credit-scrub install --adapter claude
ai-credit-scrub install --adapter windsurf
ai-credit-scrub install --adapter copilot
ai-credit-scrub install --adapter cursor
```

The installer refuses to overwrite an existing adapter file. Merge the emitted
configuration manually when a repository already owns that file.

## GitHub MCP and PRs

An agent that creates a pull request through an independent GitHub MCP server
can bypass local Git because PR text is not part of a commit. There are two
honest local-first options:

1. Instruct the agent to use `ai-credit-scrub pr create` and its local `gh`
   installation for PR creation.
2. Replace the GitHub MCP server with a local proxy MCP server that exposes only
   sanitized create/update PR tools and holds the existing GitHub credential.

The first option is available now. The second is a separate integration project
because it must be the only GitHub-writing server available to the agent in
order to be enforceable.
