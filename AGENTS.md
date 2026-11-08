# AGENTS.md — POPCORN (repo: removy-official, legacy names: REMOVY, Movie Hub)

Agent handoff guide. Read this first; open other docs only when needed.
Movie and TV social web app: browse (TMDB), rate, review, track, follow. Go API + Vue 3 SPA + PostgreSQL.
The product is POPCORN (formerly REMOVY). A few identifiers intentionally keep old names (see "Naming"). Don't rename them without a migration plan.

## Commands
```bash
docker compose --env-file server/.env up -d  # postgres + MinIO (:9000, console :9001)
cd server && go run main.go          # PORT from .env (8081 locally); run from server/ (relative paths); auto-migrates and seeds on boot
cd server && go test ./...           # unit tests
cd server && TEST_DATABASE_URL=postgres://user:pass@localhost:5432/postgres go test -tags integration ./integration/...  # real Postgres, localhost only; each test uses a throwaway database (CI has a postgres service)
cd client && npm run dev             # :3000, proxies /api -> localhost:8081 (VITE_API_PROXY_TARGET)
cd client && npm run type-check      # vue-tsc --build
cd client && npm test                # Vitest (src/__tests__)
cd client && npm run build           # type-check + vite build
cd client && npm run lint            # verify-locales + oxlint + eslint; both linters use --fix, so it EDITS files
cd client && npm run verify-locales  # locale check only
```
Verify before finishing: `go vet ./... && go test ./...`, `npm run type-check`, `npm test`, `npm run build-only`, and `npm run verify-locales` when locale files change. CI (`.github/workflows/ci.yml`) runs these plus `oxlint` on every push/PR. ESLint is not in CI yet because of existing errors.

## Layout (only what matters)
```
server/main.go              env load, CORS, HTTP server; sends /socket.io to the hub and everything else to Fiber
server/router/router.go     Module wiring and order. Start here for any endpoint. Also /share/proxy-image.
server/config/config.go     All env vars
server/middleware/          RequireAuth, RequireCurrentAdmin, JWT claims
server/database/
  db.go                     Connect, autoMigrate list, runSQLMigrations, seeds  <-- SCHEMA SOURCE OF TRUTH
  migration.go              runMigrationWithHistoryTx (advisory lock + timeouts; locks the reviews table)
  migrate/                  SQL used by migrations    seeder/achievement.json
server/internal/
  <domain>_module/          router -> handler -> service -> repository, plus dto.go and model.go
  privacy_policy/           Shared visibility policy
  identity_policy/          ProveEmail (the only place an email becomes verified) + session epoch checks
  shared/                   apierr.go, pgerr.go, media/user summaries, storage/ (MinIO)
  mailer/                   Brevo sender
client/src/
  api/endpoints/<domain>.ts  API calls (api/index.ts = axios instance)
  types/<domain>.ts          TS mirrors of DTOs
  views/  route pages        components/<area>/        layouts/ MainLayout, AdminLayout
  components/ui/  UI primitives, see "UI primitives" below    assets/styles/motion.css  motion + focus tokens
  stores/ auth, ui, feed, notification, achievement, admin
  composables/ useAuth, useSocket, useFollow, useReviewForm, useShareImage, ...
  i18n/locales/{en,th}/*.json  UI text, one file per area
```

## Request flow
`router/router.go -> <module>/router.go -> middleware -> handler -> service -> repository (GORM)`
- Dependencies are passed through each module's `RegisterRoutes`/`New`; wiring order lives in `router.go`.
- API prefix is `/api`. Most modules mount on the `protected` group. `auth`, `link`, `follow`, and `achievements` receive `api` and apply auth per route.
- **There is no success envelope.** Handlers return named keys such as `{"review": ...}` or `{"reviews": [...]}`. Errors are `{"error": "..."}`, or `{"code","error"}` through `shared.WriteAPIError`. Follow the style of the module you edit.
- Socket.io bypasses Fiber: `main.go` hands it directly to `notifHub.Handler()`, with its own CORS headers.

