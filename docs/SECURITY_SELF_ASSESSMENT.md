# Security Self-Assessment — Agentic AI Lab (Go)

**Repository:** `cedric-kiama-wachira/agentic_ai_engineering_with_go` (public, Apache-2.0)
**Status:** Pre-runtime — bootstrap + OpenSSF Foundation Hardening Phases 0–3 complete, Phase 4 in progress
**Version:** 1.0 · **Last updated:** 2026-06-14 · **Maintainer:** Cedric Kiama Wachira

> **Purpose.** This is the structured security posture of the Lab (OpenSSF
> LFEL1005). It states what the project is, the trust boundaries that exist
> today, the controls already enforced, and — with equal weight — the residual
> risks and deliberate deferrals. It is the artifact an auditor should read
> first. It is a **living document**, re-reviewed at each phase and each release
> alongside the OpenSSF Scorecard trend. It does not duplicate the governance
> evidence in `docs/REPO_SETUP.md` or the threat analysis in
> `docs/THREAT_MODEL.md`; it cross-references them so governance and security
> read as one package.

---

## 1. Overview — what the Lab is, and is not, today

The Agentic AI Lab is intended to become a secure, auditable foundation for
agentic AI engineering in Go. The defining decision is that **security and
auditability are requirements from before the first line of application code**,
not a later phase.

**Current state (important for scoping every claim below):** the repository
contains only trivial scaffolding (`cmd/agent`). **No agent runtime, no
application logic, and no AI features exist yet.** The security work completed to
date hardens the *software-production system* — the governed repository and its
CI/CD supply chain — not a running application. Where this document describes
agent-runtime concerns (tools, untrusted content, private data, network egress),
those are **design-stage and deferred**, not implemented controls.

**Governance thesis ("no vibe coding").** Every change — human- or AI-authored —
passes the identical gate: a signed commit on a branch → pull request →
independent review → required CI checks → secret scan → merge commit. Nothing
can reach `develop` or `main` unreviewed, which is the structural answer to "no
vibe coding": AI may *author*, but AI output is an untrusted contribution that
must survive the same review and the same gates as any other change.

---

## 2. Scope & trust boundaries

### 2.1 Boundaries that exist today

The only trust boundaries currently in force are those of the
production/governance system (modeled in `THREAT_MODEL.md` §2):

- **Author → repository.** Every change is an Ed25519-signed, GitHub-verified
  commit on a feature branch, entering `develop` only through a reviewed PR.
  Direct pushes to `develop`/`main` are rejected by ruleset. The "author" is
  either a human developer or **Dependabot**; both cross the *identical*
  boundary (signed commit, full check gate, independent review). Dependabot has
  no privileged path around the gate (`THREAT_MODEL.md` §4).
- **CI runners → code.** GitHub-hosted runners execute the pipeline under a
  read-only token (`contents: read`); Actions are SHA-pinned.
- **GitHub platform → trust root.** Signature verification and ruleset
  enforcement depend on github.com. This is a deliberate, recorded dependency
  with a migration delta to air-gapped GHES (§6; `THREAT_MODEL.md` §7).

### 2.2 Boundaries that are deferred (no runtime yet)

The agent-runtime boundaries — agent ↔ untrusted content ↔ tools ↔ private data
↔ network egress, i.e. the "lethal trifecta" — **do not exist yet** because no
runtime code exists. Their *isolation design* is captured in
`docs/AGENT_RUNTIME_DESIGN.md` (DESIGN-ONLY, with per-principle implementation
gates); a design document is explicitly **not** a control. A full STRIDE /
data-flow threat model of the runtime is **out of scope for this assessment**
because there is no running system to analyze; modeling an imagined system would
manufacture false confidence. When runtime code lands, that analysis lands in
`THREAT_MODEL.md` §2, and the `AGENT_RUNTIME_DESIGN.md` implementation gates
convert from design to verified controls one at a time.

---

## 3. Security functions & control inventory

All controls below are enforced and were verified on live artifacts, not merely
configured (the project's standing rule: a control configured but not proven on
a live artifact is not yet a control). Each cites its authoritative record in
`REPO_SETUP.md`.

### 3.1 Repository governance

- **Branch-protection rulesets**, Active, bypass lists empty (`REPO_SETUP.md`
  §5): `protect-main` (2 approvals, linear history) and `protect-develop`
  (1 approval, merge commits permitted). Both require signed commits, passing
  required status checks, restrict deletions, block force-pushes, and require
  branches up to date before merging.
- **Commit signing** with a dedicated Ed25519 SSH signing key, separate from the
  transport/deploy key (least privilege; `REPO_SETUP.md` §3).
- **Code ownership & contribution governance** — `CODEOWNERS` (guarding
  `.github/`, `go.mod`, `go.sum`, `AGENTS.md`), `CONTRIBUTING.md`, and the PR
  template (`REPO_SETUP.md` §8).

### 3.2 Required CI gates

The CI workflow (`.github/workflows/ci.yml`) runs under `contents: read` with
SHA-pinned Actions and `GOTOOLCHAIN=local`. There are **6 required status checks
on both rulesets — 5 CI jobs plus 1 GitHub App** (verified live):

