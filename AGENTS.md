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
cd client && npm run shots           # responsive screenshots + horizontal-overflow check (see below); needs a running stack, not in CI
```
Verify before finishing: `go vet ./... && go test ./...`, `npm run type-check`, `npm test`, `npm run build-only`, and `npm run verify-locales` when locale files change. CI (`.github/workflows/ci.yml`) runs these plus `oxlint` and `eslint src` (no --fix) on every push/PR.

**Responsive shots (local only).** `client/scripts/responsive-shots.mjs` visits guest and signed-in routes at 390/768/1024/1180/1440 px, writes PNGs + `report.json` to `client/.shots/latest/` (gitignored) and exits 1 if any page has document-level horizontal overflow (elements inside overflow-x auto/scroll/hidden containers are ignored). Start the stack first, then: `RESPONSIVE_BASE_URL=http://localhost:3000 RESPONSIVE_USER=<user> RESPONSIVE_PASSWORD=<pw> npm run shots -- --montage`. Credentials come only from env; without them authed routes are skipped. It drives system Chrome (`channel: chrome`); set `RESPONSIVE_CHROME_PATH` to use another executable. `--montage` adds a contact sheet per viewport. Routes and viewports are constants at the top of the script.

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
  composables/ useSocket, useFollow, useReviewForm, useShareImage, ...
  i18n/locales/{en,th}/*.json  UI text, one file per area
```

## Request flow
`router/router.go -> <module>/router.go -> middleware -> handler -> service -> repository (GORM)`
- Dependencies are passed through each module's `RegisterRoutes`/`New`; wiring order lives in `router.go`.
- API prefix is `/api`. Most modules mount on the `protected` group (including `follow` and `achievements` protected routes, while their public GET routes mount in `router.RegisterPublicRoutes`). `auth` and `link` receive `api` and apply auth per route.
- **There is no success envelope.** Handlers return named keys such as `{"review": ...}` or `{"reviews": [...]}`. Errors are `{"error": "..."}`, or `{"code","error"}` through `shared.WriteAPIError`. Follow the style of the module you edit.
- **Public (guest) routes:** `GET /movies/popular`, `/movies/now-playing`, `/movies/top-rated`, `/movies/upcoming`, `/movies/search`, `/movies/genres`, `/movies/:id<int>`, `/movies/:id<int>/similar`, `/movies/:id<int>/watch-providers`, `GET /tv/popular`, `/tv/now-airing`, `/tv/search`, `/tv/genres`, `/tv/:id<int>`, `/tv/:id<int>/similar`, `/tv/:id<int>/watch-providers`, `GET /actors/:id<int>`, `/actors/:id<int>/movies`, `/actors/:id<int>/tv`, `GET /reviews/media/:mediaType/:mediaId` (+ `/rating`), `GET /stats/:media_type/:media_id`, `GET /users/:userId<int>` (+ `/stats`, `/likes`, `/achievements`, `/followers`, `/following`), `GET /reviews/user/:userId<int>`, `GET /library/user/:userId<int>`, `GET /bmol/user/:userId<int>`, and note `GET /auth/session` (`{"has_session": bool}` for guest cold boot) is NOT one of them: it lives in `auth_module` with a plain per-IP limiter (`sessionProbeRateMax` per `sessionProbeRateWindow`). They are registered by `router.RegisterPublicRoutes` (per route: `OptionalAuth`, then the guest limiter `PublicReadLimiter`, `PublicReadRateMax` per `PublicReadRateWindow` per IP; signed-in users are not counted; `/movies/search` and `/tv/search` also pass `PublicSearchLimiter`, `PublicSearchRateMax` per `PublicSearchRateWindow`). `OptionalAuth`: no token = guest (handlers see no claims, viewer id 0), a present but invalid or stale token = 401, DB error = 503. Everything else — including `/movies/upcoming/:year`, `/movies/recommended`, `/tv/top-rated`, `/actors/search`, `POST /stats/.../view`, `/api/achievements`, `/users/:userId/follow-status`, `/users/:userId/follow-stats`, `/users/:userId/follow-requests`, `/api/lists`, `/api/feed` — needs a session. Public ids use `:id<int>` so fixed paths like `/movies/recommended` and `/actors/search` fall through to the protected router.
- **Fiber v3 group gotcha:** `api.Group("/", handlers...)` registers a USE on the whole prefix, so it runs in front of every route registered after it. `protected := api.Group("/", mw.RequireAuth)` therefore gates every `api` route registered after it (`/api/achievements` is registered on `protected` itself). Register public routes on `api` BEFORE that line, with per-route middleware (`middleware.GetRoute`); never `api.Group("/", OptionalAuth, ...)`. Modules with public and protected routes expose `New(...)`, `RegisterPublicRoutes` and `RegisterRoutes` sharing one handler/service.
- Socket.io bypasses Fiber: `main.go` hands it directly to `notifHub.Handler()`, with its own CORS headers.

## Feature -> files map
| Feature | Backend (`server/internal/`) | Frontend |
|---|---|---|
| Auth, email verification, Google OAuth, sessions | auth_module, user_module (password reset, email change mailers) | views/auth/, components/auth/ (AuthShell, AuthField, AuthAlert, GoogleSignInButton, PasswordStrength), stores/auth.ts, utils/urlSecrets.ts (fragment secrets); server `shared/fragment` builds the mailed links |
| Account linking PLANZI<->POPCORN (HMAC server-to-server, handoff token in URL fragment; contract in `../LINKING_SPEC.md`; disabled when `LINK_SHARED_SECRET` is empty; login falls back to remote `verify-password`) | link_module | views/auth/HandoffView.vue (back-to-PLANZI link from public `GET /api/link/remote` -> `{"web_url"}`, 404 when linking is off), api/endpoints/link.ts |
| Movie, TV, actor, search, upcoming, where-to-watch (`GET /movies/:id/watch-providers`, `/tv/:id/watch-providers`; TMDB/JustWatch, 6h cache) | movie_module (`providers.go`), tmdb_module (`providers.go`) | views/movie/, views/tv/, views/actor/, views/SearchResultView.vue, components/detail/DetailWatchProviders.vue, composables/useWatchProviders.ts |
| Ratings display (POPCORN primary; TMDB, IMDb, Rotten Tomatoes secondary). Movie/TV lists return `ratings.{popcorn,tmdb}`; detail adds `imdb`, `rotten_tomatoes` from OMDb (`OMDB_API_KEY`, empty = off). `ExternalRating` table caches 7d (1d for not_found), 3s upstream timeout. A stale row is served at once and refreshed in the background; a cold miss waits at most 700ms, then returns unavailable while the fetch fills the cache. `imdb_id` must match `^tt\d{1,10}$` or it is not_found. The 900 req/day budget is in-memory and per instance (resets on restart; N instances allow N x 900), taken inside the singleflight call. Library, likes, BMOL and feed items carry `media.ratings` via `shared.AttachRatings` | movie_module (dto/service), external_rating_module, shared/media_ratings.go | components/movie/MediaRating.vue (`card`/`detail`), PopcornScoreBadge.vue, popcornTier.ts (tiers: none, stale <2.5, fresh 2.5-3.9, hot >=4.0), utils/mediaRatings.ts |
| Movie / TV detail pages | movie_module (GET /movies/:id, /tv/:id), tmdb_module | components/detail/ (DetailPage + detailConfig adapters; reviews/ split), composables/useMediaDetail.ts, useMediaActivity.ts; views/movie/MovieDetailView.vue and views/tv/TVDetailView.vue are thin wrappers |
| Reviews, comments, ratings | review_module, like_module, media_stats_module | components/movie/MovieReviews.vue, composables/useReviewForm.ts |
| Library lists | library_module | views/library/ |
| User-made lists (named, ordered, per-list visibility; caps and codes in `list_module/model.go`, `shared/apierr.go`) | list_module, `privacy_policy.CanViewList` | |
| Watch logs | watch_log_module | api/endpoints/watchLog.ts, types/watchLog.ts |
| Best of Life (BMOL, "Hall of Fame"). Ranks may tie (several titles per rank, no renumbering). Items carry an optional `note` (<=280 runes) returned only by the privacy-gated `GET /bmol/user/:id`. Re-adding a soft-deleted title restores the row; `PUT /bmol/:id` takes `{rank?, note?}` (one required); validation is manual (rank 1-999, media_id > 0) | bmol_module | components/bmol/ (BmolBoard = whole UI; `useBmol` owns the `["bmol", userId]` query and optimistic mutations; `bmolRanks.ts` pure grouping/move logic; `useBmolReorder` drag + keyboard), `UserBmol.vue` and `BmolSection.vue` are thin wrappers, `ProfileBmolHighlight` shares the query key, share cards in components/share/BolShareSheet.vue + cards/Bol*Card.vue |
| Achievements | achievements_module, database/seeder/achievement.json | views/achievement/, stores/achievement.ts |
| Follow, feed | follow_module, feed_module | views/FeedView.vue, components/feed/, components/social/ |
| Profiles, stats, privacy | user_module, user_stats_module, analytics_module, privacy_policy | views/user/ (thin shell), components/profile/ (ProfileHero, ProfileTastePanel, FollowListModal, tasteProfile.ts), composables/useUserProfile.ts |
| Settings (profile form + avatar crop, preferences, privacy, email/password, Google/PLANZI connections) | user_module, auth_module | views/settings/SettingsView.vue (route `/settings`, anchors #profile #preferences #privacy #security #connections), components/settings/ |
| Notifications, realtime. The server stores no display copy: new rows have `message=""` plus `params` (jsonb, `NotificationParams`: media_id/media_type/media_title snapshot, or achievement action_type/target_count/exp/name; never email, review body or review owner); the actor is resolved live from `actor_id`. The client builds text in `utils/notificationText.ts` (`formatNotification`: `notifications.types.<type>` key, then legacy `message`, then title). REST and socket share `NotificationResponse`. Legacy rows (params NULL) keep their old message. Deploy: ship the client before the server | notification_module (`socket.go` = hub, `params.go`), migration `20261110_notification_params` | views/NotificationsView.vue, stores/notification.ts, utils/notificationText.ts, composables/useSocket.ts |
| Share images | `shareImageProxyHandler` in router.go | components/share/, composables/useShareImage.ts |
| Admin | admin_module | views/admin/, layouts/AdminLayout.vue, stores/admin.ts |
| Admin account erasure (permanent, after soft delete; `GET /admin/users?status=deleted` lists candidates; 204, errors `SELF_ERASURE` 403, `USER_NOT_FOUND` 404, `USER_NOT_REMOVED` 409, `CONFIRMATION_MISMATCH` 400, `FINAL_ADMIN` 409; audit action `USER_ERASED`; avatars removed from MinIO after commit) | account_erasure (`Erase(tx, userID)`), admin_module (`EraseUser`), storage `DeletePrefix` | components/admin/AdminEraseUserModal.vue, views/admin/AdminUsersView.vue |
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
- `components/movie/StarRatingInput.vue`: rating slider (0.5 steps, Arrow/Home/End) and read-only half-star display (`readonly`).
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
- `--c-*` legacy variables are gone (baseline 0); don't reintroduce them.
- `designTokens.spec.ts` is a ratchet: it fails if `.vue` hex/rgb counts or `--c-` definitions grow. Lower its baseline when counts drop.
- Breakpoints: `xs 480 / sm 640 / md 768 / lg 1024 / xl 1280`, mobile-first (`min-width`). Defined once as `--breakpoint-*` in `main.css` (Tailwind `xs:`..`xl:` utilities) and `src/constants/breakpoints.ts` (JS: `minWidthQuery`, `isAtLeast`); the two are kept in sync by `breakpoints.spec.ts`. Never write other px values in `@media` or `min-[Npx]:`. Where inverting to `min-width` is risky, `max-width` is allowed only as token minus 0.02px (`767.98px`).
- Viewport height: never plain `100vh`/`NNvh`. Use `100dvh` (write a `100vh` line right before it as the old-browser fallback) or `svh` for stable proportions. Enforced by `scanLayout` in `noHardcode.spec.ts` over the whole `src` tree.
- Safe areas: `index.html` uses `viewport-fit=cover`, so fixed/sticky edges (bottom bars, sheets, drawers, toasts, header sides) add `env(safe-area-inset-*)`, e.g. `padding-bottom: calc(0.5rem + env(safe-area-inset-bottom))` or `max(1rem, env(...))`. `body` already insets left/right.
- Input focus ring is global (`motion.css`); add `data-focus-custom` to opt out when a control draws its own.

## Hard rules (invariants)
1. Ratings run from 0.5 to 5.0 in 0.5 steps, enforced by DB constraint `chk_reviews_rating_step`. Reviews have a unique constraint per user and media (migration `review_unique_constraint_v1`).
2. Library entries must not duplicate; a partial unique index guards this (`idx_library_items_user_media_list`).
3. Reads of another user's data go through `privacy_policy`. Activity visibility is `default | public | followers | private`.
4. Admin routes use `RequireCurrentAdmin`, which checks the role in the DB on every request. Don't rely on JWT claims alone.
5. Soft-deleted rows stay in the DB. Filter `deleted_at IS NULL` in raw SQL, as `RequireCurrentAdmin` does for users. The one exception is admin account erasure (`POST /admin/users/:userId/erasure`, `account_erasure.Erase`): a second step after the admin soft delete (409 `USER_NOT_REMOVED` otherwise) that hard-deletes the user and, via FK CASCADE, their reviews, comments, likes, lists, follows and notifications. Other people's counters are decremented, review children (no FK) and activity about the user are deleted, audit logs stay with `admin_id` NULL and `target_username`/`target_email` stripped. It needs `confirm_username` to match exactly and takes `LockAdminGuard` first.
   Hard `DELETE FROM users` must cascade cleanly: every column referencing users has an FK with `ON DELETE CASCADE` (`admin_audit_logs.admin_id` and `activity_events.target_user_id` are `SET NULL`, so audit rows outlive the admin). Add new user-referencing columns to `migrate.UserFKs` (`database/migrate/user_foreign_keys.go`); `ReplaceUserFK` drops any existing FK on the column, clears orphans and re-adds it `NOT VALID`, and `ValidateUserFK` validates it in a separate migration (`20261111_user_fk_cascade_*`). Don't put relation tags on `user_lists` or `notifications.actor_id`: AutoMigrate runs before the cleanup and would add a validating FK. GORM tag is `constraint:` (a `contraint:` typo is silently ignored).
6. Auth uses `access_token` and `refresh_token` HTTP-only cookies with rotation. Handoff, OAuth session, email-verification and password-reset tokens go in the URL fragment (read via `utils/urlSecrets.ts`, stripped at once), never the query string, never logged.
7. UI text goes in `i18n/locales/{en,th}`. Keep both languages in sync (`verify-locales` checks this).
8. Account takeover guards. `users.verified_email_at` is set only by `identity_policy.ProveEmail`, which wipes credentials planted on an unverified account unless the person proves the current password. In the verify-email flow only an explicit `skip_password` ("I don't know it") wipes; a wrong password returns 400 `WRONG_PASSWORD`, changes nothing, keeps the token, and counts against `MaxVerifyAttempts` per token (then 429, resend needed). Sessions are created only in `auth_module.issueAndStoreTokensTx` (verified email required) and carry `ver` = `users.token_version`; `RequireAuth`, refresh and the socket handshake (re-checked after Join) reject a mismatch. Claims, resets, password changes and logout-all bump `token_version`; a password change reissues the caller's session. Verification, reset and link-code tokens are bound to the email they were mailed to. Email verification tokens are stored as `HashToken` (SHA-256 hex); only the mailed link carries the raw value, in the fragment (`/auth/verify-email#token=...`). `DELETE /users/:id` is self-only (403 otherwise) and is a deactivation (`is_active=false`, sessions revoked), not the admin soft delete: reviews stay public and the email stays reserved. It refuses the last active admin (409 `FINAL_ADMIN`); admin soft delete (`DELETE /admin/users/:id`) also hard-deletes the user's `user_identities`, `account_links`, `link_pending`, `link_handoff_tokens`, `password_reset_tokens`, `email_verifications` and `email_change_requests` in the same transaction so the Google/PLANZI account can sign up again (migration `20261111_release_removed_user_links` backfilled old rows; Google login also drops a stale identity of a soft-deleted user), while self-delete releases nothing; every admin-removing transaction first takes `identity_policy.LockAdminGuard` (advisory lock) before any row lock; admin delete/deactivate also bump `token_version` and revoke sessions. `POST /auth/resend-verification` answers identically for unknown, unverified and verified emails. Sensitive writes call `identity_policy.EnsureSession` inside their transaction. Never set `verified_email_at` or `password` with a bare UPDATE outside these paths.

## CSRF rule
Prod runs `COOKIE_SAMESITE=None` and Fiber's JSON binder ignores Content-Type, so every unsafe request (POST/PUT/PATCH/DELETE) under `/api` needs `X-Requested-With: XMLHttpRequest`, else 403 `XHR_HEADER_REQUIRED` (`middleware.RequireXHR`, the first handler on `api` in `router.Register`; keep it first). The SPA axios instance sends it (`CSRF_HEADERS` in `client/src/api/index.ts`); any new non-axios unsafe call must add it. CORS must keep `X-Requested-With` in `corsAllowHeaders` (`main.go`, shared by Fiber and `/socket.io`). Exemptions: `middleware.CSRFExemptPrefixes` (path prefixes, each with a why-comment); only for routes not authenticated by a browser cookie (today: `/api/internal/link/`, HMAC). Never exempt login, register, refresh or logout. Integration test env (`testenv_test.go`) installs the same middleware and its `do` helper sends the header.

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

## Testing rules (client)
- Tests are hermetic: no real network, no wall-clock sleeps. `src/__tests__/setup/noNetwork.ts` (vite `test.setupFiles`) blocks axios (global default adapter), `fetch` and `XMLHttpRequest`; an unmocked request throws `[no-network] ... METHOD URL` and the test fails in `afterEach` even if app code swallowed the error. happy-dom script/CSS/iframe loading is disabled.
- Mock at the right level: `vi.mock("@/api/endpoints/<x>")` (or `@/api/api`), set `api.defaults.adapter` for interceptor tests, or `vi.spyOn(globalThis, "fetch")`. For a failure path, mock a deliberate rejection.
- Wait deterministically: `flushPromises`, `nextTick`, `vi.waitFor(() => expect(...))`, or `vi.useFakeTimers` + `advanceTimersByTimeAsync` for debounces. Never `setTimeout` sleeps; don't raise `testTimeout`.

## Gotchas
- **Sentry (optional, off when the DSN is empty):** server `SENTRY_DSN`, `SENTRY_ENVIRONMENT` (default production), `SENTRY_TRACES_SAMPLE_RATE` (default 0.05); client (build-time) `VITE_SENTRY_DSN`, `VITE_SENTRY_ENVIRONMENT`, `VITE_SENTRY_RELEASE`. Server: `server/monitoring` (`Middleware` first on the Fiber app captures panics and 5xx only; flushed on shutdown, `FlushTimeout`). Client: `client/src/monitoring/` (`initSentry` from `main.ts`, constants in `config.ts`, no Session Replay). Scrubbed: URL fragments and token-like query values, cookies, headers, request bodies, emails in messages/breadcrumbs, user reduced to numeric id; client drops axios errors below 500 (4xx) and keeps network failures. Never log the DSN.
- `OMDB_API_KEY` (optional): enables IMDb and Rotten Tomatoes ratings on movie/TV detail. Never log it; tests must stub the OMDb HTTP client.
- `DEV_MAIL_LOG=true` prints emails (verify link, reset link, link code, email-change OTP) to the server stdout instead of sending. It activates only if every `CORS_ALLOWED_ORIGIN` entry, `APP_BASE_URL` and `FRONTEND_BASE_URL` are loopback; otherwise boot logs a WARN and it stays off. Keep it the only place codes/links are printed.
- Env precedence: real environment > `server/.env` > repo-root `.env` (`config.LoadDotEnv`). `docker compose` needs `--env-file server/.env`.
- CORS has one source: `CORS_ALLOWED_ORIGIN` (comma-separated exact origins, no wildcard), used by Fiber, Socket.io, and the `/socket.io` handler in `main.go`. Production must set its frontend origin.
- `node_modules` is platform-specific. After moving between OSes, run `npm ci` in `client/`.
- `client/dist` is committed, so `npm run build-only` dirties the tree. Restore it with `git restore client/dist && git clean -fd client/dist` unless you mean to update it.
- Specs: a click in the same millisecond as the render can be silently dropped by Vue's event invoker (both `.click()` and test-utils `trigger`, whenever an earlier listener such as AppButton's `@click.capture` stamped the event). Use `clickAfterRender` from `__tests__/helpers/waitForEnabled.ts` for buttons just rendered by v-if/modals.
- `npm run lint` rewrites files, so run it only when you accept those edits.
- `tmdb_module` caches movie and TV detail in process (`detailCacheTTL`, `detailCacheMax`; keyed by type + id + the configured TMDB language). Tests that fake TMDB should use ids no other test caches.
- `users.onboarded_at` is set once when favorite genres are saved or skipped (`PATCH /users/:id/genres`); existing users have NULL until they do. Self-delete (`DELETE /users/:id`) needs `{password}` or, for accounts without a password, `{confirm_username}`.
- **Self-service erasure (PDPA):** `POST /users/me/erasure-request` (`{password}` or `{confirm_username}`, same re-auth as `DELETE /users/:id`, one shared failed-only rate limiter) needs `Content-Type: application/json` (else 415 `UNSUPPORTED_MEDIA_TYPE`) and `X-Requested-With: XMLHttpRequest` (else 403 `XHR_HEADER_REQUIRED`; the axios instance sends it) because prod cookies are SameSite=None. It deactivates the account like self-delete, sets `users.erasure_due_at = now + config.ErasureGracePeriod` (30 days, constants in `server/config/erasure.go`), writes a `USER_ERASURE_REQUESTED` audit row (admin_id NULL, meta = due date) and mails the owner the date plus `config.ErasureSupportEmail` to cancel; 202 `{erasure_due_at}`; last active admin gets 409 `FINAL_ADMIN`. Email and Google identity stay reserved during the grace period. No in-app cancel (sessions are revoked): support asks an admin to reactivate the user (`PATCH /admin/users/:id/status` active), which clears `erasure_due_at`. The purge worker (`account_erasure.Purger`, started from `router.Register` with the router's MinIO client, tick every `ErasurePurgeInterval`, `ErasurePurgeBatch` per tick, panics recovered) has no global lock: each account is erased in its own tx (`SET LOCAL` lock/statement timeouts, `LockAdminGuard`, `FOR UPDATE` and due re-check, skips the last active admin), so several instances are safe. It targets inactive users with `erasure_due_at <= now` and admin-soft-deleted users older than `AdminDeletePurgeAfter` (30 days). Reviews are hard-deleted. Avatars are deleted after commit; each tick retries audit rows still `storage_cleanup: pending`. Audit action `USER_ERASED_SCHEDULED` (admin_id NULL, meta `trigger`, `target_role`, `storage_cleanup`; no PII). Client: `components/settings/SettingsEraseAccount.vue`; admin users list shows an "Erasure pending" badge; the audit log view labels admin-less rows (by the system / by the account owner). Migration `20261112_users_erasure_due_at`.
- **Guest list cap:** `GET /library/user/:userId` and `GET /users/:userId/likes` return at most `library_module.GuestLibraryLimit` / `like_module.GuestLikesLimit` (100) items to a guest, ordered `created_at DESC, id DESC`, plus `has_more` (true when cut; always false for signed-in viewers, who get the full list). The client shows `components/profile/tabs/GuestCapNote.vue`.
- **Email change errors:** `POST /users/:id/email` returns `{code, error}` with `EMAIL_INVALID`, `EMAIL_SAME` (400) or `EMAIL_IN_USE` (409); the client maps codes in `components/settings/emailChange.ts`, never the `error` text.
- Activity-type defaults live only in `privacy_policy/activity_defaults.go` (feed SQL is generated from it).
- Huge files: UserBmol.vue (~2.4k), BmolSection.vue (~1.8k), TVDetailView / EditProfile / MovieDetailView (~1.6k), HomeView (~1.4k), MainLayout (~1.3k), MovieReviews (~1.3k). Use grep + ranged reads, not full reads.
- `server/tmp/` is air build output (`.air.toml`).

## Backups
- Daily encrypted logical backup: [`.github/workflows/db-backup.yml`](.github/workflows/db-backup.yml) runs daily off-peak + `workflow_dispatch`, streams `pg_dump` (PG 16) into `age`, and keeps encrypted `.dump.age` artifacts for `BACKUP_RETENTION_DAYS` (30 days).
- Required GitHub Secrets: `BACKUP_DATABASE_URL` (read-only Neon role on the **direct**, non-pooler host with `sslmode=require`) and `BACKUP_AGE_RECIPIENT` (`age1...` public key; the private key stays offline).
- Restore script & runbook: [`scripts/db-restore.sh`](scripts/db-restore.sh) (guards `*.neon.tech` unless `--i-know-this-is-prod`) and [`docs/backup-restore.md`](docs/backup-restore.md) (role SQL, <6h PITR vs >6h backup table, manual restore drill, 30-day PDPA retention note).

## Docs (trust level)
- Code > this file > `README.md` (may lag code) > `client/README.md`.
- `../LINKING_SPEC.md` is the shared linking contract with PLANZI (`../cooked-app`). 
- Commit style: `feat: ...` / `fix: ...` / `refactor: ...` / `docs: ...`.
- Claude Code subagents and dispatch rules are described in `CLAUDE.md`.
