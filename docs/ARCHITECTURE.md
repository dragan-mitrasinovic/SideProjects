# Architecture

> Current state only. For *why* things are the way they are, see the
> [decision records](adr/README.md).
>
> Last updated: 2026-10-31

## Overview

SideProjects is a single-project monorepo for an **app generator**: a self-hosted platform where a prompt
produces a generated, deployed, public static web app at `<app>.<PLATFORM_DOMAIN>`. The prototype design
is [ADR-0003](adr/0003-prototype-architecture.md). The platform service is a single Rust binary that
drives the Copilot CLI (ADR-0003). Progress is tracked in
[`ROADMAP.md`](ROADMAP.md). No application code exists yet. So far the repo holds project guidelines,
decision records and agent tooling.

## Components

- **Agent instructions** — [`AGENTS.md`](../AGENTS.md), loaded automatically by coding
  agents every session. (ADR-0002)
- **Architecture doc** — this file; living description of the current system. (ADR-0002)
- **Decision records** — [`docs/adr/`](adr/README.md), append-only history of
  architectural decisions. (ADR-0001)
- **Roadmap** — [`docs/ROADMAP.md`](ROADMAP.md), the living list of milestones and the code each one
  needs. (ADR-0003)

## Directory layout

- `AGENTS.md` — agent instructions
- `docs/ARCHITECTURE.md` — this file
- `docs/ROADMAP.md` — milestones
- `docs/adr/` — architecture decision records + index