## Feature -> files map
| Feature | Backend (`server/internal/`) | Frontend |
|---|---|---|
| Auth, email verification, Google OAuth, sessions | auth_module, user_module (password reset, email change mailers) | views/auth/, components/auth/ (AuthShell, AuthField, AuthAlert, GoogleSignInButton, PasswordStrength), stores/auth.ts, composables/useAuth.ts |
| Account linking PLANZI<->POPCORN (HMAC server-to-server, handoff token in URL fragment; contract in `../LINKING_SPEC.md`; disabled when `LINK_SHARED_SECRET` is empty; login falls back to remote `verify-password`) | link_module | views/auth/HandoffView.vue (back-to-PLANZI link from public `GET /api/link/remote` -> `{"web_url"}`, 404 when linking is off), api/endpoints/link.ts |
| Movie, TV, actor, search, upcoming, where-to-watch (`GET /movies/:id/watch-providers`, `/tv/:id/watch-providers`; TMDB/JustWatch, 6h cache) | movie_module (`providers.go`), tmdb_module (`providers.go`) | views/movie/, views/tv/, views/actor/, views/SearchResultView.vue, components/detail/DetailWatchProviders.vue, composables/useWatchProviders.ts |
| Ratings display (POPCORN primary; TMDB, IMDb, Rotten Tomatoes secondary). Movie/TV lists return `ratings.{popcorn,tmdb}`; detail adds `imdb`, `rotten_tomatoes` from OMDb (`OMDB_API_KEY`, empty = off). `ExternalRating` table caches 7d (1d for not_found), 3s upstream timeout. A stale row is served at once and refreshed in the background; a cold miss waits at most 700ms, then returns unavailable while the fetch fills the cache. `imdb_id` must match `^tt\d{1,10}$` or it is not_found. The 900 req/day budget is in-memory and per instance (resets on restart; N instances allow N x 900), taken inside the singleflight call. Library, likes, BMOL and feed items carry `media.ratings` via `shared.AttachRatings` | movie_module (dto/service), external_rating_module, shared/media_ratings.go | components/movie/MediaRating.vue (`card`/`detail`), PopcornScoreBadge.vue, popcornTier.ts (tiers: none, stale <2.5, fresh 2.5-3.9, hot >=4.0), utils/mediaRatings.ts |
| Movie / TV detail pages | movie_module (GET /movies/:id, /tv/:id), tmdb_module | components/detail/ (DetailPage + detailConfig adapters; reviews/ split), composables/useMediaDetail.ts, useMediaActivity.ts; views/movie/MovieDetailView.vue and views/tv/TVDetailView.vue are thin wrappers |
| Reviews, comments, ratings | review_module, like_module, media_stats_module | components/movie/MovieReviews.vue, composables/useReviewForm.ts |
| Library lists | library_module | views/library/ |
| Watch logs | watch_log_module | api/endpoints/watchLog.ts, types/watchLog.ts |
| Best Movie of Life (BMOL) | bmol_module | views/library/components/bmol/, components/profile/UserBmol.vue |
| Achievements | achievements_module, database/seeder/achievement.json | views/achievement/, stores/achievement.ts |
| Follow, feed | follow_module, feed_module | views/FeedView.vue, components/feed/, components/social/ |
| Profiles, stats, privacy | user_module, user_stats_module, analytics_module, privacy_policy | views/user/, components/profile/ |
| Notifications, realtime | notification_module (`socket.go` = hub) | views/NotificationsView.vue, stores/notification.ts, composables/useSocket.ts |
| Share images | `shareImageProxyHandler` in router.go | components/share/, composables/useShareImage.ts |
| Admin | admin_module | views/admin/, layouts/AdminLayout.vue, stores/admin.ts |
| Onboarding, donate, about | — | views/onboarding/, views/aboutus/ |

