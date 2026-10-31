# 0003. Prototype architecture

- **Status:** Accepted
- **Date:** 2026-10-24

## Context

The project is a for-fun, self-hosted app generator: you type a prompt on a website and get back a
generated, deployed web app, like Lovable. Later a few friends may use it too. Usage will be very
low, and it isn't meant to make money.

The constraints:

- **One Azure VM** running Ubuntu (Standard_B4as_v2: 4 burstable vCPUs, 16 GiB RAM, no nested
  virtualization), paid from a $150/month Azure credit. The VM, its disk and its public IP use most of
  that credit, so the design must not add Azure resources.
- **A personal GitHub Copilot account** with a large allowance generates the code. Copilot bills per
  token, in GitHub AI Credits (1 credit = $0.01).
- **The owner's own domain** hosts the platform.
- **The owner wants to write the service in Go.**
- **Simplicity first:** build the smallest prototype that works, and decide the rest as we go.

## Decision

We will build the prototype as **one Go platform service on the VM** that generates and hosts
**public static web apps**.

- **Apps are static sites, for now.** Each app starts from a fixed template (Vite, React, TypeScript,
  Tailwind) and is built into plain files. No generated code runs on the server. Apps are public:
  anyone with the URL can open them.
- **One platform service serves everything.** The platform UI and API live at a dedicated subdomain
  of the owner's domain (`PLATFORM_DOMAIN`, e.g. `apps.example.com`), and each app lives at
  `<app>.<PLATFORM_DOMAIN>`. The service reads the `Host` header to pick the app's folder. A wildcard
  DNS record points every app subdomain at the VM, so creating an app needs only a folder and a
  database row: no DNS or proxy change.
- **The service is written in Go, using the standard library wherever it's enough:**
  - `net/http` for HTTP, with the `ServeMux` routing patterns from Go 1.22;
  - `http.FileServerFS` over an `os.Root` for app files, so paths and symlinks can't escape the
    app's folder, falling back to `index.html`;
  - `modernc.org/sqlite` (pure Go, no cgo) through `database/sql`;
  - `os/exec` with `exec.CommandContext` for `copilot`, `npm` and `git`;
  - config from `os.Getenv`: `PLATFORM_DOMAIN`, `PLATFORM_PORT` and `PLATFORM_DATA_DIR`;
  - logs to stderr with `log/slog`.

  The platform UI is plain HTML and JS, embedded in the binary with `embed`.
- **Caddy sits in front.** It gets Let's Encrypt certificates on demand, after asking the platform service
  whether a hostname is a real app. It asks for a password (basic auth) on the platform host only,
  until friends get a real login.
- **Generation runs the GitHub Copilot CLI in programmatic mode** (`copilot -p`) in the app's working
  folder:
  - as a low-privilege user that can only write to that folder;
  - with `--no-ask-user` and `--allow-all-tools`, with web fetching and the built-in GitHub MCP server
    turned off;
  - with `--max-ai-credits` as a per-job credit cap;
  - with `--output-format=json`, whose JSONL events go into the job log;
  - one job at a time, with a timeout. Each follow-up prompt is a fresh run in the same folder.

  The CLI signs in as the owner with a fine-grained personal access token that has only the
  **Copilot Requests** permission. The token is passed as `COPILOT_GITHUB_TOKEN`, from a systemd
  `EnvironmentFile` that only root can read. Afterwards, the platform rebuilds the app itself,
  publishes `dist/` (regular files only) and commits to the app's git repo. A job that fails, or
  that changes `package.json`, is rolled back with git.
- **SQLite** holds the platform's bookkeeping: apps and jobs.
- **Caddy and the platform service run directly on the VM** as systemd services. A deploy script
  builds a static binary on the development machine (`CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build`)
  and copies it to the VM, so the VM needs no Go toolchain. No containers, no PaaS.
- **We deliver in milestones**, tracked in [`ROADMAP.md`](../ROADMAP.md).

## Alternatives considered

- **Another language (Rust, TypeScript): rejected.** The owner wants Go: it's quicker to write than
  Rust, and its standard library covers HTTP, processes, JSON and embedding.
- **A web framework (Gin, Echo, chi), an ORM or query builder (GORM, sqlc), or a frontend framework:
  not now.** They're more machinery than two tables and a few pages need.
- **`mattn/go-sqlite3`: rejected.** It needs cgo, which rules out a static cross-compiled binary.
- **The apex domain for the platform: rejected.** A dedicated subdomain keeps the wildcard DNS record,
  on-demand certificates and generated pages apart from the rest of the domain.
- **Apps behind the password: rejected.** Static apps run nothing on the server, and basic auth would
  ask for the password again on every app subdomain.
- **Cloudflare in front (Tunnel and Access): not now.** It adds an external dependency, decrypts traffic
  in the middle, and limits uploads and request time. Its main benefit is login without writing code,
  which doesn't matter while only the owner uses the platform. It can be added later without a redesign.
- **A PaaS (Coolify, Dokploy, Dokku): not now.** A static app is just a folder; building an image and
  running a container for each one is slower and heavier. The platform also needs direct access to the
  VM. Revisit when apps need databases or servers.
- **The Copilot SDK for Go (`github.com/github/copilot-sdk/go`), with our own agent tools: not now.** It runs the same
  Copilot runtime as the CLI, with more code. Revisit it, or the CLI's built-in command sandbox, before
  friends can send prompts (see Consequences).
- **The Copilot coding agent and its Agent Tasks API: rejected.** They need Copilot Business or
  Enterprise and a GitHub repo per app, and they run in GitHub's cloud.
- **Calling Copilot's backend API directly: rejected.** It's undocumented and unofficial.
- **A container per app, or a sandbox container per job: not now.** Static apps run nothing on the
  server, and only the owner writes prompts. Containers (and gVisor, which works without KVM) come back
  for full-stack apps, or before friends can send prompts.
- **Postgres, a NoSQL database, or a database per app: not now.** Static apps have no data. This gets
  decided when stateful apps are designed.
- **Firecracker or Kata microVMs: rejected.** They need KVM, which this VM doesn't support.
- **GitHub Spark or GitHub Models: rejected.** Both were retired in 2026.

## Consequences

- There's one static Go binary to write and run, almost no cost per app, and publishing an app is a
  folder copy.
- The development machine needs a supported Go release (1.27 at the time of writing). The VM needs only
  Node.js 22 or newer, for Vite and the Copilot CLI (`npm install -g @github/copilot`).
- Ports 80/443 are open, so the VM is exposed to the internet directly. Anything a generated app
  contains is public.
- **A known gap:** commands that Copilot runs execute as the same user as the CLI, so a determined
  command could read the token. That's acceptable while only the owner writes prompts. It must be closed
  before friends can send prompts.
- **Friends must not use the owner's Copilot account.** GitHub's terms allow a login to be used by
  only one person. Friends' jobs must use their own Copilot token, or a model provider's API key
  (BYOK).
- `--max-ai-credits` is a soft cap in public preview (minimum 30). The job timeout is the hard stop,
  and a budget on the GitHub account is the final backstop. The token expires and must be rotated.
- All apps share one registrable domain, so browsers treat them as the same site. The platform API must
  accept JSON only and check the `Origin` header, so that app pages can't drive it.
- The first visit to a new app waits a few seconds for its certificate. Let's Encrypt allows 50 new
  certificates per registered domain per week, shared with the rest of the owner's domain.
