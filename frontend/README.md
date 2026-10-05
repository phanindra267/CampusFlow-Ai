# CampusCare AI — Frontend

Next.js 16 App Router frontend for the CampusCare AI campus platform at Lovely Professional
University. See the [root README](../README.md) for the full system documentation.

## Stack

| Technology | Version | Notes |
| --- | --- | --- |
| Next.js | 16.3.8 | App Router, Turbopack |
| React | 19.2.8 | All pages are client components |
| TypeScript | 5.x | `strict` mode |
| Tailwind CSS | 4.x | Tokens via `@theme inline` in `globals.css` |
| Vitest | 3.2 | Unit tests |
| ESLint | 9 | `eslint-config-next` |
| `@mlc-ai/web-llm` | 0.2.85 | In-browser LLM inference |

## Getting Started

```bash
npm install
npm run dev        # → http://localhost:3000
```

The frontend talks to the API at `NEXT_PUBLIC_API_URL`, defaulting to
`http://localhost:8080/api/v1`. The backend must be running for anything but the login screen to
load data.

## Scripts

```bash
npm run dev        # Dev server with Turbopack
npm run build      # Production build
npm start          # Serve the production build
npm run lint       # ESLint
npm test           # Vitest
npx tsc --noEmit   # Typecheck
```

Or from the repo root: `make frontend-install`, `make frontend-dev`, `make frontend-build`,
`make frontend-lint`, `make frontend-typecheck`.

## Environment

Create `frontend/.env.local`:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1

# On-device generation (WebLLM on WebGPU)
NEXT_PUBLIC_WEBLLM_MODEL=Qwen2.5-0.5B-Instruct-q4f16_1-MLC
NEXT_PUBLIC_WEBLLM_CONTEXT_WINDOW=4096
NEXT_PUBLIC_WEBLLM_TEMPERATURE=0.7
NEXT_PUBLIC_WEBLLM_TOP_P=0.95
NEXT_PUBLIC_WEBLLM_MAX_TOKENS=512
```

There is no provider selection variable. WebLLM is the only generative runtime, and nothing is
downloaded until a request actually needs generation. Structured requests and search never touch the
model.

## Routes

| Route | Purpose |
| --- | --- |
| `/` | Redirect shim → `/dashboard` or `/login` |
| `/login` | Split-screen sign-in |
| `/dashboard` | Role-split home: member greeting, or admin campus overview |
| `/discover` | Events and groups with category filters |
| `/community` | Discussions and groups; deep-links via `?discussion=<id>` |
| `/campus` | Campus services, resource booking, booking management |
| `/career` | Opportunities, applications, interview preparation |
| `/research` | Open projects and research groups |
| `/ai` | AI assistant: deterministic routing, hybrid retrieval, on-demand WebLLM generation |
| `/notifications` | Notification centre with optimistic mark-as-read |
| `/profile` | Personal details, settings, privacy |
| `/admin/users` | Member administration |
| `/admin/analytics` | Campus analytics |
| `/admin/approvals` | Club verification |
| `/admin/operations` | Service desk queue |

Everything except `/` and `/login` lives in the `(protected)` route group, whose layout redirects
to `/login` when there is no session.

## Architecture Notes

**Auth uses an external store.** `contexts/AuthContext.tsx` implements `useSyncExternalStore`
over a `localStorage`-backed store with a cached snapshot, rather than plain React state. This is
correct under concurrent rendering, and a `window` `storage` listener propagates sign-out across
tabs. Keys: `campuscare_user`, `campuscare_access_token`, `campuscare_refresh_token`.

**`lib/api.ts` handles token refresh transparently.** On any `401` with a refresh token present it
calls `POST /auth/refresh`, stores the rotated pair, and retries the original request exactly once.
An in-flight refresh promise is shared across concurrent requests so parallel page loads cannot
race and burn the refresh token. On unrecoverable refresh failure the session is cleared through a
registered handler, avoiding a circular import.

**`hooks/useAsyncData.ts` is the single fetch hook.** Returns
`{ data, error, isLoading, refresh, setData }`, guards against `setState` after unmount, and
re-runs when `deps` change. Exposing `setData` enables the optimistic updates used by
mark-as-read, join/leave, booking cancellation and admin approvals — each rolls back via `refresh()`
on failure.

**`lib/ai/router.ts` classifies every request before any model is involved.** A regex/keyword/phrase
match returns one of three outcomes:

| Outcome | Handling | Model |
| --- | --- | --- |
| `structured` | The page calls the corresponding existing endpoint (`api.events`, `api.clubs`, `api.opportunities`, `api.communityResources`, …), so auth, RBAC and audit behaviour are unchanged | none |
| `search` | `api.aiContext()` → `GET /ai/context` on the API, which fuses BM25 and Weaviate results | none |
| `generative` | The same retrieval runs first for grounding, then `hooks/useLLM` generates a reply | WebLLM |

Keeping routing in code rather than in a model is what makes "show me events" fast and free.

**`hooks/useLLM.ts` wraps the engine runtime** into `{ status, model, progress, error, isStreaming,
initialize, generate, stop }`. `generate` consumes the engine's async generator and invokes `onDelta`
per token; `stop` aborts the in-flight request. The engine is a module-level singleton, so repeated
questions reuse one loaded model, and a failed load can be retried.

**The WebLLM runtime** (`lib/llm/`) runs inference in a dedicated worker:

- `config.ts` — model id and sampling settings, with validation
- `protocol.ts` — the typed main-thread ↔ worker message contract
- `engine.ts` — worker lifecycle, single-flight initialisation, streaming, cancellation, disposal
- `webllm.worker.ts` — the only file that imports `@mlc-ai/web-llm`, so the library is fetched lazily
  and never on the server
- `capabilities.ts` — WebGPU detection, so a browser without it gets an explanation instead of a
  crash

Because a worker cannot be unit-tested without a real WebGPU adapter, the runtime is covered by
`capabilities.test.ts`; the routing logic is covered by `router.test.ts`.

## Styling

Tailwind v4 with design tokens declared in `globals.css` via `@theme inline`, mapped to CSS custom
properties in `:root`. Brand colours: LPU orange `#F07C00` for primary actions, indigo `#6366F1`
reserved for AI surfaces.

