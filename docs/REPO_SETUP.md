# Repository Setup & Governance Record

**Repository:** `cedric-kiama-wachira/agentic_ai_engineering_with_go`
**Visibility:** Public · **License:** Apache-2.0
**Branching model:** Git Flow
**Record prepared:** 2026-06-05 · **Last updated:** 2026-06-14
**Prepared by:** Cedric Kiama Wachira (repository architect)
**Status:** Bootstrap + OpenSSF Phases 0–3 complete; Phase 4 in progress (deliverables (a), (c) landed) — pending migration to air-gapped GitHub Enterprise Server (GHES)

> **Auditor note.** This document records the controls configured during the
> bootstrap phase, the evidence that each was verified on live artifacts, and
> the issues detected and remediated along the way. The rule values in Section 5
> reflect the intended configuration; before formal submission, confirm them
> against the live rulesets at `Settings → Rules → Rulesets`.

---

## 1. Purpose

This repository is the bootstrap foundation for an Agentic AI engineering
initiative implemented in Go. It is built so that governance controls enforce
themselves — secure, reviewed, signed, spec-driven development — rather than
relying on individual discipline. This record is the evidence package for that
posture, suitable for audit review.

---

## 2. Repository Overview

| Item | Value |
|------|-------|
| Working directory | `~/Agentic_AI_Lab/GO_Labs/lab_1` |
| Default branch | `develop` |
| Protected branches | `main`, `develop` |
| Go module | `github.com/cedric-kiama-wachira/agentic_ai_engineering_with_go` |
| Go version | 1.26.4 (pinned in CI with `GOTOOLCHAIN=local` + CI guard; stated in README prerequisites) |
| Source layout | `cmd/agent/main.go`, `cmd/agent/main_test.go` |
| Governance files | `README.md`, `CONTRIBUTING.md`, `LICENSE`, `SECURITY.md`, `AGENTS.md`, `.github/CODEOWNERS`, `.github/pull_request_template.md`, `.github/dependabot.yml`, `docs/REPO_SETUP.md`, `docs/THREAT_MODEL.md`, `docs/AGENT_RUNTIME_DESIGN.md`, `docs/TOOLCHAIN_FRESHNESS.md`, `.github/security-insights.yml` |
| CI | `.github/workflows/ci.yml` — four jobs (Section 6) |
| Security tooling | `govulncheck`, `gosec`, `staticcheck` via the Go `tool` directive (pinned in `go.sum`) |
| Dependency updates | Dependabot — `gomod` + `github-actions`, weekly, governed PRs (Section 15.4) |
| Secret scanning | GitGuardian (GitHub App) |

---

## 3. Cryptographic Key Architecture

