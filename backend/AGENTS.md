# Card Game Backend Extension Rules

- This directory contains game code only. Do not import the complete host application, dashboard, provider relay, QQ system or private deployment configuration into the release.
- Preserve new-api / QuantumNous identifiers, attribution and the original license and notice terms.
- Authentication and account/wallet/database adapters belong to the deployment owner's host. Never substitute client-supplied identities or local saves for server authority.
- Before changing authentication flows, read the applicable OWASP Authentication, Session Management and CSRF guidance; verify failure, replay, expiry and bypass cases.
- Keep wallet mutations, game state and PvP escrow transitions in the same transaction. Apply cache changes only after a committed transaction. Retain integer quota units, bounds and request idempotency.
- Run affected Go tests in a compatible, disposable host. Database behavior changes require real SQLite, MySQL and PostgreSQL coverage, and fresh/upgrade/idempotency cases for migration changes. Do not claim an unsupported host integration is tested.
- Use the host's JSON wrappers and shared dialect-safe row-lock helper; do not introduce incompatible database syntax.
- Keep Go builds at `GOMAXPROCS=1 go build -p 1` and tests at `GOMAXPROCS=1 go test -p 1 -parallel 1` inside the shared hard-capped container. Do not run other CPU-heavy jobs outside the shared half-CPU wrapper.
