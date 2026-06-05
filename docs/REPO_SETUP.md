# Repository Setup & Governance Record

**Repository:** `cedric-kiama-wachira/agentic_ai_engineering_with_go`
**Visibility:** Public · **License:** Apache-2.0
**Branching model:** Git Flow
**Record prepared:** 2026-06-05
**Prepared by:** Cedric Kiama Wachira (repository architect)
**Status:** Bootstrap complete — pending organisation migration at team onboarding

> **Auditor note.** This document records the controls configured during the
> bootstrap phase and the evidence that each was verified on live artifacts.
> The rule values in §5 reflect the intended configuration; before formal
> submission, confirm them against the live rulesets at
> `Settings → Rules → Rulesets`, as they may be tuned over time.

---

## 1. Purpose

This repository is the foundation for an Agentic AI engineering initiative
(implemented in Go). It is designed to onboard ~30 developers behind controls
that enforce themselves, so that secure, reviewed, spec-driven development can
proceed without relying on individual discipline. This record documents the
governance posture established before the team joins.

---

## 2. Repository Overview

| Item | Value |
|------|-------|
| Default branch | `develop` |
| Protected branches | `main`, `develop` |
| Source layout | `cmd/agent/` (entrypoint), Go module `github.com/cedric-kiama-wachira/agentic_ai_engineering_with_go` |
| Governance files | `README.md`, `CONTRIBUTING.md`, `.github/CODEOWNERS`, `.github/pull_request_template.md` |
| CI | `.github/workflows/ci.yml` |

---

## 3. Cryptographic Key Architecture

Three distinct key-roles are used. A deliberate design decision was made to use
**separate keys for transport and signing**, because GitHub enforces that any
single public key is globally unique to one role across the platform — a repo
deploy key cannot simultaneously be registered as an account key. Separating
the keys also follows least-privilege: transport (push) and identity (signing)
are independent concerns.

| Role | Key | Algorithm | Where registered | Purpose |
|------|-----|-----------|------------------|---------|
| **Transport / deploy** | `id_dev_agentic_ai_lab_ed25519` | Ed25519 | Repo **Deploy key** (write access); referenced in local `~/.ssh/config` under Host alias `dev_agentic_ai` | Authenticates `git push`/`pull` to the repository |
| **Commit signing (local)** | `id_signing_agentic_ai_lab_ed25519` | Ed25519 | `git config user.signingkey` + `~/.ssh/allowed_signers` | Signs commits; enables local signature verification |
| **Commit signing (GitHub)** | same public key as above | Ed25519 | Account **SSH and GPG keys → Signing key** (`SHA256:QlTAWcCOBrHWsxpOeqJQhQS4gR5z8QLac1uZObv4m/w`) | Enables GitHub-side "Verified" badge on signed commits |

**Notes for audit:**
- Signing keys are **not** placed in `~/.ssh/config` — signing performs no network
  connection, so it has no SSH Host entry. This is correct and intentional.
- Merge commits created through the GitHub web UI are signed by GitHub's own
  web-flow key and display as "Verified"; this satisfies the signed-commit rule.
  Local `git log --show-signature` on such commits reports "cannot check
  signature — no public key" because GitHub's key is not in the local keyring;
  this is cosmetic, not a verification failure.

---

## 4. Branching Model (Git Flow)

Direct commits to `main` and `develop` are prohibited by ruleset (§5). All
changes enter through reviewed pull requests.

| Branch | Branches from | Merges into | Purpose |
|--------|---------------|-------------|---------|
| `main` | — | — | Production-ready, tagged releases |
| `develop` | `main` | `main` (via PR) | Integration branch (default) |
| `feature/*` | `develop` | `develop` (via PR) | New work |
| `bugfix/*` | `develop` | `develop` (via PR) | Non-urgent fixes |
| `hotfix/*` | `main` | `main` + `develop` | Urgent production fixes |

Branch naming and commit conventions (Conventional Commits) are documented in
`CONTRIBUTING.md`.

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
| Bypass list | **Empty** (no actor, including admin, may bypass silently) |

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
list means any override is performed deliberately and is logged as a recorded
rule bypass — an auditable event rather than a silent one.

---

## 6. Continuous Integration

**File:** `.github/workflows/ci.yml` — workflow `CI`, job `build / vet / test`.

| Property | Value |
|----------|-------|
| Triggers | `pull_request` and `push` to `develop`, `main` |
| Token permissions | `contents: read` (least privilege) |
| Runner | `ubuntu-latest` |
| Go version | `1.26.1` (pinned to match local toolchain) |
| Steps | checkout → setup-go (cached) → verify modules tidy → `go build ./...` → `go vet ./...` → `go test -race -coverprofile=coverage.out ./...` |

The module-tidiness step is written to tolerate the absence of `go.sum` (no
external dependencies yet) and will automatically begin guarding `go.sum` once
the first dependency is added. The race detector is enabled because the
agentic runtime is concurrency-heavy.

