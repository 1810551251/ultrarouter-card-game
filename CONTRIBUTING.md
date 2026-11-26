# Contributing To The Card Game

Contributions target the game only. Do not add the host dashboard, provider relay, QQ system or production deployment configuration to this repository.

- Describe the player-visible issue, expected behavior and tests in a focused change.
- The authoritative shared catalog is `backend/model/gacha_catalog.json`; keep browser display definitions compatible with server rules.
- The browser submits intent, never authoritative saves, victories, rewards, balance changes or ownership claims.
- Run `npm test` for game UI changes and `npm run test:release` for preview/release changes.
- Go tests run in your own compatible host checkout. Database behavior changes require real SQLite, MySQL and PostgreSQL verification, including upgrade and idempotency cases when migrations change. A browser mock or source scan is not a backend test.
- Preserve authenticated-session checks, request IDs, row locking, wallet escrow and committed-only cache updates. Never use real production accounts or data as fixtures.
- Use one shared half-CPU wrapper for builds and CPU-heavy tests; do not run uncapped parallel jobs. `scripts/build-half-cpu.sh --run` also caps foreground test containers.
- Run `python scripts/audit-release.py` before packaging. Never put connection details, credentials, logs or private database files into a PR, issue, screenshot or test output.

Retain the original new-api / QuantumNous identifiers, AGPL terms and notices. Record the source and permitted license for new art separately; AI generation alone does not establish redistribution terms.
