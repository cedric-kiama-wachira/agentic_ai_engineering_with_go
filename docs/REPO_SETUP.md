# Repository Setup & Governance Record

**Repository:** `cedric-kiama-wachira/agentic_ai_engineering_with_go`
**Visibility:** Public · **License:** Apache-2.0
**Branching model:** Git Flow
**Record prepared:** 2026-06-05
**Prepared by:** Cedric Kiama Wachira (repository architect)
**Status:** Bootstrap complete — pending migration to air-gapped GitHub Enterprise Server (GHES)

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
| Go version | 1.26.1 (pinned in CI; stated in README prerequisites) |
| Source layout | `cmd/agent/main.go`, `cmd/agent/main_test.go` |
| Governance files | `README.md`, `CONTRIBUTING.md`, `LICENSE`, `SECURITY.md`, `.github/CODEOWNERS`, `.github/pull_request_template.md`, `docs/REPO_SETUP.md` |
| CI | `.github/workflows/ci.yml` |
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
Commits are documented in `CONTRIBUTING.md`.

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
| Require status checks to pass | Enabled — `build / vet / test`, `GitGuardian Security Checks`; require branches up to date before merging |
| Bypass list | **Empty** |

### 5.2 `protect-develop` (targets `develop`)

| Rule | Setting |
|------|---------|
| Restrict deletions | Enabled |
| Block force pushes | Enabled |
| Require linear history | **Disabled** (merge commits permitted on the integration branch) |
| Require pull request before merging | Enabled — **1** approval, dismiss stale approvals on new commits |
| Require signed commits | Enabled |
| Require status checks to pass | Enabled — `build / vet / test`, `GitGuardian Security Checks`; require branches up to date before merging |
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

**File:** `.github/workflows/ci.yml` — workflow `CI`, job `build / vet / test`.

| Property | Value |
|----------|-------|
| Triggers | `pull_request` and `push` to `develop`, `main` |
| Token permissions | `contents: read` (least privilege) |
| Runner | `ubuntu-latest` (GitHub-hosted) |
| Go version | `1.26.1` (pinned to match local toolchain) |
| Steps | checkout@v4 → setup-go@v5 (cached) → verify modules tidy → `go build ./...` → `go vet ./...` → `go test -race -coverprofile=coverage.out ./...` |

The module-tidiness step is written to tolerate the absence of `go.sum` (no
external dependencies yet) and will automatically begin guarding `go.sum` once
the first dependency is added (see Section 10, incident #2). The race detector is
enabled because the agentic runtime is concurrency-heavy. This job is a
**required** status check on both protected branches (Section 5).

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
| `.github/CODEOWNERS` | Auto-requests owner review on matching paths; `.github/`, `go.mod`, `go.sum` guarded explicitly. Validated by GitHub ("CODEOWNERS file is valid"). |
| `CONTRIBUTING.md` | Git Flow model, branch-naming convention, Conventional Commits, PR flow, required approvals. |
| `.github/pull_request_template.md` | Auto-populated checklist on every PR (branch origin, signing, tests, no-secrets). Verified auto-filling on PRs #2 and #4. |
| `README.md` | Project intro, prerequisites (Go 1.26.1), getting-started (canonical clone URL, not the local Host alias). |
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
| 1 | **Second account (`digital-factory-dm`) used as approving reviewer** on PRs #1, #2, #4 | Exercised an authentic two-person-review event during bootstrap | **Not genuine separation of duties** — both accounts controlled by one person. Superseded by independent human reviewers via enterprise IdP and teams. (Also confirm work-identity usage complies with applicable identity policy.) |
| 2 | **Shared deploy key used for human push access** | Single operator during bootstrap | Deploy keys are repo-scoped and not person-attributable. Retire for human use; each developer authenticates with their own account auth + signing keys. Reserve deploy keys for CI/CD or server deploys. |
| 3 | **Repository on a personal account and kept public** for ruleset enforcement | Only configuration under which Free-plan rulesets enforce | Migrate to **air-gapped GHES**: private + enforced is the default; no public exposure. |
| 4 | **Interim email security contact (`theemail@mail.com`) in `SECURITY.md`** | Single operator during bootstrap; PVR is the preferred channel regardless | Replace with a monitored organizational security mailbox at the GHES migration (see Section 14.2). |

### GHES (air-gapped) migration tasks — items that do not port unchanged

1. **CI runners:** replace `runs-on: ubuntu-latest` with **self-hosted runners**
   registered inside the perimeter. `actions/checkout@v4` and
   `actions/setup-go@v5` are pulled from the public marketplace — resolve via
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

---

## 12. Recommended Next-Phase Enhancements

Not blockers; logical follow-ons to the bootstrap foundation.

- GHES migration (resolves all Section 11 deviations and the Section 5 enforcement caveat).
- CodeQL code scanning (SAST) as an additional required check.
- Dependabot for dependency and action updates.
- Pin GitHub Action versions to commit SHAs (stricter supply-chain control).
- Strict Code-Owner-specific approval enforced on `main`.
- `release/*` branch flow with semantic version tagging.
- Branch auto-deletion on merge.
- `CODEOWNERS` entry for `/docs/` so future edits to this record require owner review.

---

## 13. Verification Checklist (for re-audit)

- [ ] Both rulesets `protect-main` and `protect-develop` show **Active** (green) — not disabled.
- [ ] Repository visibility supports enforcement (public on Free plan, or hosted on GHES/Team).
- [ ] Bypass lists are empty on both rulesets.
- [ ] Required status checks on both: `build / vet / test`, `GitGuardian Security Checks`.
- [ ] No stray free-text required checks (e.g. "GitGuardian / Any source").
- [ ] `protect-main` requires 2 approvals; `protect-develop` requires 1 (confirm live).
- [ ] Signing key present on account as type **Signing**; deploy key present on repo.
- [ ] CI workflow token permission is `contents: read`.
- [ ] GitGuardian installation scope confirmed and recorded (Section 7).
- [ ] A test PR confirms checks gate merges with nothing stuck "pending".
- [ ] `SECURITY.md` present at repo root; Private Vulnerability Reporting enabled (Section 14).

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

## Change Log

| Date | Author | Change |
|------|--------|--------|
| 2026-06-05 | Cedric Kiama Wachira | Initial bootstrap record. |
| 2026-06-05 | Cedric Kiama Wachira | Ruleset enforcement found disabled after repo set to private on personal Free plan; restored by reverting to public. Permanent remediation: migration to air-gapped GitHub Enterprise (scheduled). |
| 2026-06-05 | Cedric Kiama Wachira | Record completed post-bootstrap: added PR #4 evidence, Section 10 incidents/detections (key-role collision, CI go.sum guard, stray binary, stray GitGuardian check, enforcement-disabled incident), and Section 11 air-gapped GHES migration tasks. |
| 2026-06-06 | Cedric Kiama Wachira | OpenSSF Hardening Phase 0: established Scorecard v5.4.0 baseline (5.0 to 5.6); added coordinated-disclosure `SECURITY.md` (Security-Policy 0 to 10, PR #6) and enabled GitHub Private Vulnerability Reporting; committed scan evidence `docs/scorecard-phase0.json`. Added Section 14 and interim deviation #4 (interim email security contact pending an organizational security mailbox at GHES migration). |
