# Agent Instructions

## Every session

1. Read [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) before doing anything else.
2. Skim the ADR index at [`docs/adr/README.md`](docs/adr/README.md) and read any ADR
   relevant to your task.

## Architecture doc

- `docs/ARCHITECTURE.md` describes the system **as it is now**. Present tense, current
  state only — history and rationale belong in ADRs.
- Update it in the same change whenever you add, remove, or rename a component, change
  data flow, add a significant dependency or external service, or change the directory
  layout. A task is not done until the doc matches the code.
- Reference the ADR that justifies each significant element, e.g. `(ADR-0004)`.
- Keep it short and skimmable; it is read at the start of every session.

## Architecture decisions (ADRs)

- ADRs live in `docs/adr/NNNN-kebab-case-title.md`, numbered sequentially, created from
  [`0000-template.md`](docs/adr/0000-template.md). Add a row to the index for every ADR.
- Write an ADR when a choice is hard to reverse or affects more than one module:
  languages/frameworks, storage, major dependencies, API/protocol shape, project
  structure, deployment, auth, cross-cutting patterns. Not for bug fixes, local
  refactors, naming, or minor libraries.
- Create new ADRs with `Status: Proposed`. Change it to `Accepted` only after the user
  explicitly approves that ADR.
- While an ADR is uncommitted or unpushed it is a work in progress and may be edited
  freely as the feature evolves.
- Once an ADR has been committed **and pushed**, it is immutable. To change that
  decision, write a new ADR with `Supersedes: NNNN`, change only the old ADR's status
  line to `Superseded by NNNN`, and update both rows in the index — all in the same
  change.
- Before making a change that contradicts an Accepted ADR, stop and ask the user. Do not
  re-propose an alternative an ADR already rejected unless you have new information.

## Git boundaries

The user reviews all work as **uncommitted changes** in the working tree, then commits
and pushes themselves.

- **Allowed anytime:** read-only git — `status`, `diff`, `log`, `show`, `blame`,
  `branch` (listing), `fetch`, `rev-parse`, `ls-files`, `stash list`, etc.
- **Never commit or push unless the user explicitly asks in the current conversation.**
  The same applies to other history- or remote-changing operations (`merge`, `rebase`,
  `tag`, `pull`, branch create/switch/delete, `gh pr create`/`merge`). Permission covers
  that one request only and does not carry over to later changes. "Implement X" or
  "make the change" is not permission to commit.
- **Never discard or hide uncommitted work** (`stash`, `reset --hard`,
  `checkout -- <path>`, `restore <path>`, `clean`, force-push). If one seems necessary,
  explain why and let the user run it.

## Finishing a task

End with a short summary that states the changes are **left uncommitted**, lists the
changed files, and calls out any ADR created or `docs/ARCHITECTURE.md` update.
