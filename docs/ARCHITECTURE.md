# Architecture

> Current state only. For *why* things are the way they are, see the
> [decision records](adr/README.md).
>
> Last updated: 2026-10-24

## Overview

SideProjects is a single-project monorepo. No application code exists yet; only
project guidelines and agent tooling are in place.

## Components

- **Agent instructions** — [`AGENTS.md`](../AGENTS.md), loaded automatically by coding
  agents every session. (ADR-0002)
- **Architecture doc** — this file; living description of the current system. (ADR-0002)
- **Decision records** — [`docs/adr/`](adr/README.md), append-only history of
  architectural decisions. (ADR-0001)

## Directory layout

- `AGENTS.md` — agent instructions
- `docs/ARCHITECTURE.md` — this file
- `docs/adr/` — architecture decision records + index
