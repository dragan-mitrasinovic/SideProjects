# 0004. Stepwise agent for from-scratch work

- **Status:** Accepted
- **Date:** 2026-11-03

## Context

The platform is being built from scratch as a hobby project. The user wants close, low-level
control over code quality: seeing each component grow one requirement at a time and reviewing
every increment before the next starts, without going back to writing everything by hand. Default
agent behaviour delivers a whole task at once, which is too coarse to follow. Established projects
don't need this, so it should not burden every session.

Prior art: Kent Beck's "augmented coding" (a `plan.md` of next tests, advanced with "go") and
Harper Reed's spec → small prompt list → `todo.md` workflow.

## Decision

We will provide an opt-in **Stepwise** custom agent (`.github/agents/stepwise.agent.md`) and a
`/step` prompt (`.github/prompts/step.prompt.md`):

- The agent first slices the task into a `## Steps` checklist inside the task's `PLAN-*.md`. Each
  step is one requirement realised as one function, one test, or one small change.
- `/step` (or "go") implements only the next unchecked step, with strict YAGNI. It runs the
  project's build, lint and test checks on the touched code as a sanity check, reports what changed
  and what was deliberately left out, and then stops for the user's human review. Passing checks
  never replace that review. A step is ticked only once the user approves it.
- Tidy-only steps are marked and kept separate from behaviour changes. Test-first is not required.
- A PLAN file with a `## Steps` list stays until its last step is done (an exception in
  `AGENTS.md`).

`AGENTS.md` remains the single always-loaded instruction file (ADR-0002). The Stepwise agent
only adds rules when it is selected.

## Alternatives considered

- **Always-on rules in `AGENTS.md`** — rejected: slows down routine work that doesn't need
  step-by-step review.
- **Mandatory TDD (Beck style)** — rejected: the user prefers tests as separate steps when
  wanted.
- **Separate `STEPS-*.md` file** — rejected: one PLAN file per task is simpler to track.
- **Hook enforcing one step per turn** — rejected for now: instructions should suffice; revisit
  if the agent keeps running ahead.
- **Spec-driven kits (spec-kit, Kiro, BMAD)** — rejected: heavyweight and aimed at speed, not at
  small reviewable increments.

## Consequences

- From-scratch work is slower but every increment is small and reviewed.
- Stopping after each step relies on the model following instructions; there is no enforcement.
- PLAN files may live across sessions while a stepwise task is in progress.
