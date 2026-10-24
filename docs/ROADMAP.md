# Roadmap

> A living document: milestones change as we learn. The decisions behind them are in
> [ADR-0003](adr/0003-prototype-architecture.md).
>
> **Current milestone:** M0

| | Milestone | Done when |
|---|---|---|
| **M0** | Hosting skeleton | The platform service serves its own page and each app's static files by subdomain. That works locally on `*.localhost`, and on the VM behind Caddy with HTTPS |
| **M1** | Prompt to static site | A prompt produces a working app at `<app>.<domain>`, and a follow-up prompt changes it |
| **M2** | Stateful apps | Apps can save and share data. We design this when we get there |
| **M3** | Friends | Friends can sign in and send prompts safely, within daily limits |

## M0: hosting skeleton (no AI)

Code to write:

- **A skeleton for the platform service,** configured through environment variables: domain, port and
  data folder.
- **Routing by host:**
  - the apex domain serves a placeholder platform page;
  - `<app>.<domain>` serves files from `/srv/apps/<app>/`.

  App names must match `[a-z0-9-]`. Paths can't escape the app's folder, unknown paths fall back to
  `index.html`, and unknown apps get a 404.
- **An endpoint for Caddy's on-demand certificates,** `/tls-allowed?domain=…`. It's reachable only from
  the VM itself, and approves only the apex and existing apps.
- **Deploy files:** a Caddyfile, a systemd unit and a deploy script.

To develop locally, open `http://hello.localhost:<port>`. Browsers send `*.localhost` to your own
machine, so you don't need a domain or the VM.

## M1: prompt to static site

Code to write:

- **The app template:** Vite, React, TypeScript and Tailwind, plus an `AGENTS.md` with the rules for the
  agent. Its dependencies are installed once and shared by all apps.
- **SQLite tables** for apps and jobs.
- **The job runner,** one job at a time with a 15-minute timeout. Each job:
  1. prepares the app's folder (copying the template, or reusing the app's git repo);
  2. runs the Copilot CLI and streams its events into the job log;
  3. rebuilds the app, and rejects the job if `package.json` changed;
  4. publishes `dist/` (regular files only) and commits, or rolls back with git on failure;
  5. records the credits used.
- **The UI and API:**
  - pages: an app list, "new app" (a name plus a prompt), follow-up prompts, and a live job log;
  - the API accepts JSON only and checks the `Origin` header.

## M2: stateful apps

To be designed when we get there.

## M3: friends

- A real login to replace the password prompt.
- Sandboxing the commands Copilot runs: the CLI's built-in sandbox, a container, or the SDK with our own
  tools.
- How friends' jobs use Copilot: you approve them, friends use their own Copilot, or a paid API key.
- A daily job limit per friend.

## Later, not scheduled

- Data for apps: key-value storage, collections, files, realtime, or a database per app.
- Full-stack apps with their own server code, running in containers that sleep when idle.
- Preview environments, rollback in the UI, and Caddy serving static files directly.
- A separate domain for apps, and Cloudflare in front.
- Backups, monitoring, and infrastructure-as-code.
