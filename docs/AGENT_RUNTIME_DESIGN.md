# Agent Runtime Security Design — Lethal-Trifecta Isolation

**Version:** 1.0 · **Status:** DESIGN — implemented incrementally as runtime
code lands. No agent runtime exists yet; nothing in this document is currently
enforced by code. Each principle below carries an *Implementation gate*
stating when it must become enforced and how that enforcement is verified.

**Source:** OpenSSF LFEL1012 (Secure AI/ML-Driven Software Development).
**Cross-reference:** `docs/THREAT_MODEL.md` §2 (future runtime boundaries —
this document is the design that section forward-links to). Governance for AI
*contributions to this repo* is `AGENTS.md`; this document governs the AI
*runtime this repo will build*. The two are distinct.

---

## 1. The lethal trifecta

An AI agent becomes dangerous when it simultaneously holds all three:

1. **Exposure to untrusted content** — web pages, documents, emails, tool
   outputs, or any LLM input the operator did not author.
2. **Access to private data** — credentials, internal documents, user data,
   anything whose disclosure causes harm.
3. **Ability to communicate externally** — network egress, file writes outside
   a sandbox, messages, or any channel that can carry data out.

With all three, a prompt-injection payload in untrusted content can direct the
agent to read private data and exfiltrate it. **Breaking any one leg collapses
the attack chain.** The design rule: no single agent component may ever hold
all three capabilities at once.

## 2. Isolation architecture (design)

The runtime separates the three legs into distinct components with explicit,
auditable interfaces:

- **Untrusted-content handlers** parse and summarize external input. They run
  with NO private-data access and NO egress. Output passes a defined boundary
  (typed structs, validated — allowlist, not denylist) before any other
  component consumes it.
- **Private-data accessors** read secrets/internal data. They accept input
  ONLY from trusted, validated sources — never raw untrusted content — and
  have NO direct egress.
- **Egress gateway** is the single component permitted outbound communication.
  Default-deny: every destination is allowlisted (host + port + protocol);
  every request is logged with the originating task ID for audit.

*Implementation gate:* lands with the first runtime component that touches
external content; verified by tests proving an untrusted-content handler
cannot reach credentials or open a socket.

## 3. Credential design

- **Per-tool scoped credentials.** Each tool integration gets its own
  credential, scoped to the minimum operations it performs. No shared
  "agent god-token."
- **No standing secrets.** Credentials are injected at invocation time from
  the environment/secret store and held only for the task's duration; never
  embedded in code, config files, prompts, or logs (`.gitignore` and
  GitGuardian already guard the repo side).
- **Secrets never enter model context.** Tool calls that require credentials
  perform authentication outside the LLM exchange; the model sees capability
  handles, not key material.

*Implementation gate:* lands with the first tool integration; verified by a
test asserting no credential string appears in any prompt/response log.

## 4. Sandboxing

Agent task execution runs in a constrained environment: filesystem access
limited to a per-task working directory; no ambient network (egress only via
the gateway in §2); resource limits (CPU/memory/time) to bound runaway loops.
Mechanism (container, namespace, or process-level) is an implementation
decision deferred until the runtime exists — the requirement is the
constraint, not the technology.

*Implementation gate:* lands before the first agent task executes
user-supplied or model-generated instructions.

## 5. What this document is not

- Not enforced today — there is no runtime. Treating it as a control before
  code exists would be the false confidence this project's method forbids.
- Not a threat model — `docs/THREAT_MODEL.md` owns risk analysis; this owns
  the runtime design that mitigates the AI-specific class of it.
- Not contributor governance — `AGENTS.md` owns that.

## Revision log

| Version | Date | Change |
|---------|------|--------|
| 1.0 | 2026-06-10 | Initial design (Phase 2.3, OpenSSF hardening plan). |
