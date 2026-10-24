# 0001. Record architecture decisions

- **Status:** Accepted
- **Date:** 2026-10-24

## Context

Much of this project will be built with AI coding agents. Agents start every session
without memory, so the *reasons* behind earlier choices are lost unless written down.
Without them, agents (and future me) re-propose rejected options or silently undo
deliberate decisions. When a decision is changed, it should still be possible to see
what it replaced and why.

## Decision

We will record architectural decisions as Architecture Decision Records in `docs/adr/`:

- One decision per file, numbered sequentially: `NNNN-kebab-case-title.md`, using a
  Nygard-style template extended with an **Alternatives considered** section
  ([`0000-template.md`](0000-template.md)).
- Status lifecycle: `Proposed → Accepted → (Deprecated | Superseded)`.
- An ADR is a work in progress, freely editable, until it is committed and pushed.
  After that it is immutable: a changed decision gets a new ADR with
  `Supersedes: NNNN`, and the old ADR's status line becomes `Superseded by NNNN`. That
  status line is the only edit ever made to a pushed ADR.
- [`README.md`](README.md) is an index of every ADR and its status, so agents can scan
  it cheaply instead of reading every file.
- Agents may draft ADRs as `Proposed`; they become `Accepted` only after explicit user
  approval.

## Alternatives considered

- **Plain Nygard template** — rejected: lacks a place for rejected options, which is
  what stops agents from re-suggesting them.
- **Full MADR template** — rejected: more ceremony than a side project needs, and longer
  files cost more agent context.
- **Single decision-log file** — rejected: grows unbounded, harder to link to individual
  decisions, and invites rewriting history in place.
- **adr-tools CLI** — not adopted for now; agents can follow the conventions directly.
  Can be added later without changing the format.

## Consequences

- Every significant decision costs a short write-up.
- The full chain of why a decision changed is traceable through supersede links.
- Agents must check the index before contradicting an existing decision (see `AGENTS.md`).
