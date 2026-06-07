# Contributing

## Branching Model (Git Flow)

Never commit directly to `main` or `develop` — both are protected.

| Branch       | From      | Merges into        | Purpose                      |
|--------------|-----------|--------------------|------------------------------|
| `feature/*`  | `develop` | `develop` (PR)     | New work                     |
| `bugfix/*`   | `develop` | `develop` (PR)     | Non-urgent fixes             |
| `hotfix/*`   | `main`    | `main` + `develop` | Urgent production fixes      |

### Branch naming
`<type>/<ticket-id>-<short-kebab-description>`
Examples: `feature/AAI-123-tool-calling-loop`, `bugfix/AAI-145-nil-agent-state`

## Commits
- Must be **signed** (the repo requires verified signatures).
- Follow **Conventional Commits**: `feat:`, `fix:`, `chore:`, `docs:`, `test:`, `refactor:`.

## Pull Requests
1. Branch from `develop`, push, open a PR into `develop`.
2. Required: passing checks, Code Owner review, **2 approvals for `main`**, 1 for `develop`.
3. Keep PRs small and focused. Resolve all conversations before merge.

## Reviewing Dependabot pull requests

Dependabot is enabled (see `.github/dependabot.yml`) and opens weekly PRs for Go
module and GitHub Actions updates. These PRs are **not** auto-merged — they pass
through the same governed gate as any change: all required checks plus
independent review. Apply this discipline when reviewing one:

1. **Check for pinned-Go drift (gomod PRs).** A dependency bump can pull in a
   `toolchain` directive or bump the `go` directive in `go.mod` as a side-effect.
   The CI guard ("Verify pinned Go toolchain") fails the PR if this happens — but
   confirm the diff does not silently alter the `go` line or add a `toolchain`
   line before approving. The pinned Go version changes only by deliberate human PR.

2. **Confirm commits are Verified.** Dependabot signs its commits, so they should
   show "Verified" and satisfy the signing rule. If a commit shows unsigned (a
   known intermittent Dependabot bug), do **not** weaken the signing rule. Remedy,
   in order:
   - Comment `@dependabot recreate` (may regenerate a signed commit; not guaranteed).
   - If that fails: check out the branch, re-apply the change as your own signed
     commit (or interactive-rebase to re-sign), and force-push. The PR becomes
     human-authored and signed — stronger attribution, not weaker.

3. **Merge with a merge commit, never rebase.** Rebase-merge rewrites the commit
   and strips the bot's signature, tripping the signing rule. Dependabot targets
   `develop`, where merge commits are allowed, so this is the default path.

4. **Review the change itself.** A grouped minor/patch PR still warrants reading
   the diff; a major-version or security bump warrants closer scrutiny. CI catches
   build, vulnerability, and lint regressions, but the dependency change is yours
   to judge.
