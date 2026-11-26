# Game-Only Export Validation

## Ranked Arena Feature — 2026-11-26

- Added independent ranked challenges, saved defenses, rank swaps, best rank, 05:00 daily attempts and 21:00 material settlement/claims. Existing E matching remains separate.
- `npm test`: 119 browser-interface assertions and 56 tests/subtests passed. Ranked checks include registration, target/expected-rank intent, rank result rendering, private replay routing, non-economic playback, material claim intent, returned inventory updates, claimed-state disabling and attack/defense filtering.
- Python preview/release regression: 19 tests passed. Source allowlist and credential-pattern checks passed across 152 game files; the changed fixture files retain exactly the same reviewed synthetic credential matches. No private deployment-marker list was supplied for this update, so that additional scan was not claimed.
- A local ignored mock was exercised through the browser scripts against HTTP: registration at rank 101, challenge promotion, advancing to 21:00, creation of material mail, claiming into inventory with unchanged E, plus existing queue/room/replay flows. Mock results verify UI behavior only, not Go rules or database transactions.
- Headless Edge checked nine views at five widths (320, 390, 768, 1024, 1440): 45 horizontal-fit checks passed, including ranked onboarding, opponent cards, ranking and material rewards. Images were blocked/removed; no screenshots or image inspection were performed.
- Go files and integration examples passed `gofmt` syntax parsing. The new database cases are wired into the existing SQLite/MySQL/PostgreSQL matrix and cover fresh/additive/idempotent migration, repeated registration, rank swaps, cooldown/stale quotes, frozen defense, participant-only replay, 21:00 ordering, 05:00 reset, offline catch-up, repeated/concurrent claims, capacity rollback and unchanged wallets.
- **Not executed:** Go compilation, real SQLite/MySQL/PostgreSQL tests, host worker integration or deployment. This workspace does not contain a compatible host module, Docker or disposable database services. The new backend must pass those checks in its owning host before release; mock and formatting checks do not establish backend compatibility.

All local Node checks ran serially with one V8 pool worker and an affinity mask covering six of twelve CPUs. No production service settings were changed.

## Arena Interface Update — 2026-11-26

- `npm test`: 105 browser-interface assertions and 56 tests/subtests passed. Added checks cover podium escaping, ranking pagination and self-location, empty refreshed rankings, squad comparison, battle-result filters and compatible-room filtering. Existing match consent, escrow/refund display and replay checks still pass.
- Headless Edge DOM geometry checks passed for the queue, ranking, squad-comparison dialog, history and rooms at widths of 320, 390, 768, 1024 and 1440 pixels (25 cases). Image requests were blocked and image sources removed from the test snapshots; no screenshots or image inspection were used. These checks verify horizontal fit, not artwork appearance.
- Local checks ran serially under a six-CPU affinity mask on a twelve-CPU machine, with one V8 pool worker. No backend code, database schema or production service changed; no Go build or live-host integration test was performed for this interface update.

The arena update adapts the reference project's navigation, opponent-card layout, ranking browsing and record filtering to the existing voluntary PvP endpoints. It does not implement the reference prototype's local rank-swapping, sweeps or rank-reward economy.

Validation date: 2026-11-25. This export contains only the card game; no complete host application or production deployment is included.

## Checks Actually Run

- Windows, Node.js 24.22.3: 93 browser-interface assertions passed.
- Node's test runner: 56 tests/subtests passed, zero failed, covering browser login gates, server-owned state, intent-only requests, expiry/settlement, readability and retained upstream attribution.
- Python 3.13.2: 19 preview/release tests passed, including anonymous gameplay rejection, no local wallet writes, static-source/environment/data blocking, path traversal, Host/Origin checks, private-marker rejection, complete-host exclusion and ZIP hash manifests.
- Python/JSON syntax checks, shared CPU-wrapper shell syntax and game-only CI YAML/boundaries passed. The public CI workflow was inspected locally, not claimed to have run on GitHub.
- The release source audit checked 148 selected game files against 14 known private deployment markers. Four synthetic fixture matches are retained by exact hash across three reviewed files; credential values are not recorded in the review manifest.

Local browser test processes were started serially with affinity `0F` on an eight-CPU machine, limiting their process tree to four CPUs. Node used one V8 pool worker and sequential test-file execution. No production build or service CPU setting was changed.

## Game And Host Boundary

The 14 game backend source/catalog/test files are byte-identical to the previously validated game implementation. Earlier full-host validation exercised the game using real SQLite, MySQL and PostgreSQL; the full host, its images and broader platform tests are intentionally not distributed in this game-only package.

This export did **not** build or run a standalone Go game server. The existing Go integration requires the host session, wallet, cache and database contracts in `backend/INTEGRATION.md`. Browser mocks, a source scan and matching hashes are not a substitute for validating an adapter to a different host.

The original local game files retain their recorded hashes. The earlier oversized full-site export was preserved privately instead of deleting its changes. This correction used local files only and did not connect to or modify production services.

## Publication And Artwork

The generated source ZIP includes a per-file SHA-256 manifest and excludes host-platform code, production connections, environment/data files, logs, backups, installed dependencies and Git history. Known private markers are held outside the game-source directory and are not packaged.

The game's 28 character provenance entries and 95 WebP files are retained. Artwork permission remains a separate question documented in `ASSETS.md`; source validation does not certify an unconfirmed image license. At the 2026-11-25 validation snapshot, no public repository had been uploaded. Publishing this source does not deploy a production service. GitHub Actions results are reported by the workflow itself, not by the earlier local validation snapshot.