This CI job is a **required** status check on both protected branches (§5).

---

## 7. Secret Scanning

| Property | Value |
|----------|-------|
| Provider | GitGuardian (GitHub App) |
| Required check name | `GitGuardian Security Checks` |
| Status | Required on `main` and `develop` |

**Open item for audit:** confirm the installation scope of the GitGuardian app
(repository-level vs organisation-level) and record it here. A free-text
"GitGuardian / Any source" entry was briefly added in error during setup and
removed; only the app-reported `GitGuardian Security Checks` is required.

---

## 8. Code Ownership & Contribution Governance

| Artifact | Function |
|----------|----------|
| `.github/CODEOWNERS` | Auto-requests owner review on matching paths; `.github/`, `go.mod`, `go.sum` guarded explicitly. Validated by GitHub ("CODEOWNERS file is valid"). |
| `CONTRIBUTING.md` | Git Flow model, branch-naming convention, Conventional Commits, PR flow. |
| `.github/pull_request_template.md` | Auto-populated checklist on every PR (branch origin, signing, tests, no-secrets). |

---

## 9. Bootstrap Validation Evidence

Each control was proven on a live pull request, not merely configured.

| PR | Change | Outcome | Evidence |
|----|--------|---------|----------|
| #1 | Governance docs (CODEOWNERS, CONTRIBUTING, PR template) | Merged to `develop` via reviewed PR | Merge commit `cb8f769`; feature commit `e1ad152` Verified |
| #2 | Go module, agent entrypoint, CI workflow | Merged to `develop` via reviewed PR | Merge commit `1b659e7` Verified; 2 checks passed |
| #3 | Trivial change to confirm required checks gate merges | **Closed without merging** — verification only | `build / vet / test` and `GitGuardian Security Checks` both reported **Required** and green; review still correctly blocked merge |

Additional verified behaviours during bootstrap:
- A direct push to `develop` was **rejected** (`GH013` — "Changes must be made
  through a pull request" and "Commits must have verified signatures"),
  confirming the PR-only and signed-commit rules.
- Commit signatures verify end-to-end (local "Good signature"; GitHub
  "Verified" badge).

---

## 10. Known Interim Deviations & Remediation Plan

These are bootstrap-phase compromises, recorded transparently. None are intended
to persist into team operation.

| # | Deviation | Why acceptable now | Remediation at onboarding |
|---|-----------|--------------------|---------------------------|
| 1 | **Second account (`digital-factory-dm`, work email `cwachira.v@dm.gov.ae`) used as the approving reviewer** on PRs #1–#2. | Allowed an authentic two-person-review event to be exercised during bootstrap. | **Not genuine separation of duties** — two accounts controlled by one person. Real four-eyes review requires a *different human*. To be superseded by independent reviewers when the team joins. Confirm `dm.gov.ae` work-identity usage complies with the customer's own identity policy. |
| 2 | **Shared deploy key used for human push access.** | Single operator during bootstrap; acceptable for one person. | Deploy keys are repo-scoped and not tied to a person, eroding attribution. Retire for human use; each developer to authenticate with their own account auth + signing keys. Reserve deploy keys for CI/CD or server deploys only. |
| 3 | **Repository owned by a personal account.** | Fastest path to a working, protected repo. | Migrate to an **organisation-owned** repository before the 30 developers join; manage access via org **teams** (ideally SSO/SCIM) rather than individual collaborators; promote these rulesets to **organisation-level** so future repos inherit them. |

---

## 11. Recommended Next-Phase Enhancements

Not blockers; logical follow-ons to the bootstrap foundation.

- Organisation migration (resolves §10 #1–#3 together).
- CodeQL code scanning (SAST) as an additional required check.
- Dependabot for dependency and action updates.
- Pin GitHub Action versions to commit SHAs (stricter supply-chain control).
- Strict Code-Owner-specific approval enforced on `main`.
- `release/*` branch flow with semantic version tagging.
- Branch auto-deletion on merge.

---

## 12. Verification Checklist (for re-audit)

- [ ] Both rulesets `protect-main` and `protect-develop` show **Active**.
- [ ] Bypass lists are empty on both.
- [ ] Required status checks on both: `build / vet / test`, `GitGuardian Security Checks`.
- [ ] No stray free-text required checks (e.g. "GitGuardian / Any source").
- [ ] Signing key present on account as type **Signing**; deploy key present on repo.
- [ ] CI workflow token permission is `contents: read`.
- [ ] GitGuardian installation scope confirmed and recorded (§7).
- [ ] A test PR confirms checks gate merges with nothing stuck "pending".

---

## Change Log

| Date | Author | Change |
|------|--------|--------|
| 2026-06-05 | Cedric Kiama Wachira | Initial bootstrap record. |
| 2026-06-05 | Cedric Kiama Wachira | Ruleset enforcement found disabled after repo set to private on personal Free plan; restored by reverting to public. Permanent remediation: migration to air-gapped GitHub Enterprise (scheduled). |