## UI primitives (client/src/components/ui)
Use these instead of hand-rolled modals, tabs, inputs, and loading/empty/error states.
- `AppButton`: variants `primary | outline | ghost | danger`, sizes `sm | md | lg`; props `to`/`href`, `loading`, `block`, `iconOnly` (needs aria-label).
- `AppModal`: dialog with focus trap, scroll lock, and a shared modal stack (Escape closes only the top one). `profile/components/ConfirmModal` and `AdminConfirmModal` are built on it.
- `AppInput`: labeled text input with error/hint wiring (used by auth forms such as ForgotPassword).
- `AppAvatar`: image with initials fallback (`sm | md | lg | xl`).
- `AppTabs`: tablist with roving tabindex; Arrow, Home, End keys.
- `AppCard`, `AppEmptyState`, `AppErrorState`: surfaces and the standard empty/error blocks.
- `AppSpinner`, `LoadingBar`, `skeleton/`: loading states.
- `components/movie/Starratinginput.vue`: rating slider (0.5 steps, Arrow/Home/End) and read-only half-star display (`readonly`).
- Toasts: one global `ToastContainer` is mounted in `App.vue`; push toasts through the ui store, don't mount another.
- z-index scale (`main.css`): header 100 < legacy modals 9999-99999 < AppModal (`--z-modal-base` 100000, +10 per open modal) < popovers/datepicker menus (`--z-popover`) < toasts (`--z-toast`). Don't add new hard-coded z-index values above 1000.
- Icons: `primeicons` font (`pi pi-*`) is still loaded for legacy views; PrimeVue itself is removed. Prefer `lucide-vue-next` for new code.

## Design system (client)
Tokens live in `client/src/assets/styles/main.css` (`@theme static`); use them as Tailwind utilities (`bg-surface`, `text-text-muted`, `border-border`).
| Token | Value | Use |
|---|---|---|
| `surface` / `surface-2` / `surface-3` | #0b0b0c / #151517 / #202023 | page root / cards / fills |
| `border` / `border-strong` | white 8% / 14% | hairlines / emphasized edges |
| `text-primary` / `text-secondary` / `text-muted` | #f5f5f7 / #a1a1aa / #8a8a8e | text |
| `brand` / `brand-hover` / `on-brand` | #d6341f / #c42d19 / #fff (white on brand 4.81:1, on hover 5.62:1; hover is darker because a lighter red fails 4.5:1) | actions |
| `success` / `warning` / `danger` / `info` | #34c759 / #fbbf24 / #ff5a52 / #38abeb | status (`danger` = destructive) |
| `rating`, `focus`, `accent` | #fbbf24, #f5f5f7, #ffd93b | stars, focus ring, highlight |
| `font-sans` | Noto Sans Thai first | only font stack; loaded once from `index.html` |
| `text-caption..display`, `rounded-card`, `rounded-btn` | see file | type scale (not migrated yet), radii |
- Light mode later = one `[data-theme=light]` override block of these variables. Avoid `dark:` variants.
- Hard-coded color mapping: #080808-#141414 page roots -> `surface`; #111-#1a1a1a cards -> `surface-2`; #1f1f1f/#2a2a2a fills -> `surface-3`; white alpha 5-8% -> `border`, 12-15% -> `border-strong`; #e1251b/#f5432c -> `brand` (actions) or `danger` (destructive); #f5c518/#fbbf24 -> `rating`.
- Leave alone: share templates, poster scrims, `PopcornRatingIcon` gradients.
- Convert on edit: when you touch a file, convert its hard-coded colors. Never run a codemod over the repo.
- `--c-*` variables are legacy aliases to the tokens; remove them file by file, don't add new ones.
- `designTokens.spec.ts` is a ratchet: it fails if `.vue` hex/rgb counts or `--c-` definitions grow. Lower its baseline when counts drop.
- Breakpoints for new CSS: 480/640/768/1024/1280, mobile-first (`min-width`).
- Input focus ring is global (`motion.css`); add `data-focus-custom` to opt out when a control draws its own.

## Hard rules (invariants)
1. Ratings run from 0.5 to 5.0 in 0.5 steps, enforced by DB constraint `chk_reviews_rating_step`. Reviews have a unique constraint per user and media (migration `review_unique_constraint_v1`).
2. Library entries must not duplicate; a partial unique index guards this (`idx_library_items_user_media_list`).
3. Reads of another user's data go through `privacy_policy`. Activity visibility is `default | public | followers | private`.
4. Admin routes use `RequireCurrentAdmin`, which checks the role in the DB on every request. Don't rely on JWT claims alone.
5. Soft-deleted rows stay in the DB. Filter `deleted_at IS NULL` in raw SQL, as `RequireCurrentAdmin` does for users.
6. Auth uses `access_token` and `refresh_token` HTTP-only cookies with rotation. Handoff tokens and OAuth session tokens go in the URL fragment, never the query string, and are never logged.
7. UI text goes in `i18n/locales/{en,th}`. Keep both languages in sync (`verify-locales` checks this).
8. Account takeover guards. `users.verified_email_at` is set only by `identity_policy.ProveEmail`, which wipes credentials planted on an unverified account unless the person proves the current password. In the verify-email flow only an explicit `skip_password` ("I don't know it") wipes; a wrong password returns 400 `WRONG_PASSWORD`, changes nothing, keeps the token, and counts against `MaxVerifyAttempts` per token (then 429, resend needed). Sessions are created only in `auth_module.issueAndStoreTokensTx` (verified email required) and carry `ver` = `users.token_version`; `RequireAuth`, refresh and the socket handshake (re-checked after Join) reject a mismatch. Claims, resets, password changes and logout-all bump `token_version`; a password change reissues the caller's session. Verification, reset and link-code tokens are bound to the email they were mailed to. Sensitive writes call `identity_policy.EnsureSession` inside their transaction. Never set `verified_email_at` or `password` with a bare UPDATE outside these paths.

