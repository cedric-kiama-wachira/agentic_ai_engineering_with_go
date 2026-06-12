# Contributing

## Branching Model (Git Flow)

Never commit directly to `main` or `develop` — both are protected.

| Branch       | From      | Merges into        | Purpose                      |
|--------------|-----------|--------------------|------------------------------|
| `feature/*`  | `develop` | `develop` (PR)     | New work                     |
| `bugfix/*`   | `develop` | `develop` (PR)     | Non-urgent fixes             |
| `hotfix/*`   | `main`    | `main` + `develop` | Urgent production fixes      |
| `chore/*`    | `develop` | `develop` (PR)     | Repo plumbing: CI, tooling, templates, config |
| `docs/*`     | `develop` | `develop` (PR)     | Documentation-only changes   |

### Branch naming
`<type>/<short-kebab-description>` — with a ticket ID after the type once a
tracker is in use: `<type>/<ticket-id>-<short-kebab-description>`.
Examples: `feature/agents-md`, `chore/pr-template-ai-checklist`,
`feature/AAI-123-tool-calling-loop` (with tracker).

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
   - Try `@dependabot recreate` **once, on the still-open PR** — it may regenerate
     a signed commit. It is not guaranteed, and it is fragile: on a closed PR or a
     deleted branch it can fail and even reference `@dependabot reopen`, a command
     GitHub removed on 2026-01-27.
   - If that fails, **author the bump as a normal signed human PR** (re-apply the
     change on a branch, commit it signed with your own key, open a PR). This is
     the reliable path and yields stronger attribution — a human author, signed.
   - For closing or reopening PRs, use GitHub's native UI/CLI: the `@dependabot
     close` / `reopen` / `merge` commands were removed on 2026-01-27.

3. **Merge with a merge commit, never rebase.** Rebase-merge rewrites the commit
   and strips the bot's signature, tripping the signing rule. Dependabot targets
   `develop`, where merge commits are allowed, so this is the default path.

4. **Review the change itself.** A grouped minor/patch PR still warrants reading
   the diff; a major-version or security bump warrants closer scrutiny. CI catches
   build, vulnerability, and lint regressions, but the dependency change is yours
   to judge. (Major **action** bumps are not proposed by Dependabot — they are
   suppressed in `dependabot.yml` and handled as deliberate human PRs; major
   **gomod** bumps, including security-tool majors, do surface for review.)