| Required check | Kind | What it gates |
|----------------|------|---------------|
| `build / vet / test` | CI job | modules-tidy → pinned-toolchain guard → `go build`/`go vet`/`go test -race` |
| `security / govulncheck` | CI job | reachability-scoped known-vulnerability scanning (SCA) |
| `security / gosec` | CI job | security static analysis (SAST); strict `#nosec` policy |
| `quality / staticcheck` | CI job | correctness/quality linter (not a security scanner) |
| `supply-chain / sbom-drift` | CI job | regenerate-and-diff the committed SBOM; fail on drift |
| `GitGuardian Security Checks` | GitHub App | secret scanning |

(`govulncheck` reports vulnerabilities on **reachable** code paths by design;
unreachable-but-present advisories are out of scope as a deliberate tool
property, recorded in `THREAT_MODEL.md` §3.)

### 3.3 Supply-chain integrity

- **Committed CycloneDX SBOM** (`sbom/bom.json`, CycloneDX 1.6) generated by
  `cyclonedx-gomod` via the Go `tool` directive, kept fresh by the
  `supply-chain / sbom-drift` required gate (`REPO_SETUP.md` §17.4). **Integrity
  framing, stated precisely:** the signed-commit chain provides *file
  provenance*; the drift gate provides *content freshness*; **authenticity
  attestation does not exist** — SBOM signing is deferred (§5). Until then the
  SBOM is an **informational artifact, not a trusted attestation**.
- **Pinned-Go toolchain guard** — CI asserts no `toolchain` directive and a
  `go` directive of exactly `1.26.4` (`REPO_SETUP.md` §15.3); prevents silent
  toolchain drift from a dependency bump.
- **GitHub Actions SHA-pinned** with `# vX.Y.Z` annotations (`REPO_SETUP.md`
  §15.2); closes the repointable-moving-tag risk.
- **Governed Dependabot** — `gomod` + `github-actions`, weekly, grouped,
  `open-pull-requests-limit: 5`, **no auto-merge, no ruleset bypass**; bot PRs
  face the identical 6-check + review gate (`REPO_SETUP.md` §15.4;
  `THREAT_MODEL.md` §4).
- **Inert AI-provider SDKs** — transitive Go SDKs (`anthropic-sdk-go` via
  `gosec` autofix; `openai-go`, `generative-ai-go` unneeded by the main module)
  are accepted as inert: no AI features are wired in, the autofix path is not
  invoked, the unneeded modules are not imported (`THREAT_MODEL.md` §3). Re-evaluate
  when/if the runtime calls an AI provider.

### 3.4 Toolchain-freshness procedure (compensating, human-executed)

`docs/TOOLCHAIN_FRESHNESS.md` is a weekly operator procedure that closes the
**staleness** gap the pinned-toolchain guard does not cover (the guard prevents
drift, not aging behind published patches). It is explicitly a **COMPENSATING
CONTROL, HUMAN-EXECUTED — not a CI-enforced gate** (`REPO_SETUP.md` §18.1). Its
30-day DB-age bound is **provisional and not yet calibrated** (the observation
run-log does not yet exist). It converts to a scheduled CI gate at the GHES
migration, where DB-age becomes a primary signal against a local mirror.

### 3.5 Machine-readable posture & AI guardrails

- **OpenSSF Security Insights** — `.github/security-insights.yml` (schema
  v2.2.0) publishes this posture in machine-readable form for consumers such as
  Scorecard / CLOMonitor / LFX Insights (`REPO_SETUP.md` §18.2).
- **AI-assistant guardrails** — `AGENTS.md` (OpenSSF-derived contribution rules,
  CODEOWNERS-guarded) plus the PR-template **AI Assistance** section, where the
  author declares AI use and the *reviewer* attests (logic verified, no
  hallucinated APIs, no unvetted dependencies, no leaked context) — never
  pre-ticked by the author (`REPO_SETUP.md` §16).

### 3.6 Coordinated disclosure

`SECURITY.md` advertises GitHub Private Vulnerability Reporting (enabled) plus an
interim email contact (`REPO_SETUP.md` §14.2). The interim contact is a
bootstrap deviation resolved at GHES (§6).

---

## 4. Verification posture & coverage gaps

**What is verified, and how:** static and composition analysis in CI —
`gosec`/`staticcheck`/`go vet` (SAST/correctness), `govulncheck` (SCA, reachable
paths), GitGuardian (secrets) — plus SBOM-drift detection. Enforcement is proven
by deliberately injecting failures and confirming the gate fires, not trusted
from settings screens.

**Honest coverage gaps:**
- **No dynamic analysis (DAST).** All gates are static/composition; there is no
  running service to exercise. Flagged against LFD121's full static+dynamic
  model and outside Scorecard's scope (`THREAT_MODEL.md` §5; `REPO_SETUP.md`
  §14.3).
- **No fuzzing.** A candidate once the runtime has parseable inputs / tool-call
  boundaries.
