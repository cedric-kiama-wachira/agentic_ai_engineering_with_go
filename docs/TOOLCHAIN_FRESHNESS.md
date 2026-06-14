# Toolchain & Dependency Freshness — Weekly Operator Procedure

STATUS: ACTIVE OPERATOR PROCEDURE — COMPENSATING CONTROL, HUMAN-EXECUTED
(not CI-enforced; converts to a scheduled gate at GHES — see § 5 Conversion)

## 1. Purpose

The Phase-1 toolchain pin-guard (`GOTOOLCHAIN=local` + the CI assertion that
`go.mod` carries no `toolchain` line and the `go` directive is exact) prevents
toolchain *drift* — an unintended version change. It does **not** detect
*staleness*: a newly-published advisory against the *pinned* toolchain or a
pinned dependency does not announce itself, because nothing in the repo changed.

This gap is not hypothetical. During Phase 3, 18 stdlib advisories against the
then-pinned `go 1.26.1` were discovered **incidentally** while verifying an
unrelated change — not by any standing control. Absent that coincidence they
would have sat unnoticed. This procedure is the weekly check that closes the
gap until it is automated at GHES.

It is a **compensating control, human-executed**: its reliability depends on the
operator actually running it on cadence. That human dependency is itself the
residual risk the GHES conversion (§ 5) closes by moving the check into
scheduled CI.

## 2. Cadence

**Weekly.** The *run* cadence is deliberately independent of how often the Go
vulnerability database updates. That database is curated and human-reviewed by
the Go Security team with no published fixed cadence, so a multi-day quiet
stretch with no new entries is normal, not a fault — which is why DB age is an
*advisory* signal here (§ 4.4), not a gate.

## 3. The check

Run from the repo root (`~/Agentic_AI_Lab/GO_Labs/lab_1`):

    go tool govulncheck -format json ./... > /tmp/gv-fresh.json 2> /tmp/gv-fresh.err; GV_EXIT=$?
    jq -rs '.[0].config' /tmp/gv-fresh.json        # config frame object, or `null` if absent
    grep -c '"finding"' /tmp/gv-fresh.json          # reachable-vuln count — finding frames ONLY
    echo "GV_EXIT=$GV_EXIT"; cat /tmp/gv-fresh.err   # exit + diagnostic on failure

The command is **de-pipelined on purpose**: govulncheck writes to a file, and
`GV_EXIT=$?` captures govulncheck's own exit with no pipe in between (avoiding
the PIPESTATUS trap where `$?` would otherwise read the last pipe stage). Do
**not** rewrap any stage in a pipe — doing so silently reintroduces that trap.

> **`grep -c` footgun:** `grep -c` prints the count and exits **1 when the count
> is 0**. That is normal here, not a failure — the authoritative exit is
> `GV_EXIT`, captured before grep runs. Never fold grep's exit into a combined
> status, and do not wrap this block in `set -e`.

## 4. Interpretation

| `config` frame | `finding` frames | Meaning | Action |
|---|---|---|---|
| present (jq → object) | `0` | DB reached; no *reachable* vulns | PASS → run § 4.4 advisory age check |
| present (jq → object) | `≥ 1` | DB reached; reachable vulns found *(schema-grounded; not locally triggered)* | READ the `finding` frames in `/tmp/gv-fresh.json`; triage |
| **absent** (jq → `null` **or** jq error) | — | No usable config frame: DB unparseable/unreachable | **INCONCLUSIVE — gate fires.** Read `/tmp/gv-fresh.err`, restore DB access, re-run. A result is not trusted until a config frame appears. |

### 4.1  osv frames are NOT findings — do not count the wrong frame
A clean run streams `osv` frames (observed: `GO-2021-0067`, `GO-2021-0069`
against `stdlib`). These mean "this vuln exists in the DB," **not** "your code
reaches it," and appear on essentially every run. Reachable-vuln detection keys
on `finding` frames only. Counting `osv` frames would false-positive on every
execution. *(Observed 2026-06-14: a clean run emitted osv frames and
`finding`-count `0` simultaneously.)*

