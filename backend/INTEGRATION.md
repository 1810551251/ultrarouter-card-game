# Integrating The Game Into Your Own Host

The exported Go files are game-specific packages from the existing new-api integration. The complete host is **not included**. The public upstream attribution/reference is recorded in `SOURCE.lock.json`; the reference is not a promise that every custom host helper exists in an arbitrary upstream release.

## Files And Initialization

Copy the game `model/gacha*.go`, `model/gacha_catalog.json` and `controller/gacha*.go` into your own compatible host packages, preserving their current new-api import identifiers. Use the examples in `integration/` to add only game routes and tables. The examples are not an automatic installer and never contact a maintainer server.

The game uses ten tables: `GachaRun`, `GachaCard`, `GachaDaily`, `GachaPlayer`, `GachaAction`, `GachaArenaGate`, `GachaArenaSeat`, `GachaArenaDuel`, `GachaArenaEntry`, and `GachaArenaRoom`. Initialize power and arena state after migrations succeed. The host must also own the user, option/configuration and log stores consumed by the game.

Serve the game files at `/gacha/` and the host APIs at the same origin. A navigation link may use `target="_blank"` and `rel="noopener noreferrer"`; no deployment domain or IP is supplied here.

Expose only the game HTML, CSS, `js/` and artwork as static assets. Do not expose the host's environment/data directories, backend source, development scripts or test files through the game mount.

## External Host Contracts

These dependencies are intentionally referenced rather than bundled as a complete account platform:

| Host Responsibility | Existing Integration Points |
| --- | --- |
| Validated browser identity | `middleware.GetSessionAuthIdentity(c)` must return a server-authenticated session identity, including `UserID`; game controllers reject non-session API credentials |
| Browser login/refresh | `/api/user/auth/refresh`, `/sign-in`; refresh returns the authenticated user and a short-lived access token kept only in browser memory |
| Database and account rows | `model.DB`, host `User`/configuration/log models, `common.UserStatusEnabled`, `common.RoleAdminUser` |
| JSON and API responses | `common.RawMessage`, `common.Marshal`, `common.Unmarshal`, `common.UnmarshalJsonStr`, `common.ApiSuccess` |
| Wallet unit and bounds | 1 E equals 500,000 integer quota; provide compatible `common.MaxWalletQuota`, `common.QuotaRound` and `common.WalletQuotaFromDecimalStrict` behavior |
| Wallet/cache consistency | host-local `cacheIncrUserQuota`, `syncCreditUserQuotaCache`; update caches only after a committed wallet mutation |
| Dialect-safe transactions | host-local `lockForUpdate(tx)`; use row locks on MySQL/PostgreSQL and compatible SQLite serialization |
| Accounting logs | host `RecordLog` and `LogTypeSystem`; do not log credentials |
| Rate limits | authenticated-user-scoped `UserBurstRateLimit`, a game action budget of 60 per 60 seconds, run-start budget of 30 per 60 seconds, and no cached private responses |
| Multiplier-card effects | `model.ActiveCardDiscount(userID, now)` returns the active server-owned discount; integrate it into your own host's multiplier policy if that feature is used |

The model also consumes non-secret host runtime flags such as batch-update/Redis availability. Adapt these to your own services rather than shipping their production environment. Atomic E exchange must not silently run against a delayed wallet mode that cannot enforce the transaction contract.

This list describes the integration boundary; it is not an implementation of the host's authentication, billing, cache or account management. If you change hosts, implement and test those adapters first. Do not bypass them with test logins, fixed IDs, static tokens or browser wallets.

## Game Requests

The route example lists the game GET/POST endpoints. The browser submits only action intent and server-issued run/card identifiers. Controllers reject unexpected fields, uploaded battle results and local saves. PvP is entered voluntarily through the queue or custom-room lobby; escrow and settlement are server-side.

The refresh interface is a host authentication endpoint, not one of the exported game handlers. Match its response contract in `js/store.js` when adapting your own host. Keep HttpOnly session cookies, trusted-origin/CSRF controls, memory-only access tokens and logout invalidation.

## Validation In A Host

Run these commands from your **own compatible host checkout**, inside the shared capped test container:

```sh
GOMAXPROCS=1 go test -p 1 -parallel 1 ./model ./controller -run TestGacha -count=1
```

`model/gacha_game_test.go` accepts `TEST_MYSQL_DSN` and `TEST_POSTGRES_DSN` for your own disposable databases. Set `GACHA_REQUIRE_DB_MATRIX=1` when validating all three engines so missing configured database targets fail instead of silently skipping. Never point these variables at production data.

The root game package has no host `go.mod`, complete database bootstrap or host image. Running `go test ./...` directly in this directory is therefore not a supported standalone-server test. Browser mocks and source audits do not substitute for host integration and database tests.
