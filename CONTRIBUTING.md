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
