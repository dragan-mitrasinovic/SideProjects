# 0005. Web fetching on for generation runs

- **Status:** Accepted
- **Date:** 2026-11-04
- **Supersedes:** [0003](0003-prototype-architecture.md) (web fetching only)

## Context

ADR-0003 runs the Copilot CLI with web fetching turned off. It gave no reason. The likely ones were:

- fetched pages are untrusted text that can carry prompt injection;
- the agent runs with `--allow-all-tools`, with no one approving its commands;
- the agent runs as the same user that holds the Copilot token.

Only the owner sends prompts for now. The owner accepts this risk and wants the agent to be able to
read current documentation, for example recent Vite or Tailwind changes, instead of relying only on
what the model already knows.

The CLI separates tool permissions from URL permissions: `--allow-all-tools` doesn't cover URL access
(`--allow-all` means tools, paths and URLs). With `--no-ask-user`, a fetch to a URL that wasn't
allowed is denied without a prompt.

## Decision

We will run generation jobs with web fetching **on**: the runner passes `--allow-all-urls` and no
longer excludes the `web_fetch` tool. All other flags from ADR-0003 stay the same, including the
built-in GitHub MCP server being turned off and path checks staying on.

## Alternatives considered

- **Keep web fetching off (ADR-0003):** rejected. The owner prefers better results to the reduced
  prompt-injection risk while only the owner sends prompts.
- **Allow only a few documentation sites (`--allow-url=…`):** not now. It's safer, but the list
  needs maintaining, and the agent fails quietly on any site that isn't on it.

## Consequences

- The agent can read current docs and examples, at the cost of extra credits and time when it does.
- Fetched pages can try to steer the agent (prompt injection). With `--allow-all-tools`, a page could
  get the agent to run commands, including reading `COPILOT_GITHUB_TOKEN` or sending data out. This
  widens the known gap in ADR-0003, and it must be closed before friends can send prompts (M3
  sandboxing).
- Runs depend on outside sites when the agent chooses to fetch. The opt-in smoke test fetches
  example.com, so it also checks that fetching keeps working.
