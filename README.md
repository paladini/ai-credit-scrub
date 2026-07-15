<div align="center">

# ai-credit-scrub

### Local guardrails for AI coding credits.

Remove explicit AI-agent credit text before it becomes a Git commit, reaches a
remote, or is sent to GitHub as a pull request.

[![Release](https://img.shields.io/github/v/release/paladini/ai-credit-scrub?display_name=tag&sort=semver)](https://github.com/paladini/ai-credit-scrub/releases)
[![CI](https://github.com/paladini/ai-credit-scrub/actions/workflows/ci.yml/badge.svg)](https://github.com/paladini/ai-credit-scrub/actions/workflows/ci.yml)
[![Go version](https://img.shields.io/github/go-mod/go-version/paladini/ai-credit-scrub)](go.mod)
[![License](https://img.shields.io/github/license/paladini/ai-credit-scrub)](LICENSE)
[![Website](https://img.shields.io/website?url=https%3A%2F%2Fpaladini.github.io%2Fai-credit-scrub%2F&label=website)](https://paladini.github.io/ai-credit-scrub/)

[Website](https://paladini.github.io/ai-credit-scrub/) ·
[Quick start](#quick-start) ·
[How it works](#how-it-works) ·
[Integrations](docs/integrations.md) ·
[Releases](https://github.com/paladini/ai-credit-scrub/releases) ·
[Contributing](CONTRIBUTING.md)

</div>

## The problem

Coding agents can append a “Generated with …” footer or an agent
`Co-authored-by` trailer to a commit or pull request. Finding it in CI is too
late: the text already exists in history or on GitHub.

ai-credit-scrub acts where the developer still has control. It is a free,
offline-first Go CLI with deterministic rules—no model calls, account, hosted
service, prompt upload, or source-code upload.

```text
agent / IDE proposes text
          │
          ├── commit-msg ──► rewrite before Git creates the commit
          ├── pre-push   ──► block an escaped credit before the remote update
          └── pr create  ──► clean title and body before calling gh
```

## Quick start

Install a binary from [Releases](https://github.com/paladini/ai-credit-scrub/releases),
or use Go:

```sh
go install github.com/paladini/ai-credit-scrub/cmd/ai-credit-scrub@latest
```

Protect the current repository:

```sh
ai-credit-scrub install --git
```

That single command installs two chained local hooks:

| Boundary | What happens | Why it matters |
| --- | --- | --- |
| `commit-msg` | Rewrites the temporary message file. | The credit never enters the new commit. |
| `pre-push` | Checks outgoing commit messages and rejects a match. | A `git commit --no-verify` bypass is caught before push. |

Use the local wrapper when creating a pull request:

```sh
ai-credit-scrub pr create \
  --title "Document local hooks" \
  --body "Explain the guardrail.\n\nGenerated with Claude Code"
```

The wrapper cleans the supplied title and body, then delegates to your existing
local `gh pr create` authentication.

## How it works

ai-credit-scrub removes complete, explicit credit lines and known agent trailers
for Codex, Claude Code, Cursor, Windsurf, and GitHub Copilot. It deliberately
does **not** match a product name alone.

| Input | Result |
| --- | --- |
| `Generated with Claude Code` | Removed. |
| `Co-authored-by: Claude <noreply@anthropic.com>` | Removed. |
| `Reviewed in Cursor` | Preserved. |
| `Co-authored-by: Ada Lovelace <ada@example.com>` | Preserved. |

You can inspect text without changing it:

```sh
ai-credit-scrub scan CHANGELOG.md
```

Custom rules are supported through `.ai-credit-scrub.yml`, but require an
explicit `reviewed: true` acknowledgement after you review their matches. See
the [local enforcement guide](docs/local-enforcement.md) for the full policy.

## Designed for local developer control

- **No CI/CD enforcement surface.** This project prevents accidental publication
  locally instead of reporting it afterward.
- **No identity rewrite.** Git author and committer information stay untouched.
- **No forced cloud dependency.** Cleaning and checking text require no network
  access. Only `pr create` calls your already-authenticated GitHub CLI.
- **Existing hooks stay intact.** Installation preserves and chains an existing
  hook before ai-credit-scrub runs.

## Tool integrations

Git hooks provide the universal enforcement layer. Optional adapters add local
guidance for the tools that use your repository:

```sh
ai-credit-scrub install --adapter codex
ai-credit-scrub install --adapter claude
ai-credit-scrub install --adapter windsurf
ai-credit-scrub install --adapter copilot
ai-credit-scrub install --adapter cursor
```

Read [agent integration boundaries](docs/integrations.md) before enabling an
adapter in a repository that already has agent configuration.

## Honest boundaries

`git push --no-verify` is an intentional client-side override. A pull request
created directly through the GitHub website or a separate GitHub-writing MCP
server also bypasses Git. Use `ai-credit-scrub pr create`, or replace that
writer with a local sanitized MCP proxy you control.

Confirm that removing a credit is compatible with your organization’s policy,
contributor agreement, and applicable law.

## Project

ai-credit-scrub is MIT licensed. Contributions must keep matching conservative:
every new signature needs both a removal fixture and a false-positive fixture.

- [Contributing guide](CONTRIBUTING.md)
- [Security policy](SECURITY.md)
- [Support](SUPPORT.md)
- [Code of conduct](CODE_OF_CONDUCT.md)

```sh
go test ./...
go vet ./...
node scripts/validate-site.mjs
```
