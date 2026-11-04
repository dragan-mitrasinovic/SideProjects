# Architecture

> Current state only. For *why* things are the way they are, see the
> [decision records](adr/README.md).

## Overview

SideProjects is a single-project monorepo for an **app generator**: a self-hosted platform where a prompt
produces a generated, deployed, public static web app at `<app>.<PLATFORM_DOMAIN>`. The prototype design
is [ADR-0003](adr/0003-prototype-architecture.md). The platform service is a single Go binary that
drives the Copilot CLI (ADR-0003). Progress is tracked in
[`ROADMAP.md`](ROADMAP.md). The Go module lives in `platform/`. So far it holds only the Copilot runner;
the HTTP service and job runner don't exist yet.

## Components

- **Copilot runner** — [`platform/internal/copilot`](../platform/internal/copilot/copilot.go):
  `Runner.Run` runs one `copilot -p` process in an app's folder and copies its JSONL stdout to an
  `io.Writer` as it arrives. It refuses to start without a token, and passes only `PATH`, `HOME` and
  `COPILOT_GITHUB_TOKEN` to the process. Flags: no prompts, all tools, all URLs (web fetching on),
  built-in MCP servers off, a 1000-credit soft cap, no auto-update. The process runs in its own
  process group, which is killed on cancellation and after every run. Errors include the CLI's
  stderr. The caller sets the timeout through `ctx`. Standard library only. Tests use a fake CLI; an
  opt-in smoke test (`COPILOT_SMOKE=1`) runs the real one. (ADR-0003, ADR-0005)
- **Agent instructions** — [`AGENTS.md`](../AGENTS.md), loaded automatically by coding
  agents every session. (ADR-0002)
- **Stepwise agent** — [`.github/agents/stepwise.agent.md`](../.github/agents/stepwise.agent.md)
  and the [`/step`](../.github/prompts/step.prompt.md) prompt. This is an opt-in mode that splits a
  task into a `## Steps` list in its PLAN file, implements one tiny step per `/step`, and stops for
  review after each one. (ADR-0004)
- **Architecture doc** — this file; living description of the current system. (ADR-0002)
- **Decision records** — [`docs/adr/`](adr/README.md), append-only history of
  architectural decisions. (ADR-0001)
- **Roadmap** — [`docs/ROADMAP.md`](ROADMAP.md), the living list of milestones and the code each one
  needs. (ADR-0003)

## Directory layout

- `AGENTS.md` — agent instructions
- `.github/agents/`, `.github/prompts/` — custom agents and prompts
- `.github/instructions/` — file-scoped agent instructions (Go style: `go-style.instructions.md`)
- `docs/ARCHITECTURE.md` — this file
- `docs/ROADMAP.md` — milestones
- `docs/adr/` — architecture decision records + index
- `platform/` — the Go module (`github.com/dragan-mitrasinovic/SideProjects/platform`)
  - `.golangci.yml` — golangci-lint config (blank-line checks only); run `golangci-lint run ./...`
  - `internal/copilot/` — Copilot runner