Three distinct key-roles are used. A deliberate decision was made to use
**separate keys for transport and signing**. This was reinforced by a hard
constraint discovered during setup (see Section 10, incident #1): GitHub enforces that
any single public key is globally unique to one role across the platform — a
repository deploy key cannot also be registered as an account key. Separating
keys also follows least-privilege; transport (push) and identity (signing) are
independent concerns.

| Role | Key | Algorithm | Registered where | Purpose |
|------|-----|-----------|------------------|---------|
| **Transport / deploy** | `id_dev_agentic_ai_lab_ed25519` | Ed25519 | Repo **Deploy key** (write access); referenced in local `~/.ssh/config` under Host alias `dev_agentic_ai` | Authenticates `git push`/`pull` |
| **Commit signing (local)** | `id_signing_agentic_ai_lab_ed25519` | Ed25519 | `git config --local user.signingkey` + `~/.ssh/allowed_signers` | Signs commits; enables local signature verification |
| **Commit signing (account)** | same public key as signing key above | Ed25519 | Account **SSH and GPG keys → Signing key** — `SHA256:QlTAWcCOBrHWsxpOeqJQhQS4gR5z8QLac1uZObv4m/w` | Enables GitHub-side "Verified" badge |

**Notes for audit:**
- The SSH key was generated with `ssh-keygen -t ed25519`; an `-b 4096` flag used
  initially is silently ignored for Ed25519 (key size is fixed at 256-bit) — the
  key is valid and strong.
- Signing keys are **not** placed in `~/.ssh/config` — signing performs no
  network connection and therefore has no SSH Host entry. This is intentional.
- Git identity (`user.name`, `user.email`) is set **locally** (`--local`) to
  avoid a stray global identity mis-attributing commits.
- Merge commits created via the GitHub web UI are signed by GitHub's web-flow
  key and display as "Verified", satisfying the signed-commit rule. Local
  `git log --show-signature` on such commits reports "cannot check signature —
  no public key" because GitHub's key is not in the local keyring; this is
  cosmetic, not a verification failure.

---

## 4. Branching Model (Git Flow)

Direct commits to `main` and `develop` are prohibited by ruleset (Section 5). All
changes enter through reviewed pull requests.

| Branch | Branches from | Merges into | Purpose |
|--------|---------------|-------------|---------|
| `main` | — | — | Production-ready, tagged releases |
| `develop` | `main` | `main` (via PR) | Integration branch (repository default) |
| `feature/*` | `develop` | `develop` (via PR) | New work |
| `bugfix/*` | `develop` | `develop` (via PR) | Non-urgent fixes |
| `hotfix/*` | `main` | `main` + `develop` | Urgent production fixes |

Branch naming (`<type>/<ticket-id>-<short-kebab-description>`) and Conventional
Commits are documented in `CONTRIBUTING.md`. A `chore/*` prefix was used on
PR #22 for repo-plumbing work — recorded as a conscious deviation on that PR;
the branch-naming table amendment lands via governed PR (see Section 16.5).

---

## 5. Branch Protection Rulesets

Two **rulesets** (GitHub's current branch-protection mechanism, chosen over
legacy branch-protection rules for organisation-wide promotability, bypass
audit trails, and Evaluate mode) are active, each targeting one branch.

### 5.1 `protect-main` (targets `main`)

| Rule | Setting |
|------|---------|
| Restrict deletions | Enabled |
| Block force pushes | Enabled |
| Require linear history | Enabled |
| Require pull request before merging | Enabled — **2** approvals, dismiss stale approvals on new commits, require Code Owner review |
| Require signed commits | Enabled |
| Require status checks to pass | Enabled — `build / vet / test`, `security / govulncheck`, `security / gosec`, `quality / staticcheck`, `GitGuardian Security Checks`; require branches up to date before merging |
| Bypass list | **Empty** |

### 5.2 `protect-develop` (targets `develop`)

| Rule | Setting |
|------|---------|
| Restrict deletions | Enabled |
| Block force pushes | Enabled |
| Require linear history | **Disabled** (merge commits permitted on the integration branch) |
| Require pull request before merging | Enabled — **1** approval, dismiss stale approvals on new commits |
| Require signed commits | Enabled |
| Require status checks to pass | Enabled — `build / vet / test`, `security / govulncheck`, `security / gosec`, `quality / staticcheck`, `GitGuardian Security Checks`; require branches up to date before merging |
| Bypass list | **Empty** |

**Design rationale:** `main` is held strictly (two-person review, linear
history) as the release branch; `develop` is slightly relaxed (single approval,
merge commits allowed) as the high-traffic integration branch. The empty bypass
lists mean any override is performed deliberately and is logged as a recorded
rule bypass — auditable rather than silent.

> **Enforcement caveat (critical).** On a personal GitHub Free account, rulesets
> are enforced on **public** repositories only; on a **private** repository the
> configuration is retained but **enforcement is silently disabled**. The repo
> is therefore kept **public** for the bootstrap phase. This dependency is the
> primary driver for migration to GHES (Section 11). See incident #5 in Section 10.

---

## 6. Continuous Integration

**File:** `.github/workflows/ci.yml` — workflow `CI`, four jobs (all required
checks; Section 5).

| Property | Value |
|----------|-------|
| Triggers | `pull_request` and `push` to `develop`, `main` |
| Token permissions | `contents: read` (least privilege) |
| Runner | `ubuntu-latest` (GitHub-hosted) |
| Go version | `1.26.4` (pinned; workflow-level `GOTOOLCHAIN: local` on all jobs) |
| Actions | SHA-pinned with `# vX.Y.Z` annotations (`actions/checkout`, `actions/setup-go`) — Section 15.2 |

| Job (required check name) | Purpose |
|---------------------------|---------|
| `build / vet / test` | Modules tidy → toolchain guard → `go build` → `go vet` → `go test -race` |
| `security / govulncheck` | Known-vulnerability scanning (reachability-scoped), `go tool govulncheck ./...` |
| `security / gosec` | SAST; strict `#nosec` policy (rule ID + justification required) |
| `quality / staticcheck` | Correctness gate; community-default `staticcheck.conf` at repo root |

The module-tidiness step is written to tolerate the absence of `go.sum` (a
bootstrap condition; `go.sum` now exists and is guarded — see Section 10,
incident #2). The **"Verify pinned Go toolchain"** step asserts no `toolchain`
directive exists in `go.mod` and the `go` directive is exactly `1.26.4`
(Section 15.3). The race detector is enabled because the agentic runtime is
concurrency-heavy. All four jobs are **required** status checks on both
protected branches (Section 5). All security tooling is invoked via the Go
`tool` directive — versions pinned in `go.sum`, no marketplace SAST actions —
for air-gap portability.

---

## 7. Secret Scanning

| Property | Value |
|----------|-------|
| Provider | GitGuardian (GitHub App) |
| Required check name | `GitGuardian Security Checks` |
| Status | Required on `main` and `develop`; observed passing ("No secrets detected") |

**Open item for audit:** confirm and record the installation scope of the
GitGuardian app (repository-level vs organisation-level). A stray free-text
"GitGuardian / Any source" required check was added in error during setup and
removed (see Section 10, incident #4); only the app-reported `GitGuardian Security
Checks` is required.

---

## 8. Code Ownership & Contribution Governance

| Artifact | Function |
|----------|----------|
| `.github/CODEOWNERS` | Auto-requests owner review on matching paths; `.github/`, `go.mod`, `go.sum`, `AGENTS.md` guarded explicitly. Validated by GitHub ("CODEOWNERS file is valid"). |
| `CONTRIBUTING.md` | Git Flow model, branch-naming convention, Conventional Commits, PR flow, required approvals, secure-coding standard, Dependabot PR review procedure. |
| `.github/pull_request_template.md` | Auto-populated checklist on every PR (branch origin, signing, tests, no-secrets, AI Assistance declaration + reviewer attestation). Verified auto-filling on PRs #2, #4, and #23 (post-AI-checklist). |
| `AGENTS.md` | Instructions governing AI coding assistants contributing to the repo (OpenSSF-derived; Section 16.1). |
| `README.md` | Project intro, prerequisites (Go 1.26.4), getting-started (canonical clone URL, not the local Host alias). |
| `SECURITY.md` | Coordinated-disclosure policy + intake channels (Section 14.2). |
| `LICENSE` | Apache-2.0 (explicit patent grant, suited to multi-contributor + corporate use). |

---

## 9. Bootstrap Validation Evidence

Each control was proven on a live pull request, not merely configured. Author /
approver / merger roles are distinct in the record.

| PR | Change | Approved by | Merge commit | Outcome |
|----|--------|-------------|--------------|---------|
| #1 | Governance docs (CODEOWNERS, CONTRIBUTING, PR template) | `digital-factory-dm` | `cb8f769` | Merged to `develop` |
| #2 | Go module, agent entrypoint, CI workflow | `digital-factory-dm` | `1b659e7` | Merged to `develop` (after CI fix, incident #2) |
| #3 | Trivial change to verify required checks gate merges | — | — (closed, not merged) | **Verification only** — both required checks reported green & marked Required; review correctly still blocked merge |
| #4 | Audit record (`docs/REPO_SETUP.md`) + README Go-version update | `digital-factory-dm` | `5d9a1fd` | Merged to `develop` (during which incident #5 was detected & remediated) |

**Repository history (develop):**

```
*   5d9a1fd  Merge PR #4  — audit record + remediation log
|\
| * 1e434f4  docs: log ruleset-enforcement remediation in change log
| * 2473a38  docs: add repository setup and governance record for audit
|/
*   1b659e7  Merge PR #2  — Go module, agent entrypoint, CI workflow
|\
| * a4554df  ci: make module-tidy check robust when go.sum is absent
| * ccbfb52  feat: scaffold Go module, agent entrypoint, and CI workflow
|/
*   cb8f769  Merge PR #1  — CODEOWNERS, contributing guide, PR template
|\
| * e1ad152  docs: add CODEOWNERS, contributing guide, and PR template
|/
* 81214ad  (main) chore: initialize repository with docs, license, and gitignore
```

**Other verified behaviours:**
- A deliberate direct push to `develop` was **rejected** (`GH013` — "Changes must
  be made through a pull request" and "Commits must have verified signatures"),
  confirming the PR-only and signed-commit rules.
- Commit signatures verify end-to-end (local "Good signature"; GitHub "Verified"
  badge) using the dedicated account signing key.
- The required-status-checks gate was confirmed to block merges until both
  checks report green, with no check stuck "pending" (PRs #3 and #4).

---

## 10. Incidents, Detections & Remediations (Bootstrap)

Issues caught during setup and how each was resolved. These demonstrate a
functioning detect-and-correct process, not an absence of problems.

| # | Detection | Root cause | Remediation |
|---|-----------|------------|-------------|
| 1 | "Key is already in use" when registering the signing key | Attempted to reuse the deploy key as an account signing key; GitHub keys are globally unique to one role | Generated a **dedicated** Ed25519 signing key; registered separately as an account Signing key (Section 3) |
| 2 | CI `build / vet / test` failed in ~15s: `fatal: go.sum: no such path in the working tree` | `git diff --exit-code go.mod go.sum` errors when `go.sum` does not yet exist (zero dependencies) | Rewrote the tidy step to use `git status --porcelain` over existing paths; self-arms once `go.sum` appears |
| 3 | Stray compiled binary `agent` in repo root after `go build ./...` | `go build ./...` emits the binary into the working directory | Removed the binary; hardened `.gitignore` (`/agent`, `coverage.out`); build artifacts excluded |
| 4 | Required-checks list contained a non-functional "GitGuardian / Any source" entry | Free-text check name added in error; nothing reports a check by that literal name → would block all merges as permanently pending | Removed the stray entry; required only the real `GitGuardian Security Checks` |
| 5 | **Ruleset enforcement silently disabled** (rulesets showed grey/disabled; PR review became optional) | Repository visibility was set to **private** on a personal Free account, where rulesets are not enforced on private repos | Reverted visibility to **public**; enforcement immediately restored. **Permanent remediation: migration to air-gapped GHES** (Section 11), where private + enforced is the default |

**Process note:** incident #5 is the most significant — a control that silently
stops enforcing yields false confidence. The lesson adopted: enforcement must be
**verified on a live PR** (as done in PRs #3 and #4), not assumed from
configuration.

---

## 11. Known Interim Deviations & Remediation Plan

Bootstrap-phase compromises, recorded transparently. None are intended to
persist into team operation; all are resolved by the GHES migration.

| # | Deviation | Why acceptable now | Remediation at GHES migration |
|---|-----------|--------------------|-------------------------------|
| 1 | **Second account (`digital-factory-dm`) used as approving reviewer** on PRs #1, #2, #4 and all subsequent governed PRs | Exercised an authentic two-person-review event during bootstrap | **Not genuine separation of duties** — both accounts controlled by one person. Superseded by independent human reviewers via enterprise IdP and teams. (Also confirm work-identity usage complies with applicable identity policy.) |
| 2 | **Shared deploy key used for human push access** | Single operator during bootstrap | Deploy keys are repo-scoped and not person-attributable. Retire for human use; each developer authenticates with their own account auth + signing keys. Reserve deploy keys for CI/CD or server deploys. |
| 3 | **Repository on a personal account and kept public** for ruleset enforcement | Only configuration under which Free-plan rulesets enforce | Migrate to **air-gapped GHES**: private + enforced is the default; no public exposure. |
| 4 | **Interim email security contact (`theemail@mail.com`) in `SECURITY.md`** | Single operator during bootstrap; PVR is the preferred channel regardless | Replace with a monitored organizational security mailbox at the GHES migration (see Section 14.2). |

### GHES (air-gapped) migration tasks — items that do not port unchanged

1. **CI runners:** replace `runs-on: ubuntu-latest` with **self-hosted runners**
   registered inside the perimeter. `actions/checkout` and
   `actions/setup-go` are pulled from the public marketplace — resolve via
   **Actions sync** (mirrored into the instance) or vendored copies.
2. **Secret scanning:** the cloud GitGuardian app cannot reach its SaaS in an
   air-gapped network. Replace with **GHES Advanced Security secret scanning +
   push protection**, or a self-hosted GitGuardian deployment.
3. **Org-level governance & identity:** promote `protect-main` / `protect-develop`
   to **organisation-level rulesets** so all repos inherit them; wire identity to
   the **enterprise IdP** (SAML/LDAP) to enable genuine separation of duties and
   per-developer signing — retiring deviations #1 and #2 above.
4. **Vulnerability intake:** GitHub Private Vulnerability Reporting is a
   GitHub.com feature; on air-gapped GHES use the GHES security-advisory
   mechanism or the enterprise vulnerability-intake process (see Section 14.4).
5. **Dependabot:** requires GitHub Connect or a self-hosted setup on GHES
   (Section 15.4).
6. **govulncheck vulnerability database:** air-gapped operation requires a
   mirrored/local copy of the Go vuln DB on a defined refresh cadence. **A stale
   DB yields a false-green gate** — the same false-confidence pattern as a
   silently-disabled ruleset; treat DB freshness as a monitored control
   (cross-ref `docs/THREAT_MODEL.md` §7).

---

## 12. Recommended Next-Phase Enhancements

Status of the original backlog, updated as phases complete.

- GHES migration (resolves all Section 11 deviations and the Section 5 enforcement caveat). — **Open**
- ~~CodeQL code scanning (SAST) as an additional required check.~~ — **Superseded:** SAST delivered via Go-native `gosec` + `staticcheck` (Phase 1, Section 15.1); CodeQL deliberately not adopted (marketplace/cloud dependency vs air-gap portability).
- ~~Dependabot for dependency and action updates.~~ — **Done** (Phase 1, Section 15.4).
- ~~Pin GitHub Action versions to commit SHAs.~~ — **Done** (Phase 1, Section 15.2).
- Strict Code-Owner-specific approval enforced on `main`. — **Open**
- `release/*` branch flow with semantic version tagging. — **Open** (Phase 3)
- Branch auto-deletion on merge. — **Open** (manual deletion practiced consistently)
- `CODEOWNERS` entry for `/docs/` so future edits to this record require owner review. — **Open**
- Scheduled toolchain-freshness watcher (e.g. scheduled govulncheck run): the CI guard prevents drift but does not detect staleness — gap detected 2026-06-12 when 18 stdlib advisories against go1.26.1 surfaced incidentally. — **Addressed (Phase 4 (c))** via a human-executed compensating control, `docs/TOOLCHAIN_FRESHNESS.md` (Section 18.1); full scheduled-CI conversion deferred to GHES (age becomes primary on the mirror).
- Enforced `si-validate` CI gate for `.github/security-insights.yml` (`cue` in the repo `tool` directive + vendored schema). — **Deferred** (Section 18.4): lands with the SBOM-style enforcement plumbing or at GHES migration, whichever first; until then validation is the local cue-vet recipe (Section 18.3).

---

## 13. Verification Checklist (for re-audit)

- [ ] Both rulesets `protect-main` and `protect-develop` show **Active** (green) — not disabled.
- [ ] Repository visibility supports enforcement (public on Free plan, or hosted on GHES/Team).
- [ ] Bypass lists are empty on both rulesets (Dependabot NOT on either bypass list).
- [ ] Required status checks on both: `build / vet / test`, `security / govulncheck`, `security / gosec`, `quality / staticcheck`, `supply-chain / sbom-drift`, `GitGuardian Security Checks`.
- [ ] No stray free-text required checks (e.g. "GitGuardian / Any source"); no check stuck "Expected — waiting".
- [ ] `protect-main` requires 2 approvals; `protect-develop` requires 1 (confirm live).
- [ ] Signing key present on account as type **Signing**; deploy key present on repo.
- [ ] CI workflow token permission is `contents: read`; all `uses:` lines SHA-pinned with `# vX.Y.Z` comments.
- [ ] `go.mod` has NO `toolchain` directive; `go` directive is exactly `1.26.4`; workflow sets `GOTOOLCHAIN: local`.
- [ ] GitGuardian installation scope confirmed and recorded (Section 7).
- [ ] A test PR confirms checks gate merges with nothing stuck "pending".
- [ ] `SECURITY.md` present at repo root; Private Vulnerability Reporting enabled (Section 14).
- [ ] `AGENTS.md` present at root with CODEOWNERS guard; PR template carries the AI Assistance section (Section 16).
- [ ] `.github/security-insights.yml` present and validates against SI schema v2.2.0 via the local cue-vet recipe (Section 18.3).

---

## 14. OpenSSF Hardening — Phase 0 (Baseline & Security Policy)

This section records the first phase of the OpenSSF Foundation Hardening Plan
(`OPENSSF_FOUNDATION_HARDENING_PLAN.md`), executed through the governed PR flow.

### 14.1 Scorecard baseline

OpenSSF Scorecard **v5.4.0** was run against the live repository to establish a
measurable security-posture baseline before any hardening changes.

| Run | Commit scanned | Aggregate | Notes |
|-----|----------------|-----------|-------|
| Baseline (pre-policy) | `11fc900` (develop tip before PR #6) | **5.0 / 10** | Captured from terminal output; Security-Policy 0/10 |
| Phase 0 exit (post-policy) | `6bfc061` (PR #6 merge) | **5.6 / 10** | Machine-readable evidence: `docs/scorecard-phase0.json` |

The full check breakdown is preserved in `docs/scorecard-phase0.json` (scanned
at commit `6bfc061`, 2026-06-06). Scores of note at baseline:

- **10/10:** Binary-Artifacts, Dangerous-Workflow, License, Token-Permissions,
  Vulnerabilities — validating the existing bootstrap controls (least-privilege
  CI token, no committed binaries, no dangerous workflow patterns).
- **5/10 Branch-Protection:** the three `Warn` items (1 approval on `develop`,
  CODEOWNERS review not required on `develop`, last-push-approval disabled) are
  **deliberate** — the governance model holds `main` strict (2 approvals, linear
  history, CODEOWNERS) and `develop` relaxed as the integration branch. Scorecard
  scores against a maximalist single-branch ideal that does not model this split.
  Not treated as a defect; documented as an accepted, auditable design choice.
- **0/10 or `?` (planned backlog, not defects):** Security-Policy (closed this
  phase), SAST, Pinned-Dependencies, Dependency-Update-Tool, Fuzzing,
  Signed-Releases, Packaging — each scheduled in Phases 1–3 of the hardening plan.
- **Maintained 0/10** (repo <90 days) and **Contributors 0/10** (single-operator
  bootstrap) resolve with time and the team/GHES migration respectively.

### 14.2 Security policy (Security-Policy 0 → 10)

A coordinated-disclosure `SECURITY.md` was added at repository root (PR #6,
commit `5c5838d`, signed and verified). Scorecard confirmed all four sub-signals:
policy file detected, linked content, disclosure/timelines present, and policy
text present.

The policy advertises two intake channels:
1. **GitHub Private Vulnerability Reporting** — enabled at the repository level
   (Settings → Advanced Security). Verified live: the Security → Advisories
   triage queue is operational. *Note: this is a repository setting, not a
   committed artifact, so it is recorded here rather than carried by a PR.*
2. **Interim email contact** (`theemail@mail.com`) — a bootstrap-phase
   stand-in, to be replaced by a monitored organizational security mailbox at
   the GHES migration. Recorded as interim deviation #4 in Section 11.

### 14.3 Scope note — DAST gap

Scorecard measures repository/supply-chain posture and does not perform dynamic
analysis (DAST). LFD121's verification model calls for both static and dynamic
analysis. Static analysis is addressed in Phase 1 (`govulncheck`, `gosec`);
dynamic testing of the agent runtime (untrusted-content ingestion, tool-call
surface) is recorded here as an open gap to be addressed once an agent runtime
exists, and folded into the threat model (Phase 1).

### 14.4 Air-gapped portability note

GitHub Private Vulnerability Reporting is a GitHub.com feature. On air-gapped
GHES the equivalent is GHES's own security-advisory mechanism or the enterprise
vulnerability-intake process. Added to the set of items (alongside cloud
GitGuardian, hosted runners, and — for later phases — cosign keyless / SLSA
public infrastructure) that do not port unchanged to the air-gapped target.

---

## 15. OpenSSF Hardening — Phase 1 (Secure-Development Gates & Supply Chain)

> **Recording note (honesty).** Phase 1 completed on 2026-06-07 but this record
> was not updated at the time — the omission was detected on 2026-06-12 during
> the Phase 2 closeout and is corrected retroactively here. All evidence cited
> (PRs, commits, Scorecard JSON) is contemporaneous and verifiable in the repo
> history; only this narrative is late. Detection-and-correction logged in the
> Change Log, consistent with the Section 10 process discipline.

All Phase 1 controls landed via the governed PR flow and use the **Go `tool`
directive** (tool versions pinned in `go.sum`) rather than marketplace actions —
a deliberate air-gap-portability decision.

### 15.1 Security & quality gates (PRs #8, #10, #12)

| Gate (required check) | Tool & version | Invocation | Notes |
|-----------------------|----------------|------------|-------|
| `security / govulncheck` | govulncheck v1.3.0 | `go tool govulncheck ./...` | Reachability-scoped known-vuln scanning; separate named job for audit granularity. Clean at adoption. |
| `security / gosec` | gosec v2.27.1 | `go tool gosec -nosec-require-rules -nosec-require-justification ./...` | Strict: every `#nosec` requires a rule ID + written justification. 0 issues at adoption. |
| `quality / staticcheck` | staticcheck v0.7.0 (2026.1) | `go tool staticcheck ./...` | Community-default `staticcheck.conf` at repo root (all SA/S/U checks; 6 subjective ST style checks excluded). Named `quality /` — it is a correctness gate, not a security gate. |

Each check was landed via PR, allowed to report green once, then added to
**both** rulesets via dropdown autocomplete (never free-typed — the incident #4
lesson), then proven to gate via a throwaway PR closed unmerged (PRs #9, #11,
#13).

### 15.2 Actions SHA-pinning (PR #14)

All 8 `uses:` lines across the 4 CI jobs pinned to full commit SHAs with
`# vX.Y.Z` annotations (e.g. `actions/checkout@34e11487… # v4.3.1`,
`actions/setup-go@40f1582b… # v5.6.0`). SHAs fetched authoritatively via
`git ls-remote` (lightweight tags → the bare ref IS the commit SHA; annotated
tags require the peeled `^{}` SHA). Closes the repointable-moving-tag
supply-chain hole. Scorecard Pinned-Dependencies → 10/10.

### 15.3 Pinned-Go toolchain guard (PR #15)

Workflow-level `env: GOTOOLCHAIN: local` on all 4 jobs, plus a named
`build / vet / test` step **"Verify pinned Go toolchain"** asserting (a) no
`^toolchain ` line exists in `go.mod`, and (b) the `go` directive is exactly
`go 1.26.1`. Rationale: a dependency bump can inject a `toolchain` line or bump
the `go` directive as collateral; `go mod tidy` alone is insufficient. A
Dependabot `ignore` rule for go/toolchain was evaluated and **rejected as
unverifiable**; the CI guard is the verified control. **Verified both ways on a
live artifact:** deliberate injection of `toolchain go1.27.0` → CI red at the
guard; reverted → green.

### 15.4 Dependabot, governed (PRs #16, #19; live-bot verification on #17/#18)

- `.github/dependabot.yml`: ecosystems `gomod` + `github-actions`, weekly,
  `chore` prefix, minor/patch grouped per ecosystem, open-PR limit 5, **no
  auto-merge**, merge-commit only.
- **No ruleset bypass:** Dependabot PRs face the identical gate — five required
  checks + independent review (bypass lists verified empty).
- `version-update:semver-major` ignore applied to **github-actions only**
  (PR #19); **gomod majors deliberately still surface** (a security-tool major
  can carry detection improvements).
- Verified end-to-end on live bot PR #18: Verified commit signature, `chore:`
  prefix, 5/5 checks green, diff preserved the SHA + `# vX.Y.Z` pin convention.
  Both major-bump PRs (#17, #18) closed with documented defer reasons.
- Unsigned-bot-commit recovery procedure documented in `CONTRIBUTING.md`
  (recreate once on the open PR, else author a signed human PR; the signing
  rule is never weakened), reconciled with GitHub's 2026-01-27 removal of
  Dependabot comment commands.

### 15.5 Threat model (PR #20)

`docs/THREAT_MODEL.md` v1.0 — Phase-1 **lightweight**, deliberately scoped to
what exists (governed repo + CI/CD supply chain); the agent-runtime model is
deferred until runtime code exists, with STRIDE/DFD rigor landing in the
Phase 4 self-assessment. Records: trust boundaries (Dependabot as a second
author under the identical boundary), AI-SDK transitive dependencies verified
inert via `go mod why`, the toolchain-guard rationale, Dependabot privilege
posture, honest verification-coverage gaps (no DAST, no fuzzing), and GHES
migration deltas including the stale-vuln-DB false-green failure mode.

### 15.6 Phase 1 checkpoint — Scorecard 5.6 → 7.1

Movers: **Pinned-Dependencies → 10/10**, **Dependency-Update-Tool → 10/10**,
CI-Tests 9/10. Non-movers explained and documented:

- **SAST 0/10 is a Scorecard detection blind spot, not a control gap** —
  Scorecard pattern-matches marketplace SAST actions/CodeQL and cannot see
  `go tool gosec` / `go tool staticcheck` invocations. The gates are real and
  required; the metric is not chased at the cost of air-gap portability.
- Branch-Protection 5/10 (deliberate develop-relaxed design, Section 14.1);
  Maintained 0/10 (repo age); Contributors/CII/Fuzzing 0 (single-operator /
  future); Packaging & Signed-Releases `?` (no releases yet — Phase 3).

---

## 16. OpenSSF Hardening — Phase 2 (AI Guardrails, LFEL1012)

Phase 2 executed 2026-06-10/12 via three governed PRs (one artifact per PR,
ratified at session start), plus the interim checkpoint. Decisions ratified:
**(1)** lethal-trifecta isolation lands as a design document only — no
enforcement scaffolding until runtime code exists; **(2)** one artifact per PR.

### 16.1 `AGENTS.md` — AI assistant instructions (PR #21, merge `e469e35`)

Tool-agnostic instruction file at repo root, derived from the OpenSSF
*Security-Focused Guide for AI Code Assistant Instructions*: never commit
secrets; validate input at boundaries (allowlists); tests required for new
code paths; no unjustified dependencies; flag uncertainty rather than
fabricate; error-handling discipline; repo-specific constraints (pinned
toolchain, `tool` directive over marketplace actions, auditable suppressions);
explicit prohibited-actions list. Same PR added an explicit `/AGENTS.md`
CODEOWNERS entry — declarative (the `*` rule already covers it), marking the
path governance-sensitive, same treatment as `/go.mod`. A "Review expectation"
section was deliberately stripped because the checklist it references did not
yet exist (restored in PR #22 — no merged doc carries a dangling reference).

### 16.2 PR-template AI Assistance checklist (PR #22, merge `2e0ab0f`)

`.github/pull_request_template.md` gains an **AI Assistance** section: every PR
makes an explicit AI-assisted-or-not declaration (the section cannot be
silently skipped), and AI-assisted changes carry four **reviewer**
attestations — logic verified, no hallucinated APIs, no unvetted dependencies,
no leaked context/secrets. Same PR restored the `AGENTS.md` "Review
expectation" cross-reference. Source branch used a `chore/*` prefix — recorded
as a conscious deviation in a PR #22 comment (see 16.5).

### 16.3 Agent runtime isolation design (PR #23, merge `002d02e`)

`docs/AGENT_RUNTIME_DESIGN.md` v1.0 — the lethal-trifecta isolation design for
the future runtime: break-one-leg principle; separated untrusted-content
handlers / private-data accessors / single default-deny egress gateway;
per-tool scoped credentials; no standing secrets; secrets never in model
context; sandboxing requirements. **Status DESIGN** — each principle carries an
*Implementation gate* stating when it converts to an enforced, verified
control; the document explicitly disclaims being enforced today (treating
design as control before code exists would manufacture false confidence).
Two-line back-link added to `docs/THREAT_MODEL.md` §2, preserving that
section's promise that threat-level modeling still lands there once runtime
code exists.

### 16.4 Verification on live artifacts

- **PR #23 was the first PR opened against the new template** — the AI
  Assistance section rendered, the author declared AI-assisted, and the
  reviewing account ticked the four reviewer attestations before approving:
  the checklist mechanism exercised end-to-end on first use. (Caveat recorded:
  the attestation was made by the second account — the standing
  separation-of-duties deviation, Section 11 #1; the *mechanism* is what is
  verified here, genuine independent attestation arrives with real reviewers
  at GHES.)
- Template-renders-from-base-branch behavior confirmed: PR #22 itself rendered
  the old template; #23 the new one — expected GitHub behavior, documented so
  the sequencing reads correctly in the audit trail.

### 16.5 Phase 2 checkpoint — Scorecard 7.1 → 7.2 — and follow-ups

Interim Scorecard (same pinned v5.4.0 instrument; evidence
`docs/scorecard-phase2-interim.json`): **7.2 / 10**. Single mover:
**CI-Tests 9 → 10/10** ("15 out of 15 merged PRs checked") — the Phase 2 PRs
all carried green checks. No other movement, as predicted: Phase 2 artifacts
are documentation/template controls, which Scorecard does not score. Trend:
**5.0 → 5.6 → 7.1 → 7.2**. The formal re-measure remains Phase 4.

Open follow-ups from Phase 2: **(a)** amend the `CONTRIBUTING.md` branch-naming
table to include `chore/*` (deviation recorded on PR #22; lands as closeout
PR 2); **(b)** Phase 3 (artifact integrity) carries a flagged design gap —
cosign keyless + slsa-github-generator depend on public Fulcio/Rekor/OIDC
infrastructure, incompatible with air-gapped GHES; Phase 3 requires a
materially different design for the target environment, and cosign ≥ 2.6.2
(CVE-2026-22703) when adopted.

---

## 17. OpenSSF Hardening — Phase 3 (Artifact Integrity, scoped)

Phase 3 executed 2026-06-12. The hardening plan's original prescription
(cosign keyless + slsa-github-generator) was **superseded before
implementation**: both depend on public Fulcio/Rekor/GitHub OIDC, none of
which exist in air-gapped GHES — building them would create controls that die
at migration. Four options were analyzed; **option 3 (SBOM-only now, signing
deferred) was ratified**, cross-checked against two parallel architect
sessions and reconciled independently. The signing architecture decision
(key-based cosign >= 2.6.2 vs private Sigstore) remains an open item in the
GHES migration backlog.

### 17.1 SBOM tooling (PR #26, merge `f174f54`)

`cyclonedx-gomod` v1.10.0 added via the Go `tool` directive (pinned in
`go.sum`, no marketplace action) — tag verified via `git ls-remote`. A first
`go get -tool` attempt targeted the module root (not a `main` package), left
a half-applied go.mod state, and was fully reverted before the correct
`cmd/cyclonedx-gomod` path was applied; the merged diff reflects one clean
operation. Tool execution verified: reported version, Go version, and a
ModuleSum matching `go.sum`.

### 17.2 Detour 1 — toolchain patch bump (PR #30, merge `4370ca9`)

During PR #26 verification, govulncheck surfaced **18 informational stdlib
advisories, all `stdlib@go1.26.1`** (0 reachable at symbol level) — the
pinned toolchain had aged behind three patch releases. Bumped
**go 1.26.1 -> 1.26.4** in lockstep: go.mod directive, CI guard assertion,
four `setup-go` pins, local toolchain first (tarball SHA-256 verified against
go.dev metadata), doc sweep (Section 15.3 left as contemporaneous history).
Post-bump govulncheck: 18 -> 0. **Gap recorded (Section 12):** toolchain
freshness has no scheduled watcher — these advisories surfaced incidentally;
the CI guard prevents drift but does not detect staleness.

### 17.3 Detour 2 — Dependabot security wave (PRs #27–#29)

PR #26's dependency tree triggered 11 Dependabot alerts (2 high / 6 moderate
/ 3 low: go-git, go-billy, circl — all CI-time tool-tree exposure, none
reachable from application code; the go-git high was a commit
signature-verification flaw, directly relevant to this repo's integrity
model). PR #29 (go-git 5.19.1) was merged after Dependabot's automatic
rebase onto the post-#30 develop; its MVS resolution also satisfied #27
(go-billy 5.9.0) and #28 (circl 1.6.3), which Dependabot auto-closed.
All 11 alerts closed; verified on the alerts view at commit `b8376b5`.
Post-merge functional check: the SBOM tool builds and runs on the bumped
tree that upstream v1.10.0 never tested.

### 17.4 SBOM baseline (PR #31, merge `4a21e96`) and drift gate (PR #32, merge `2b88924`)

**Placement:** `sbom/bom.json` committed **in-repo**, generated by
`go tool cyclonedx-gomod mod -json -noserial -notimestamp -licenses` —
determinism proven by three byte-identical generations, flags verified on
the **live binary** (the project README is stale and omits
`mod -notimestamp`; one parallel review wrongly flagged the flag as
hallucinated from docs — the binary is the authority). Stated precisely:
the signed-commit chain proves **file provenance** (who committed it,
unaltered since), NOT **content correctness**. Correctness is enforced by
the **`supply-chain / sbom-drift`** CI gate: regenerate with the identical
invocation, normalize, diff, fail on mismatch — the SBOM's **consumer**
under the Phase 3 decision rule, which is what earned it required-check
status on both rulesets (sixth required check).

**Two gate defects caught before enforcement** (the land->green-once->require
sequence working as designed):

1. *Cross-commit churn:* the main component's version is a git-HEAD-derived
   pseudo-version that changes every commit — a raw diff fails on every PR
   forever. The PR #31 determinism proof ran at a single commit and could
   not surface this. Fix: normalize each file's own main pseudo-version
   (transitives remain drift-detectable; becomes a no-op once a release tag
   exists).
2. *Cross-environment churn:* `go tool` compiles the tool per-machine; the
   tool stamps its own binary hashes into `metadata.tools[].hashes`, which
   differ between workstation and runner. Caught by the first real CI run
   (failing while not yet required — zero cost). Fix: strip those hashes in
   normalization (tool identity retained).

**Gate proof (PR #33, closed unmerged):** a dependency change without SBOM
regeneration went RED at `supply-chain / sbom-drift` with the Required badge
visible; merge blocked. Bonus finding: the test's MVS cascade (gosec
2.27.1->2.26.1) broke `security / gosec` with a flag-parse error — the CI
invocation is coupled to the pinned tool version and **fails loud** on
uncoordinated downgrades rather than silently weakening the scan.

### 17.5 Deferrals (deliberate scoping, with consequences)

1. **SBOM signing deferred** (ratified option 3): until the signing
   architecture lands, `sbom/bom.json` is an **informational artifact** —
   provenance-bound via the signed-commit chain and freshness-verified by
   the drift gate, but **not a trusted attestation**. Recorded in
   `docs/THREAT_MODEL.md` Section 3.
2. **First-tag/release governance deferred** until a releasable artifact
   exists (`cmd/agent` is scaffolding; a tag now would be governance
   theater and would rush the first touch of `main`). The release
   workstream additionally requires an `app`/`bin` SBOM (build-resolved,
   per-artifact) — the committed **module** SBOM does NOT pre-satisfy
   release provenance.

### 17.6 Process deviation — AI-attestation attribution (PRs #26, #30)

The four AI Assistance reviewer attestations on PRs #26 and #30 were ticked
by the **author** at open time, not by the reviewing account via description
edit as Section 16.2 requires — self-attested, no independent attribution.
Detected in-session, recorded as deviation comments on both PRs. The
mechanism was restored and exercised correctly on PRs #31 and #32 (boxes
empty at open; ticked by `digital-factory-dm` via description edit before
approval, with the open->tick->approve sequence captured). (This extends the
Section 16.4 caveat: on these two PRs the mechanism itself was bypassed, not
merely choreographed.)

---

## 18. OpenSSF Hardening — Phase 4 (Security Insights & Self-Assessment) — IN PROGRESS

> **Status (mid-phase).** Phase 4 is **in progress**. Deliverables **(c)**
> toolchain-freshness runbook and **(a)** Security Insights file are COMPLETE and
> landed via governed PRs; **(b)** prose self-assessment and **(d)** formal
> Scorecard re-measure are PENDING. This section is written in-phase as each
> deliverable lands — per the in-phase audit rule (the Phase 1 omission, Section 15
> recording note, is the precedent not to repeat); (b) and (d) are appended here on
> completion. **Scorecard is NOT re-measured until (d)** — the trend
> (5.0 → 5.6 → 7.1 → 7.2 → 7.2, pinned v5.4.0) is unchanged by this phase so far.

Phase 4 delivers the project's machine- and human-readable security-posture
documentation: an OpenSSF Security Insights file, a prose self-assessment, the
operational toolchain-freshness procedure, and a formal Scorecard re-measure.

### 18.1 Deliverable (c) — Toolchain-freshness runbook (PR #36, merge `32ebd12`)

`docs/TOOLCHAIN_FRESHNESS.md` — **STATUS: ACTIVE OPERATOR PROCEDURE —
COMPENSATING CONTROL, HUMAN-EXECUTED** (not CI-enforced; converts to a scheduled
gate at GHES). Closes the **staleness** gap the Phase 1 pinned-toolchain guard
does not cover: the guard prevents *drift* (an unexpected `toolchain` line or a
changed `go` directive), not *staleness* (a pinned-but-aged toolchain behind
published security patches) — the exact condition that surfaced incidentally as
18 stdlib advisories in Phase 3 (Section 17.2). Addresses the Section 12
"scheduled toolchain-freshness watcher" backlog item.

Design (every gate cell verified on live `govulncheck` v1.3.0 runs 2026-06-14,
clean and forced-failure):
- Weekly run cadence (independent of the Go vuln DB's irregular curated update
  cadence).
- **Reachability gate keys on `config`-frame presence, not exit code** —
  `jq -rs '.[0].config'`: object = DB reached; `null`/error = no usable config
  frame → INCONCLUSIVE, gate fires.
- **Reachable-vuln signal = `finding`-frame count**, NOT `osv` frames (osv frames
  stream on every run — they mean "exists in DB," not "your code reaches it" —
  counting them would false-positive every run).
- Exit code is overloaded (vulns found OR DB unreachable; json-mode semantics
  differ) and is therefore NOT in the gate logic — captured for diagnosis only via
  de-pipelined `GV_EXIT=$?`.
- Toolchain assertion: config `go_version` must equal the `go.mod` directive
  (`go1.26.4`).
- **30-day DB-age bound, PROVISIONAL** — the Go vuln DB is curated with no
  published cadence (a 12-day-old DB is benign), so age is advisory on the
  canonical online source, never a gate. Calibration TODO open (calibrate against
  the observed `db_last_modified` series in `docs/freshness-runlog.md`).
- **GHES conversion:** at the air-gapped mirror, **age flips from advisory to
  PRIMARY** — the stale-but-valid-mirror false-green (frame present, old
  `db_last_modified`, exit 0) is invisible to frame-presence. Full conversion
  (scheduled CI on a self-hosted runner, `-db` → mirror, mirror-sync monitor) is
  recorded in the Section 11 GHES backlog, not duplicated in the runbook.
- **Observation run-log:** weekly observations append to a SEPARATE tracked file
  `docs/freshness-runlog.md` (signed/dated evidence the check ran — the teeth of a
  human-executed control). Created on FIRST WEEKLY USE, not by PR #36; the
  one-PR-per-week vs periodic-batch cadence is deferred to that file's own header.

### 18.2 Deliverable (a) — OpenSSF Security Insights file (PR #37, merge `fe269a2`)

`.github/security-insights.yml` — machine-readable, single-repository security
posture conforming to the **OpenSSF Security Insights schema v2.2.0**, for
ingestion by consumers such as Scorecard / CLOMonitor / LFX Insights.

**Filename correction.** The hardening plan
(`OPENSSF_FOUNDATION_HARDENING_PLAN.md` §7) and the rev-1 Phase 4 handoff named
this `SECURITY-INSIGHTS.yml` (uppercase); both are **stale**. The spec's filename
is lowercase `security-insights.yml`.

**Verification — every decision grounded in a live artifact, not memory:**
- **Schema v2.2.0 confirmed current** via the `ossf/security-insights` **tagged
  release** (not `main`; not the `ossf/security-insights-spec` draft repo). The
  authoritative `spec/schema.cue` is self-contained (only the stdlib `time`
  import — no cross-repo CUE imports).
- **Placement `.github/`** per the spec's own detection guidance (consumers probe
  repo root *or* the source-forge dir, e.g. `.github/`); `docs/` ruled OUT as a
  non-detected location despite house style. Matches ossf's own published file.
- **Scope: single-repository** (`header` + `repository`); the optional `project`
  section omitted by design (one-repo project).
- **`repository.security.tools` inventory — each field read from source:**

| Tool | `type` | `integration` (adhoc/ci/release) | `rulesets` source |
|------|--------|----------------------------------|-------------------|
| govulncheck | `SCA` | true / true / false | `["default"]`; `adhoc:true` = the weekly hand-run in `docs/TOOLCHAIN_FRESHNESS.md` |
| gosec | `SAST` | false / true / false | `["default"]` (default ruleset; CI `-nosec-*` flags govern `#nosec` handling only, not rule selection) |
| staticcheck | `other` | false / true / false | verbatim mirror of the `staticcheck.conf` `checks` directive |
| GitGuardian | `secret` | false / true / false | `["default"]` (schema no-customization value; App config not verifiable from repo) |
| cyclonedx-gomod | `other` | false / true / false | `["default"]` (SBOM generator, not a finding scanner) |

  `build / vet / test` is **deliberately excluded** from `tools` — it is a
  build/test gate, not a finding-producing scanner. `staticcheck` and
  `cyclonedx-gomod` are typed `other` (the honest enum members — there is no SBOM
  type, and labelling a correctness linter `SAST` would overstate a security
  control in a public file). Triggers read from `.github/workflows/ci.yml` (no
  `schedule`/release → `ci:true`, `release:false`); `license.expression`
  `Apache-2.0` read from `LICENSE`.

**Honesty framing (carried in-file and here, not softened):**
- GitGuardian `rulesets: ["default"]` is the schema's no-customization value but
  is **not verifiable from this repo** (App config not visible).
- The SBOM is **informational, NOT a trusted attestation** — signing deferred
  (Section 17.5); this file does not change that.
- `core-team` carries `name` + `primary` only — personal email omitted from a
  **public** file (signing-identity migration off personal Gmail remains a GHES
  backlog item).
- `staticcheck` `rulesets` is a frozen mirror of `staticcheck.conf` with **no
  automated SI-drift gate** (see 18.4); the human freshness procedure owns keeping
  them in sync, and the in-file comment names `staticcheck.conf` as canonical.

**Governed landing (discipline held end-to-end):** signed commit `8a0c8fa`
(Verified) → 6 required checks green → the four AI Assistance attestations ticked
by `digital-factory-dm` via description edit during review (the attribution
discipline held — contrast the Section 17.6 deviation) → approval after the final
commit → **merge commit `fe269a2`** (not rebase — signature chain intact) →
branch deleted → local synced. Post-merge re-vet of the file on `develop`: clean.

### 18.3 Validation toolchain & recipe (recorded so it is not tribal knowledge)

The SI file is validated **locally** against the authoritative schema before
commit. The recipe is recorded here as a real artifact (not session-bound memory)
because the enforced CI equivalent is deferred (18.4):

- **Tool:** `cue` **v0.16.1** (`cuelang.org/go/cmd/cue`), pinned via the Go `tool`
  directive in a **separate validation workspace** — NOT the repo `go.mod` (see
  18.4). Go-native and air-gap-portable like the rest of the toolchain; the CUE
  OCI module-registry path is dormant for loose-file vetting.
- **Authoritative schema:** `ossf/security-insights` `spec/schema.cue` at tag
  **v2.2.0**, vendored into the validation workspace.
- **Invocation (offline-proven):**
  `CUE_REGISTRY=bogus.invalid go tool cue vet -d '#SecurityInsights' <file> schema.cue`
- **Expected signals (this IS the definition of "working"):** clean = **EXIT 0,
  no output**; failure = **EXIT 1 with a constraint error naming the field**.
  Proven by a deliberate `status`-enum break → `repository.status: 8 errors in
  empty disjunction`, enumerating the valid enum. The teeth-test is the
  documentation.
- **Offline guarantee:** `CUE_REGISTRY` pointed at an unreachable host; vet passes
  with **zero `downloading` lines** → no network path touched. Air-gap-clean.

### 18.4 Deferral — enforced `si-validate` CI gate (with firing trigger)

Validation of the SI file is currently the **documented local procedure** (18.3),
NOT an enforced CI gate. An enforced **`si-validate`** required check (`cue` added
to the repo `tool` directive + a vendored `schema.cue` + a CI job + registration
as a 7th required check on both rulesets via the
land→green-once→dropdown→throwaway-proof sub-procedure) is **DEFERRED**.

- **Firing trigger (not open-ended):** lands **with the SBOM-style enforcement
  plumbing, or at the GHES migration, whichever is first** — it rides the shared
  `cue` + vendored-schema infrastructure those build anyway. Logged in the
  Section 11 GHES backlog and Section 12.
- **Rationale (change-rate asymmetry):** an enforced schema gate earns its keep in
  proportion to how often the gated artifact changes and how likely a change is
  malformed. The SI file is near-static and deliberately hand-edited (low churn,
  low malformation probability); the SBOM is the inverse (mutates on every
  dependency bump, often via automated PRs). The first required-check spend
  belongs on the high-churn artifact, not this one.
- **Residual (honest):** until the gate lands, the local recipe (18.3) is the
  control, and its **human-execution residual** is the gap — the same class as the
  (c) runbook's. Recorded, not hidden.

---
## Change Log

| Date | Author | Change |
|------|--------|--------|
| 2026-06-05 | Cedric Kiama Wachira | Initial bootstrap record. |
| 2026-06-05 | Cedric Kiama Wachira | Ruleset enforcement found disabled after repo set to private on personal Free plan; restored by reverting to public. Permanent remediation: migration to air-gapped GitHub Enterprise (scheduled). |
| 2026-06-05 | Cedric Kiama Wachira | Record completed post-bootstrap: added PR #4 evidence, Section 10 incidents/detections (key-role collision, CI go.sum guard, stray binary, stray GitGuardian check, enforcement-disabled incident), and Section 11 air-gapped GHES migration tasks. |
| 2026-06-06 | Cedric Kiama Wachira | OpenSSF Hardening Phase 0: established Scorecard v5.4.0 baseline (5.0 to 5.6); added coordinated-disclosure `SECURITY.md` (Security-Policy 0 to 10, PR #6) and enabled GitHub Private Vulnerability Reporting; committed scan evidence `docs/scorecard-phase0.json`. Added Section 14 and interim deviation #4 (interim email security contact pending an organizational security mailbox at GHES migration). |
| 2026-06-12 | Cedric Kiama Wachira | **Retroactive Phase 1 record (Section 15)** — Phase 1 completed 2026-06-07 (PRs #8–#20: govulncheck/gosec/staticcheck required gates via Go `tool` directive, Actions SHA-pinning, pinned-toolchain CI guard, governed Dependabot, lightweight threat model; Scorecard 5.6→7.1) but this record was not updated at the time. Omission detected during Phase 2 closeout; corrected with stale-section fixes (Sections 2, 5, 6, 8, 11, 12, 13 updated to reflect the five required checks, four CI jobs, SHA-pinned actions, and completed backlog items). |
| 2026-06-12 | Cedric Kiama Wachira | OpenSSF Hardening Phase 2 (AI guardrails, LFEL1012) complete via PRs #21–#23: `AGENTS.md` + CODEOWNERS guard; PR-template AI Assistance checklist with reviewer attestation (enforcement-visibility proven on PR #23 first render); `docs/AGENT_RUNTIME_DESIGN.md` lethal-trifecta design (design-only by ratified decision, THREAT_MODEL.md §2 back-link). Interim Scorecard 7.1→7.2 (CI-Tests 9→10); evidence `docs/scorecard-phase2-interim.json`. Added Section 16. `chore/*` branch-prefix deviation recorded on PR #22; CONTRIBUTING amendment follows in closeout PR 2. |
| 2026-06-12 | Cedric Kiama Wachira | Toolchain patch bump go 1.26.1 → 1.26.4: 18 Go stdlib advisories (all `stdlib@go1.26.1`, 0 reachable per symbol-level govulncheck) surfaced incidentally during Phase 3 SBOM-tool verification (PR #26). Bumped `go.mod` directive, CI guard assertion, and `setup-go` pins in lockstep; local toolchain upgraded first (tarball SHA-256 verified against go.dev release metadata). Doc sweep: README, AGENTS.md, Sections 2/6/8/13, THREAT_MODEL.md. Section 15.3 left as contemporaneous history. Known gap recorded in Section 12: no scheduled toolchain-freshness watcher. |
| 2026-06-12 | Cedric Kiama Wachira | OpenSSF Hardening Phase 3 (artifact integrity, scoped) complete: option 3 ratified (SBOM now, signing deferred — keyless/SLSA infeasible in air-gapped GHES); cyclonedx-gomod v1.10.0 via tool directive (PR #26); committed module SBOM `sbom/bom.json` with deterministic flags (PR #31); `supply-chain / sbom-drift` freshness gate, sixth required check on both rulesets, proven by throwaway PR #33 (RED + Required). Detours absorbed through the governed gate: toolchain bump 1.26.1->1.26.4 (PR #30, 18 stdlib advisories -> 0) and an 11-alert Dependabot security wave (PR #29; #27/#28 auto-superseded). Deferrals with consequences in 17.5; attestation deviation + restoration in 17.6. Added Section 17. |
| 2026-06-14 | Cedric Kiama Wachira | OpenSSF Hardening Phase 4 (Security Insights & Self-Assessment) STARTED; recorded in-phase (Section 18). Deliverable (c) toolchain-freshness runbook `docs/TOOLCHAIN_FRESHNESS.md` (PR #36, merge `32ebd12`) — human-executed compensating control closing the staleness gap (Section 17.2 / Section 12). Deliverable (a) OpenSSF Security Insights `.github/security-insights.yml`, schema v2.2.0 (PR #37, merge `fe269a2`) — single-repo posture, every `tools[]` assertion read from a repo artifact, validated offline with pinned `cue` v0.16.1 (clean + deliberate-break teeth-test). Filename corrected to lowercase vs plan §7. Local cue-vet recipe recorded (18.3); enforced `si-validate` CI gate deferred with a firing trigger (18.4). Deliverables (b) self-assessment and (d) Scorecard re-measure PENDING — Phase 4 remains in progress; Scorecard not yet re-measured. |