`globals.css` also contains a global interaction layer built on element selectors — button hover
lift, gradient shift, focus rings via `color-mix`, sidebar nav transforms, card hover borders — plus
keyframe utilities (`.ai-shimmer`, `.skeleton`, `.page-transition`) and a complete
`prefers-reduced-motion: reduce` block that disables every animation and transform.

> `tailwind.config.ts` is retained for the v3-style `content` glob and `darkMode` setting, but v4
> does the token work in CSS. Note that dark mode is **scaffolded but not yet functional** — the
> `.dark` token block exists, but the class is never applied and pages hardcode light utilities.

## Accessibility

Global `:focus-visible` ring; semantic `<label htmlFor>` on all inputs; `aria-label` on icon-only
controls; `role="alert"` on error banners; status conveyed by text as well as colour; `en-IN` date
formatting; Enter-to-submit with Shift+Enter for newlines in the chat composer.

## Known Gaps

- Dark mode is scaffolded but inactive (see Styling above)
- The `⌘K` hint in the top bar is decorative — no keyboard handler is bound, and the command
  palette has no arrow-key navigation or `Esc` dismissal
- Profile notification and privacy checkboxes are uncontrolled and not persisted
- Test coverage is `api.test.ts`, `ai/router.test.ts`, `llm/capabilities.test.ts` and
  `ai/privacy.test.ts`; there are no page or component tests, and the WebLLM worker itself is only
  exercised manually because it needs a real WebGPU adapter
- `lucide-react` is a declared but unused dependency — icons are inline SVG components
- Admin pages are gated client-side by `AdminOnly` for UX; real authorization is enforced by the API