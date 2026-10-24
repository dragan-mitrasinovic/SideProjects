# 0002. Living architecture doc read by agents

- **Status:** Accepted
- **Date:** 2026-10-24

## Context

ADRs explain *why* but are a history; reconstructing the current system from a chain of
superseded decisions is slow and error-prone. Agents need a concise picture of the
system *as it is now* at the start of every session, and it must stay current as agents
change the code.

## Decision

- `docs/ARCHITECTURE.md` describes the current state only (present tense, no history),
  citing the ADR behind each significant element.
- `AGENTS.md` at the repo root is the single agent-instruction file. It is loaded
  automatically by Copilot and most other agents, and instructs them to read
  `docs/ARCHITECTURE.md` at the start of every session.
- Updating `docs/ARCHITECTURE.md` is part of the definition of done for any change that
  affects components, data flow, significant dependencies, external services, or the
  directory layout.

## Alternatives considered

- **Architecture inlined in `AGENTS.md`** — rejected: mixes rules with facts and bloats
  the always-loaded file as the project grows.
- **Script that generates an architecture section into `AGENTS.md`** — rejected: extra
  tooling and two copies that can drift. Revisit if agents are seen skipping the read.
- **`.github/copilot-instructions.md` / `CLAUDE.md`** — rejected: tool-specific;
  `AGENTS.md` covers the tools in use. Claude Code is not used.

## Consequences

- One extra file read per session (kept short to limit context cost).
- Relies on agents following the instruction to read and update the doc; reviewing
  diffs for missing doc updates is part of normal review.
