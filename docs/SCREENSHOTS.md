# Screenshots — Capture Guide

The main `README.md` links to the images in this folder. This file explains exactly what to
capture so the README renders with real product screenshots rather than placeholders.

## Why these are placeholders

Screenshots are binary artefacts best produced from a running instance rather than committed blind.
Each slot in `docs/screenshots/` currently holds a **placeholder SVG** that renders cleanly
anywhere, showing the route and the intended content — so the README is presentable as-is and
simply gets better when real captures replace them.

## Capture environment

For consistent, presentable screenshots, use:

| Setting | Value |
| --- | --- |
| Browser | Chrome or Edge, latest, with DevTools device toolbar |
| Viewport | 1440 × 900 (desktop) and 390 × 844 (mobile) |
| Zoom | 100% |
| Theme | Light |
| Seed data | `make migrate-up` then `go run ./cmd/seed` |
| Signed in as | `admin@campuscare.test` (admin views), `student@campuscare.test` (member views) |
| Password | `ChangeMe123!` |

Sign in at `http://localhost:3000/login` after `npm run dev` in `frontend/`.

## Images the README expects

| Filename | Route | What to capture |
| --- | --- | --- |
| `login` | `/login` | The split-screen sign-in page, brand panel visible on the left |
| `dashboard-member` | `/dashboard` | Member dashboard — greeting, 3 stat tiles, AI insight card |
| `dashboard-admin` | `/dashboard` | Same route as an admin — "Campus Overview" with period-over-period deltas |
| `discover` | `/discover` | Events tab with category filter chips and register buttons |
| `community` | `/community` | Discussions tab with a thread expanded and replies visible |
| `campus-services` | `/campus` | Services grid grouped by resource type |
| `career` | `/career` | Opportunities tab, split Open / Closed |
| `ai-assistant` | `/ai` | Chat mid-stream — show the provider badge and a partial response |
| `notifications` | `/notifications` | Unread notifications with type icons |
| `profile` | `/profile` | Personal tab with the editable name field |
| `admin-operations` | `/admin/operations` | Service request queue with Start work / Resolve |
| `admin-analytics` | `/admin/analytics` | KPI tiles and events-by-category breakdown |
| `admin-approvals` | `/admin/approvals` | Pending club verification list |
| `mobile-dashboard` | `/dashboard` | Mobile viewport showing the bottom navigation bar |

## Tips for a good demo capture

- **Collapse the AI model download.** A 4 GB progress bar in a screenshot looks like a bug. Run
  the model once, then capture the chat after it reports ready.
- **Seed first.** An empty dashboard makes for a weak slide. Run `go run ./cmd/seed` so there are
  clubs, events and opportunities to show.
- **Use the organiser account** for organiser API screenshots, and the admin account for the
  admin consoles — member accounts get redirected.
- **Blur nothing.** The seed data is explicitly fictional placeholder content, so there is no
  personal information in a seeded instance to redact.
- Prefer a light background when embedding into a slide deck.

## Replacing the placeholders

Each placeholder is `docs/screenshots/<name>.svg`. To swap in a real capture:

1. Save the screenshot as `docs/screenshots/<name>.png`
2. Delete `docs/screenshots/<name>.svg`
3. Update the path in the root `README.md`

Steps 2 and 3 can be done in one pass across every screenshot:

```bash
# from the repository root
for f in docs/screenshots/*.svg; do
  base="${f%.svg}"
  [ -f "$base.png" ] && rm "$f"
done
sed -i 's#\(docs/screenshots/[a-z0-9-]*\)\.svg#\1.png#g' README.md
```

If you prefer to keep SVG slots, simply re-save each capture over the placeholder at the same path
instead — the README needs no changes at all.

## Diagrams

Architecture, ER, request-flow, AI-pipeline and user-journey diagrams are committed SVGs in
`docs/`, so they render on GitHub, in VS Code's Markdown preview and in most Markdown viewers
with no tooling.