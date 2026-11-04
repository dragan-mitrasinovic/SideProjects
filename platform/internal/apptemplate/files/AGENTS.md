# App rules

You are building a small web app from the user's prompt. These rules always apply.

## What to build

- The whole app is this folder's `index.html`: one self-contained file, with CSS in `<style>` and
  JavaScript in `<script>`. Don't create other files.
- Replace the placeholder page entirely on the first prompt. On later prompts, change the existing
  page to do what's asked, and keep everything else working.
- The page is served as a static file at `https://<app>.<domain>/`. There is no build step, no npm
  and no server code.
- Libraries, fonts and styles may be loaded from public CDNs with `<script>` and `<link>` tags.
- Store data only in the browser (`localStorage`). There is no backend.

## Quality

- Make it work well on phones and desktops, and keep it accessible: semantic HTML, labels on
  inputs, and enough contrast.
- The page is public. Never put secrets, tokens or personal data in it.

## How to work

- Nobody answers questions during a run. When something is unclear, make a sensible choice and keep
  going.
- Don't start servers or long-running processes. You don't need to serve the page to finish.
- Don't edit this `AGENTS.md`.