## Recipes
**New endpoint:** dto -> service (+test) -> handler -> module `router.go` -> new module only: register in `router/router.go` -> `client/src/types` + `client/src/api/endpoints` -> view.
**New table:** add the model to the `autoMigrate` list in `database/db.go`.
**New SQL migration:** add a `runMigrationWithHistory` call with a new unique version string in `runSQLMigrations` (`database/db.go`). Use `runMigrationWithHistoryTx` when it needs timeouts. Never edit a version that has already been applied; `migration_history` will skip it.
**New achievement:** edit `database/seeder/achievement.json`; it is seeded on boot.
**New UI text:** add the key to both `en` and `th`.

## Naming (intentionally not renamed)
Changing any of these affects stored data, live sessions, or other systems:
- `removy-official.pages.dev` URLs (needs a new domain) and `removy.official@gmail.com` sender (needs a verified Brevo sender).
- `removy-private` bucket: stored files must be moved first.
- Repo/folder `removy-official`, Go module `github.com/arinsuda/movie-hub`, DB `moviehub`, containers `moviereview_*`.
- Legacy fallback only: `LEGACY_LOCALE_STORAGE_KEY` (`removy-locale`) is still read. PLANZI still accepts `removy` as an input alias for `popcorn`.

## Gotchas
- `OMDB_API_KEY` (optional): enables IMDb and Rotten Tomatoes ratings on movie/TV detail. Never log it; tests must stub the OMDb HTTP client.
- `DEV_MAIL_LOG=true` prints emails (verify link, reset link, link code, email-change OTP) to the server stdout instead of sending. It activates only if every `CORS_ALLOWED_ORIGIN` entry, `APP_BASE_URL` and `FRONTEND_BASE_URL` are loopback; otherwise boot logs a WARN and it stays off. Keep it the only place codes/links are printed.
- Env precedence: real environment > `server/.env` > repo-root `.env` (`config.LoadDotEnv`). `docker compose` needs `--env-file server/.env`.
- CORS has one source: `CORS_ALLOWED_ORIGIN` (comma-separated exact origins, no wildcard), used by Fiber, Socket.io, and the `/socket.io` handler in `main.go`. Production must set its frontend origin.
- `node_modules` is platform-specific. After moving between OSes, run `npm ci` in `client/`.
- `client/dist` is committed, so `npm run build-only` dirties the tree. Restore it with `git restore client/dist && git clean -fd client/dist` unless you mean to update it.
- `npm run lint` rewrites files, so run it only when you accept those edits.
- Huge files: UserBmol.vue (~2.4k), BmolSection.vue (~1.8k), TVDetailView / EditProfile / MovieDetailView (~1.6k), HomeView (~1.4k), MainLayout (~1.3k), MovieReviews (~1.3k). Use grep + ranged reads, not full reads.
- `server/tmp/` is air build output (`.air.toml`). `components/HelloWorld.vue`, `TheWelcome.vue`, and `WelcomeItem.vue` look like Vite scaffold leftovers; check usage before relying on or removing them.

## Docs (trust level)
- Code > this file > `README.md` (may lag code) > `client/README.md`.
- `../LINKING_SPEC.md` is the shared linking contract with PLANZI (`../cooked-app`). 
- Commit style: `feat: ...` / `fix: ...` / `refactor: ...` / `docs: ...`.
- Claude Code subagents and dispatch rules are described in `CLAUDE.md`.
