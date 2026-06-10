# Threat Model - Agentic AI Lab (Go)

**Status:** Phase 1 (lightweight) · **Version:** 1.0 · **Last updated:** 2026-06-07
**Maintainer:** Cedric Kiama Wachira

> **Scope note.** This document models the **current** state of the system: a
> governed source repository and its CI/CD supply chain. The Lab's **agent
> runtime does not yet exist** (no application code has landed). The
> agent-runtime threat model - trust boundaries between the agent, untrusted
> content, tools, private data, and network egress (the "lethal trifecta") -
> will be added here as application code is written (Phase 2 onward). This
> lightweight artifact is intentionally scoped to what exists; the full
> STRIDE/data-flow-level analysis lands in the **Phase 4 security
> self-assessment** (LFEL1005), once there is a running system to analyze.

---

## 1. Scope & method

**In scope (now):** repository governance (branch protection, signed commits,
review), the CI/CD pipeline, and the dependency/supply-chain tooling that gates
every change.

**Deferred:** the agent runtime and its data flows - modeled when the code
exists, to avoid analyzing an imagined system (which manufactures false
confidence). Forward-linked to Phase 2 (AI guardrails) and Phase 4
(self-assessment).

**Method:** lightweight, control-oriented. For each risk we state the control
in place and any residual gap, cross-referencing `REPO_SETUP.md` (governance
evidence) rather than duplicating it.

---

## 2. Trust boundaries (current state)

The surfaces that exist today:

```
author ── signed commit ──> feature branch ── PR ──> CI runners ──> GitHub (develop/main)
                                               |
                                 (5 required checks + independent review)
```

- **Author to repository:** every change is a signed (Ed25519, verified)
  commit on a feature branch, entering `develop` only through a reviewed PR.
  Direct pushes to `develop`/`main` are rejected by ruleset.
- **CI runners to code:** GitHub-hosted runners execute the pipeline with a
  read-only token (`contents: read`). Actions are SHA-pinned.
- **GitHub platform:** the trust root for signature verification and ruleset
  enforcement (a github.com dependency; see section 7 for the GHES migration delta).

**Two commit authors, one boundary.** The "author" above is either a human
developer or **Dependabot** (for dependency bumps). Both are subject to the
*identical* boundary - signed/verified commit, PR, the full check gate, and
independent review. Dependabot holds no privileged path around this (see
section 4).

**Future boundaries (not yet present):** agent, untrusted content, tools,
private data, network egress. To be modeled in Phase 2 when the runtime lands.
Isolation design now captured in `docs/AGENT_RUNTIME_DESIGN.md` (Phase 2.3);
threat-level modeling still lands here once runtime code exists.

---

## 3. Supply-chain risks & controls

**Known-vulnerability exposure.** `govulncheck` runs as a required CI gate.
Note its scope: it reports vulnerabilities on **called** code paths
(reachability analysis). Vulnerabilities present in the dependency tree but not
on a reachable path are out of scope **by design** - this is a deliberate
property of the tool, not a coverage miss. Recorded so the distinction is
explicit to a reviewer.

**AI-provider SDK transitive dependencies.** The module graph contains
AI-provider Go SDKs introduced by the security tooling, not by application code:

- `github.com/anthropics/anthropic-sdk-go` is a live transitive import, pulled
  in by `gosec`'s autofix feature
  (`gosec/cmd/gosec -> gosec/autofix -> anthropic-sdk-go`, per `go mod why`).
- `github.com/openai/openai-go` and `github.com/google/generative-ai-go` appear
  in the module graph but are **not needed by the main module** (`go mod why`
  reports "main module does not need package ...").

All are accepted as **inert**: no AI features are wired into the Lab, the
autofix path is not invoked (no autofix flags are used), and the unneeded
modules are not imported at all. They present no runtime attack surface today.
Documented decision, not an unmanaged risk; re-evaluate if/when the runtime
actually calls an AI provider.