- **No runtime/behavioral testing** — there is no runtime.
- **SAST metric vs reality.** Scorecard reports SAST 0/10 because it
  pattern-matches marketplace SAST actions and cannot detect `go tool gosec` /
  `go tool staticcheck` invocations. This is a **detection blind spot, not a
  control gap** (`REPO_SETUP.md` §15.6); the metric is not chased at the cost of
  air-gap portability.

---

## 5. Residual risks, known limitations & deliberate deferrals

Each item below is a conscious scoping decision with a recorded remediation
path; none is an unmanaged risk.

- **No runtime → runtime threat model deferred.** STRIDE/DFD analysis and the
  lethal-trifecta isolation controls are design-only (`AGENT_RUNTIME_DESIGN.md`)
  until application code exists.
- **Artifact authenticity not yet established.** The SBOM is informational;
  **signing is deferred** to the GHES signing-architecture decision (key-based
  cosign ≥ 2.6.2 / CVE-2026-22703 vs private Sigstore). No trusted attestation
  exists until then (`REPO_SETUP.md` §17.5; `THREAT_MODEL.md` §3).
- **Tag/release governance deferred.** No tags/releases/`release/*` flow yet
  (`cmd/agent` is scaffolding). A release workstream additionally needs a
  build-resolved `app`/`bin` SBOM — the committed **module** SBOM does not
  pre-satisfy release provenance (`REPO_SETUP.md` §17.5).
- **SI-file validation is local, not enforced.** `.github/security-insights.yml`
  is validated by a documented local `cue vet` recipe; an enforced `si-validate`
  CI gate is deferred to land with the SBOM-style enforcement plumbing or at
  GHES, whichever is first (`REPO_SETUP.md` §18.3–18.4).
- **Toolchain-freshness is human-executed.** A compensating control, not an
  enforced gate; its age bound is uncalibrated (§3.4). The human-execution
  residual is the gap until the GHES scheduled-gate conversion.
- **Separation of duties is interim.** A single operator controls both the
  author account and the `digital-factory-dm` reviewer account, and a shared
  deploy key is used for human push — **not genuine separation of duties**.
  Resolved by independent reviewers and per-developer keys via the enterprise
  IdP at GHES (`REPO_SETUP.md` §11).
- **Enforcement depends on the repo staying public.** On a personal Free plan,
  rulesets enforce on public repositories only; setting the repo private
  silently disables enforcement. The repo is kept public until GHES, where
  private + enforced is the default (`REPO_SETUP.md` §5, §11; incident #5).
- **github.com platform dependencies** (PVR, GitGuardian cloud, hosted runners,
  Dependabot, the `govulncheck` vuln DB) do not port unchanged to air-gapped
  GHES; each has a recorded migration delta, and a **stale vuln-DB mirror is a
  false-green failure mode** treated as a monitored control
  (`THREAT_MODEL.md` §7; `REPO_SETUP.md` §11).

The standing GHES migration backlog (`REPO_SETUP.md` §11) is the single place
where these deferrals' remediations are tracked.

---

## 6. Bootstrap deviations (interim)

Recorded in full in `REPO_SETUP.md` §11 and not duplicated here: the
second-account reviewer, the shared deploy key for human push, the interim
`SECURITY.md` email contact, and keeping the repo public for enforcement. All
are interim and resolve at the GHES migration.

---

## 7. Regulatory alignment — EU CRA (informational)

The Lab is operated by a UAE government agency; the EU Cyber Resilience Act is
**not** a current obligation. It is noted only for forward awareness: the
controls already in place — a machine-readable SBOM, coordinated vulnerability
disclosure (`SECURITY.md` + PVR), and secure-by-default governance — provide
substantial CRA-style alignment **should EU exposure ever arise** (e.g. shipping
a component or SDK reaching EU users). No dedicated CRA work is undertaken now;
if that day comes, the additions would be a defined support/patch window and a
formalized coordinated-disclosure timeline.

---

## 8. Cross-references & maintenance

- `docs/REPO_SETUP.md` — governance/audit record (controls, evidence,
  incidents, deviations, GHES backlog, per-phase history).
- `docs/THREAT_MODEL.md` — current-state threat model (trust boundaries,
  supply-chain risks, Dependabot posture, verification gaps, GHES deltas).
- `docs/AGENT_RUNTIME_DESIGN.md` — lethal-trifecta isolation design (DESIGN-ONLY).
- `docs/TOOLCHAIN_FRESHNESS.md` — the weekly toolchain-freshness procedure (c).
- `.github/security-insights.yml` — machine-readable Security Insights (a).
- `SECURITY.md` — coordinated disclosure policy and intake channels.

**Maintenance.** This self-assessment is reviewed at each OpenSSF hardening phase
and at each release, paired with the Scorecard trend as the two top-level posture
indicators. The Scorecard trend to date is **5.0 → 5.6 → 7.1 → 7.2 → 7.2**
(pinned binary v5.4.0); the formal Phase 4 re-measure is pending (deliverable
(d)) and is **not** pre-stated here.