### 4.2  The "absent" cell — observed vs. defensively handled
*Observed path (2026-06-14):* a forced DB failure (dead localhost port,
`-db http://127.0.0.1:1`) produced a **0-byte** stdout file; `jq -rs` on empty
input returns `null` (exit 0), `GV_EXIT` was `1`, and stderr carried
`creating client: unrecognized vulndb format`.
*Defensive handling:* a different failure (e.g. an endpoint returning a
truncated-but-non-empty body) could instead yield malformed JSON on which `jq`
*errors*. That path was **not** tested — but it needs no separate handling:
**any jq outcome other than a populated object — `null` or a jq error — means
"no usable config frame → gate fires."** The cell is robust to failure modes
not enumerated.

### 4.3  Toolchain assertion
The config frame's `go_version` must equal the `go` directive in `go.mod`
(currently `go1.26.4`). A mismatch means the scan did not run on the pinned
toolchain — investigate before trusting the result. *(Observed 2026-06-14:
`go_version: go1.26.4`, matching the pin.)*

### 4.4  Advisory age check — PROVISIONAL bound
Read `db_last_modified` from the config frame. If older than **30 days**,
*investigate* whether the endpoint/network is serving fresh data — do **not**
auto-fail. On the canonical online source (`https://vuln.go.dev`) staleness is
benign (quiet curated feed), so age is advisory here, never a gate.
**30 days is a provisional "raise-an-eyebrow" bound, NOT measured against
observed entry spacing.**

Each run, record `db_last_modified` and the run date in the operator freshness
run-log (`docs/freshness-runlog.md` — see § 4.5). That accumulating series will
be the observed cadence the 30-day bound is calibrated against, once it exists;
without it the calibration has no data path to resolution.
[TODO: calibrate the bound against the accumulated series once enough
observations exist — not yet done; the series does not yet exist.]

### 4.5  Where observations are recorded
The weekly observations (run date, `db_last_modified`, finding count, outcome)
are appended to `docs/freshness-runlog.md` — a separate tracked file, never
merged into this procedure. Tracking it in-repo rather than operator-local is
deliberate: for a human-executed compensating control whose named weakness is
*unverifiable execution*, signed, dated, append-only commits are the auditable
evidence that the check was actually run — and that record is the control's
teeth. This runbook stays authoritative on *how to run the check*; the run-log
accumulates the *evidence*, and will be the series the § 4.4 bound is calibrated
against once it accumulates.

**Landing mechanism (interim):** `develop` is PR-only by enforced ruleset, so
each append lands through the full governed PR flow — there is no direct-commit
shortcut on a protected branch. Whether weekly observations land as
one-PR-per-week or a periodic batched commit is an open operational decision, to
be settled before the first weekly run and recorded in the run-log's own header.
This cost largely dissolves at the GHES conversion (§ 5), where the scheduled CI
job emits the observation directly. The run-log file does not yet exist; it is
created on its first weekly use, not by the PR that introduces this runbook.

## 5. Conversion to an enforced control at GHES (brief — see backlog for design)

At the air-gapped GHES target this procedure converts from human-executed to a
scheduled CI gate. The one design change that cannot be carried over silently:
**age flips from advisory to *primary*.** On a private mirror the
*stale-but-valid* case — mirror reachable, serving an old but schema-valid DB
(config frame present, old `db_last_modified`, exit 0) — is a **false-green that
frame-presence cannot catch**, because the frame *is* present. Only an age gate,
bounded by the mirror's own sync cadence, catches a frozen mirror.

Full conversion design — scheduled workflow on a self-hosted runner, `-db`
repointed at the in-perimeter mirror, the age-primary threshold, and a separate
"did the mirror sync run" monitor — is recorded authoritatively in the GHES
migration backlog (`docs/REPO_SETUP.md`, migration-tasks section). This runbook
states only the trigger and the reason, to avoid two specs that drift.

## 6. Provenance

Every interpretation cell in § 4 was verified on live govulncheck runs on
**2026-06-14** (clean tree + forced DB failure via a dead localhost port), not
derived from documentation. The check command in § 3 is the exact form verified
as a unit. Environment at verification: govulncheck `v1.3.0`, scanning toolchain
`go1.26.4`, DB `https://vuln.go.dev` with `db_last_modified` `2026-06-02`
(12 days old at verification — inside the § 4.4 advisory bound, so the
verification run itself exercised the advisory-age path without a false-positive,
a live demonstration that the bound is correctly advisory rather than a gate).