**Pinned-Go toolchain integrity.** A dependency bump (via `go get` or Dependabot)
can introduce a `toolchain` directive or raise the `go` directive in `go.mod`
as a *side-effect* of bumping another dependency. This would silently drift the
deliberately pinned Go version (1.26.1). **Control:** a CI guard ("Verify pinned
Go toolchain") asserts no `toolchain` directive exists and the `go` directive is
exactly `1.26.1`, plus workflow-level `GOTOOLCHAIN=local` so the build refuses to
silently switch toolchains. Verified on a live artifact (deliberate injection ->
CI red at the guard; reverted -> green).

**Action supply chain.** GitHub Actions are pinned to full commit SHAs (with
`# vX.Y.Z` annotations), closing the repointable-moving-tag risk. Dependabot
keeps the pins current via governed PRs (see section 4).

**Forward item (Phase 3).** When artifact signing is added, pin **cosign
>= 2.6.2** from the outset (CVE-2026-22703). Not a current risk - no cosign is
present yet - but captured now so the Phase 3 work inherits the decision.

---

## 4. Dependabot privilege & branch posture

- **No ruleset bypass.** Dependabot is **not** on the bypass list of either
  `protect-develop` or `protect-main` (both verified empty). Its PRs are gated
  **identically to human PRs** - five required checks plus independent review.
  The "confused-deputy with bypass" path (a privileged bot merging unreviewed
  code) does not exist here, by design.
- **Dependabot branch posture.** Dependabot's own branches (`dependabot/*`) are
  not directly targeted by the rulesets, but they **cannot merge** into a
  protected branch without passing the full gate. The branch-injection /
  confused-deputy class therefore cannot reach `develop` or `main`: an
  unreviewed or check-failing change is blocked at merge.
- **Attack-surface reduction (Jan 2026).** GitHub removed several Dependabot
  comment commands on 2026-01-27 (`merge`, `cancel merge`, `squash and merge`,
  `close`, `reopen`). This eliminated the `@dependabot merge`-based
  confused-deputy variant documented in public research - a net reduction in
  attack surface.
- **Commit signing.** Dependabot commits are signed and show **Verified** in the
  normal case (observed on a live actions-bump PR: Verified badge, five checks
  green, `chore:` prefix). An intermittent upstream bug can occasionally produce
  an **unsigned** Dependabot commit; this is handled by the recovery procedure
  in `CONTRIBUTING.md` and the signing rule is **never weakened** to accommodate
  the bot.
- **Recovery procedure (current GitHub behavior).** If a Dependabot commit lands
  unsigned: try `@dependabot recreate` **once on the still-open PR**; if it fails
  (it is fragile on closed PRs / deleted branches, and can reference the removed
  `reopen` command), **author the bump as a normal signed human PR** - the
  reliable path. Use GitHub's native UI/CLI for close/reopen (the Dependabot
  commands no longer exist).

---

## 5. Verification-coverage gaps (honest limitations)

- **No dynamic analysis (DAST).** The verification gates are static/composition
  analysis - `gosec`, `staticcheck`, `go vet` (SAST), `govulncheck` (SCA),
  GitGuardian (secret scanning). There is **no dynamic application security
  testing**. Flagged against LFD121's full static+dynamic verification model;
  also outside OpenSSF Scorecard's scope. To be closed when there is a running
  service to exercise.
- **No fuzzing.** Not yet present (cf. Scorecard's Fuzzing heuristic). A
  candidate once the runtime has parseable inputs / tool-call boundaries.
- **Scope of this document.** This is the **Phase-1 lightweight** threat model.
  STRIDE/data-flow-level rigor is deferred to the **Phase 4 security
  self-assessment (LFEL1005)**, when the runtime exists to be analyzed - so the
  intentional thinness here is a sequencing decision, not an analysis gap.

---

## 6. Bootstrap deviations

These are recorded in full in `REPO_SETUP.md` (Section 11) and **not duplicated
here**: the second-account independent reviewer (`digital-factory-dm`), the
shared deploy key for human push, and keeping the repository public on the
personal Free plan for ruleset enforcement. All are interim, all resolve at the
GHES migration. See `REPO_SETUP.md` for the deviation log and remediation plan.

---

## 7. GHES (air-gapped) migration deltas

Several current controls are github.com-dependent and change in the air-gapped
GitHub Enterprise Server target (full list in `REPO_SETUP.md`):

- **Dependabot** -> requires GitHub Connect or a self-hosted setup.
- **GitGuardian (cloud app)** -> GHES Advanced Security secret scanning, or
  self-hosted GitGuardian.
- **GitHub-hosted runners** -> self-hosted runners inside the perimeter; Actions
  resolved via Actions sync / vendored copies.
- **Private Vulnerability Reporting (cloud)** -> GHES advisories.
- **`govulncheck` vulnerability database** -> govulncheck pulls advisory data
  from the Go vulnerability database (vuln.go.dev) over the network. Air-gapped
  operation requires a **mirrored/local copy of the vuln DB, refreshed on a
  defined cadence**. Note the failure mode: a stale vuln DB yields a
  **false-green** gate - it passes because it does not know about newer CVEs.
  This is the same false-confidence pattern as a silently-disabled ruleset;
  treat DB freshness as a monitored control, not a set-and-forget mirror.

This threat model is revisited at migration and at each release.
