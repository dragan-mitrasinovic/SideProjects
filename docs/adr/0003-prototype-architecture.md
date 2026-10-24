# 0003. Prototype architecture

- **Status:** Proposed
- **Date:** 2026-10-24

## Context

The project is a for-fun, self-hosted app generator: you type a prompt on a website and get back a
generated, deployed web app. Later a few friends may use it too. Usage will be very low, and it isn't
meant to make money.

The constraints:

- **One Azure VM** (Standard_B4as_v2: 4 burstable vCPUs, 16 GiB RAM, no nested virtualization), paid from
  a $150/month Azure credit. The VM, its disk and its public IP use most of that credit, so the design
  must not add Azure resources.
- **A personal GitHub Copilot account** with a large allowance generates the code. Copilot bills per token,
  in AI credits.
- **Simplicity first:** build the smallest prototype that works, and decide the rest as we go.

## Decision

We will build the prototype as **one platform service on the VM** that generates and hosts **static web
apps**.

- **Apps are static sites, for now.** Each app starts from a fixed template (Vite, React, TypeScript,
  Tailwind) and is built into plain files. No generated code runs on the server.
- **One platform service serves everything.** The platform UI and API live at the apex domain, and each
  app at `<app>.<domain>`. The service reads the `Host` header to pick the app's folder,
  `/srv/apps/<app>/`. A wildcard DNS record points every subdomain at the VM, so creating an app needs
  only a folder and a database row: no DNS or proxy change.
- **Caddy sits in front.** It gets Let's Encrypt certificates on demand, after asking the platform service
  whether a hostname is a real app. It also asks for a password (basic auth) until friends get a real
  login.
- **Generation runs the GitHub Copilot CLI as a command** (`copilot -p`) in the app's working folder:
  - as a low-privilege user that can only write to that folder;
  - with a fine-grained token limited to Copilot requests;
  - with web access and the GitHub tools turned off, and a credit cap per job;
  - one job at a time, with a timeout.

  Afterwards, the platform rebuilds the app itself, publishes `dist/` (regular files only) and commits to
  the app's git repo. A job that fails, or that changes `package.json`, is rolled back with git.
- **SQLite** holds the platform's bookkeeping: apps and jobs.
- **Caddy and the platform service run directly on the VM** as systemd services, updated by a deploy
  script. No containers, no PaaS.
- **We deliver in milestones**, tracked in [`ROADMAP.md`](../ROADMAP.md).

## Alternatives considered

- **Cloudflare in front (Tunnel and Access): not now.** It adds an external dependency, decrypts traffic
  in the middle, and limits uploads and request time. Its main benefit is login without writing code,
  which doesn't matter while only the owner uses the platform. It can be added later without a redesign.
- **A PaaS (Coolify, Dokploy, Dokku): not now.** A static app is just a folder; building an image and
  running a container for each one is slower and heavier. The platform also needs direct access to the
  VM. Revisit when apps need databases or servers.
- **The Copilot SDK with our own agent tools: not now.** It's more code for the same result. Revisit it,
  or the CLI's built-in command sandbox, before friends can send prompts (see Consequences).
- **A container per app, or a sandbox container per job: not now.** Static apps run nothing on the
  server, and only the owner writes prompts. Containers (and gVisor, which works without KVM) come back
  for full-stack apps, or before friends can send prompts.
- **Postgres, a NoSQL database, or a database per app: not now.** Static apps have no data. This gets
  decided when stateful apps are designed.
- **Firecracker or Kata microVMs: rejected.** They need KVM, which this VM doesn't support.
- **GitHub Spark or GitHub Models: rejected.** Both were retired in 2026.

## Consequences

- There's one service to write and run, almost no cost per app, and publishing an app is a folder copy.
- Ports 80/443 are open, so the VM is exposed to the internet directly.
- **A known gap:** commands that Copilot runs execute as the same user as the CLI, so a determined
  command could read the token. That's acceptable while only the owner writes prompts. It must be closed
  before friends can send prompts.
- All apps share one registrable domain, so browsers treat them as the same site. The platform API must
  accept JSON only and check the `Origin` header, so that app pages can't drive it.
- Basic auth asks for the password once per app subdomain.
- The first visit to a new app waits a few seconds for its certificate. Let's Encrypt allows 50 new
  certificates per registered domain per week.
- Still open: the domain (a free deSEC `dedyn.io` subdomain or a real domain), the platform service's
  language, and the Copilot plan.
