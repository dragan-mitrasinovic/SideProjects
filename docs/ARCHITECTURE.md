# Architecture

> Current state only. For *why* things are the way they are, see the
> [decision records](adr/README.md).
>
> Last updated: 2026-10-31

## Overview

SideProjects is a single-project monorepo for an **app generator**: a self-hosted platform where a prompt
produces a generated, deployed, public static web app at `<app>.<PLATFORM_DOMAIN>`. The prototype design
is [ADR-0003](adr/0003-prototype-architecture.md). The platform service is a single Go binary that
drives the Copilot CLI (ADR-0003). Progress is tracked in
[`ROADMAP.md`](ROADMAP.md). The Go module lives in `platform/`. So far it holds only the Copilot runner;
the HTTP service and job runner don't exist yet.

## Components

- **Copilot runner** — [`platform/internal/copilot`](../platform/internal/copilot/copilot.go): runs
  one `copilot -p` process in an app's folder and streams its JSONL output line by line. It passes
  only `PATH`, `HOME` and `COPILOT_GITHUB_TOKEN` to the process, starts it in its own process group,
  and kills that group on timeout, on cancellation and after the run. Standard library only. (ADR-0003)
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
- `platform/` — the Go module (`github.com/dragan-mitrasinovic/SideProjects/platform`)
  - `internal/copilot/` — Copilot runner
