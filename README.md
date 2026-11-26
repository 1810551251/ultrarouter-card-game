# Starfrost Card Game

This repository contains **only the card game**, not the maintainer's website or the complete new-api platform.

Included: the browser game, 28 character definitions, game artwork, server-owned game rules, dungeon battles, progression, inventory, token shop, rankings, voluntary PvP matchmaking, custom rooms, and game regression tests.

The ranked arena is a separate, free mode: register at the bottom of a ladder initially containing 100 NPC defenders, challenge a saved defense, and exchange positions when defeating a higher-ranked opponent. Five daily attempts reset at 05:00 Beijing time, with a ten-second cooldown between challenges. At 21:00 Beijing time, rank determines a daily material entitlement; players claim it into their game inventory. Saved defenses, rank changes, battle replays, offline settlement catch-up and claims are server-owned. Existing E matchmaking and custom rooms remain independent and do not change ladder ranks or consume ranked attempts.

The interface includes opponent cards, lineup comparisons, defense snapshot updates, current/best rank, a top-three podium, ranking pages, attack/defense history, a settlement countdown, reward tiers and claim records. It retains the game's cream, navy and gold styling. The ranked arena requires the new game tables, routes and host-owned settlement worker described in `backend/INTEGRATION.md`; updating static files alone does not enable the backend feature. Sweeps, purchased attempts and first-rank milestone rewards are not part of this implementation.

Not included: the website dashboard, account-management implementation, API-provider relay, QQ bot/binding system, payment platform, production deployment scripts, server addresses, credentials, real accounts, or database files.

## What Can Run Locally

Use Node.js 24+ and Python 3.10+:

```sh
npm ci
npm test
npm run test:release
python scripts/preview.py
```

Open `http://localhost:8080/gacha/` for a **read-only frontend preview**. The preview does not log anyone in, create a wallet, or simulate payable rewards. Playing requires your own authenticated backend. Editing browser storage never creates valid game assets.

The Go files are a **game extension**, not a standalone server. They currently use the host's authenticated sessions, wallet transaction/cache helpers, and database connection. The host implementation is deliberately not bundled. See `backend/INTEGRATION.md` for the required integration points; copying the files into an arbitrary new-api release without adapting those dependencies is not a supported one-click installation.

## Source Layout

| Path | Game Content |
| --- | --- |
| `index.html`, `style.css`, `mobile.css`, `js/` | Game interface, server-state synchronization, battle playback |
| `assets/` | Character, enemy, background and UI artwork; AI provenance |
| `backend/model/gacha*.go` | Game state, rewards, battles, progression, guilds, matchmaking and settlement |
| `backend/model/gacha_catalog.json` | Authoritative character and stage configuration |
| `backend/controller/gacha*.go` | Session-gated game endpoints and intent validation |
| `backend/integration/` | Game-only route and migration examples for your own host |
| `tests/` | Browser authority/readability checks, preview and release-boundary checks |
| `scripts/` | Read-only preview, source audit, packaging and shared CPU-budget tooling |

All browser endpoints are same-origin relative paths. This project does not need or provide access to the original maintainer's server. Contributors supply their own test accounts and databases.

## Development And Release

Gameplay, rewards, inventory, E balances and settlement remain server-authoritative. Keep the host session guard, transaction boundaries, escrow, cache consistency and idempotency when adapting the backend. See `CONTRIBUTING.md`.

```sh
python scripts/audit-release.py
python scripts/package-release.py --name starfrost-game
```

The packager includes only game source and checks the game-only backend allowlist. Local data, secrets, dependencies and Git history are excluded. Private deployment-marker checks are described in `SECURITY.md`. Publish only the reviewed source allowlist or ZIP, never an entire host working directory. A new public repository should start with fresh Git history so private deployment history cannot be included.

Publishing this source does not deploy or connect to any production service. Check `VALIDATION.md` for the checks actually performed and the external-host limitation. The GitHub Actions workflow checks game source and frontend behavior; it does not deploy the game or prove compatibility with an arbitrary host backend.

## Attribution And Artwork

The existing game backend is based on [new-api / QuantumNous](https://github.com/QuantumNous/new-api). Its original identifiers and license/notice files are retained; preserving attribution does not mean the complete platform is included. Code follows the bundled GNU AGPLv3 terms and notices.

The maintainer confirmed that all 28 character images are AI-generated. Artwork licensing is separate from code licensing; see `ASSETS.md` and `assets/PROVENANCE.json`. Do not assume an unconfirmed image license permits commercial redistribution.
