# AI Assistant Instructions

These instructions govern any AI coding assistant contributing to this
repository. They are derived from the OpenSSF *Security-Focused Guide for AI
Code Assistant Instructions* and this repository's governance model
(`docs/REPO_SETUP.md`, `CONTRIBUTING.md`).

**Authority:** these instructions do not relax any repository control. Where
they conflict with `CONTRIBUTING.md` or the branch rulesets, the stricter rule
wins. AI output is an untrusted contribution: it receives the identical
governed gate as human work — reviewed PR, signed commit, passing required
checks, clean secret scan. No exceptions.

## Hard rules

1. **Never commit secrets.** No credentials, tokens, API keys, private keys,
   or `.env` files — not in code, tests, fixtures, comments, or commit
   messages. Use `.env.example` placeholders. The repository is secret-scanned;
   a leak blocks the merge and triggers rotation.

2. **Validate all external input at trust boundaries.** Use allowlists, not
   denylists. Treat all network, file, environment, and LLM-derived input as
   untrusted. Never interpolate untrusted data into commands, SQL, templates,
   or file paths.

3. **Every new code path ships with tests.** New or changed behavior requires
   new or updated tests in the same change. Tests must pass locally
   (`go test -race ./...`) before a PR is opened.

4. **No new dependencies without written justification.** Adding to `go.mod`
   requires a stated rationale in the PR body: why it is needed, why the
   standard library is insufficient, and a note on the module's maintenance
   posture. Dependency manifests are CODEOWNERS-guarded.

5. **Flag uncertainty; never fabricate.** If unsure whether an API, flag,
   package, or behavior exists, say so explicitly rather than inventing it.
   Hallucinated APIs and invented module paths are treated as defects.

6. **Errors are handled, never swallowed.** Wrap errors with context
   (`fmt.Errorf("...: %w", err)`); do not discard them with `_` except with an
   auditable justification comment.

## Repository-specific constraints

- **Go toolchain is pinned.** The `go` directive in `go.mod` is exactly
  `go 1.26.4`. Never add a `toolchain` directive and never bump the `go`
  directive — CI enforces both and the build will fail.
- **Security tooling runs via the Go `tool` directive**, not marketplace
  actions (air-gap portability). Do not add marketplace actions; if one is
  unavoidable, it must be pinned to a full commit SHA with a `# vX.Y.Z`
  comment.
- **Suppressions are auditable.** Any `#nosec` requires a rule ID and written
  justification (CI enforces this). Any `//lint:ignore` or staticcheck
  exclusion requires an inline rationale comment.
- **Commits:** Conventional Commits, types `feat|fix|chore|docs|test|refactor`
  only. All commits are signed; unsigned commits are rejected by ruleset.
- **Branching:** branch from `develop` (`feature/*`, `bugfix/*`); never target
  `main` directly; never rebase shared branches (rebase strips signatures).

## What the assistant must not do

- Bypass, weaken, or propose disabling any required check, ruleset, or
  signing requirement — even temporarily, even on a branch.
- Modify `.github/workflows/`, `.github/dependabot.yml`, `CODEOWNERS`,
  `go.mod`, or `go.sum` without explicitly calling the change out in the PR
  description.
- Generate code that disables TLS verification, logs secrets, broadens file
  permissions, or executes dynamically constructed shell commands.
- Claim tests pass, tools ran, or research was done when it was not.

## Review expectation

Every AI-assisted PR is reviewed by a human under the AI Assistance checklist
in `.github/pull_request_template.md`: logic verified, no hallucinated APIs,
no unvetted dependencies, no leaked context or secrets.
