# Contributing To The Card Game

Contributions target the game only. Do not add the host dashboard, provider relay, QQ system or production deployment configuration to this repository.

- Describe the player-visible issue, expected behavior and tests in a focused change.
- The character/stage catalog is `backend/model/gacha_catalog.json`. Keep card and artifact rules in `backend/model/gacha_cultivation.json`, five-slot gear rules in `backend/model/gacha_equipment.json`, and item quality/source/use descriptions in `backend/model/gacha_items.json`. The browser reads the workshop catalogs from the authenticated game profile; never use display text to determine a reward.
- The browser submits intent, never authoritative saves, victories, rewards, balance changes or ownership claims.
- Keep card breakthrough and effort/skill budgets in `backend/model/gacha_training.json`, wearable recipes in `backend/model/gacha_carry.json`. A new combat rule must preserve historical matchmaking and defense snapshots, and a new drop must be gated by the issued run version. Put a PR walkthrough in its own reviewed `docs/pr-*/` directory, separate from the game archive.
- The artifact-grade, vigil, cases and dispatch catalogs are separate from the item atlas. Keep puzzles as player decisions, not client-calculated reward claims. Daily actions carry the issued game date or run ID and an expected move number when needed; weekly maps carry the week and expected page. Only the player document codec may serialize the private cipher seed. Never include it in profiles, public receipts, frozen teams or logs.
- Run `npm test` for game UI changes and `npm run test:release` for preview/release changes.
- Go tests run in your own compatible host checkout. Database behavior changes require real SQLite, MySQL and PostgreSQL verification, including upgrade and idempotency cases when migrations change. A browser mock or source scan is not a backend test.
- Preserve authenticated-session checks, request IDs, row locking, wallet escrow and committed-only cache updates. Never use real production accounts or data as fixtures.
- Use one shared half-CPU wrapper for builds and CPU-heavy tests; do not run uncapped parallel jobs. `scripts/build-half-cpu.sh --run` also caps foreground test containers.
- Run `python scripts/audit-release.py` before packaging. Never put connection details, credentials, logs or private database files into a PR, issue, screenshot or test output.

Retain the original new-api / QuantumNous identifiers, AGPL terms and notices. Record the source and permitted license for new art separately; AI generation alone does not establish redistribution terms.
