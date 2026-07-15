# Local enforcement

ai-credit-scrub protects text at the local publication boundaries that Git can
control. Install it in each repository:

```sh
ai-credit-scrub install --git
```

## Commit creation

Git passes the proposed message file to `commit-msg` before it writes the commit
object. ai-credit-scrub cleans that file in place. This covers terminal Git,
most IDE Git clients, and agents that invoke local Git.

## Push prevention

The `pre-push` hook receives the refs that are about to leave the machine. It
reads the outgoing commit messages and runs `check`. If it finds an explicit
credit, the push fails before Git contacts the remote with the update.

The hook does not rewrite an existing commit because rewriting published or
shared history would be unsafe. Amend the local commit, then push again.

## Pull request creation

Git does not carry a pull request title or body. Use the local wrapper when
creating a PR from the command line:

```sh
ai-credit-scrub pr create --title "Improve docs" --body-file pull-request.md
```

The wrapper cleans the supplied text and then calls `gh pr create`. It does not
need an ai-credit-scrub account, a hosted service, or an additional token; `gh`
uses your existing local GitHub authentication.

## Escape hatches

Users can deliberately bypass client hooks with `--no-verify`. The pre-push
hook catches a bypassed commit, but `git push --no-verify` bypasses that final
local check too. Treat that command as an intentional override.
