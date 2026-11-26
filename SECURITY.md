# Game And Publication Boundaries

The package contains game source, not the maintainer's server environment. Authentication, wallet management and database connections are external host dependencies.

## Runtime Authority

- The host must authenticate the browser session before any game endpoint runs. Never trust a user ID, token, save or role supplied in the request body.
- Calculate battle results, resources, rewards, shop prices and E transfers on the server.
- Keep state revisions, single-use request IDs, transaction boundaries, row locks and wallet escrow. Editing local storage must not grant assets or permission.
- Use HTTPS, secure HttpOnly session cookies, trusted origins, session rotation and logout invalidation in a real deployment. The local preview has no login or wallet capability.
- Verify authentication changes against the OWASP Authentication, Session Management and CSRF Cheat Sheets; do not claim a full security certification from these source checks.

## Source Release

```sh
python scripts/audit-release.py
python scripts/package-release.py --name starfrost-game
```

The release allowlist deliberately excludes the complete host platform. Adding a non-game backend file causes an audit failure. Actual environment files, databases, logs, SSH material, backups, installed dependencies and Git history are not packaged.

Credential-shaped test values are reviewed by exact file hash, not by blanket directory exemptions. Changed matches require another review. The scanner never prints credential values.

Known private addresses, domains and deployment paths can be recorded as a non-empty JSON string array in the **private** `.local/release-private-markers.json`. The scanner and packager load it automatically; alternatively use `GACHA_RELEASE_PRIVATE_MARKERS` or `--private-markers` to select an external private list. The list itself is excluded from the archive. Matching checks include filenames, binary metadata and UTF-8/UTF-16 content and cannot be approved through fixture reviews.

These are source-boundary and known-marker checks, not proof that a public website's address can never be discovered through DNS, ordinary requests or historical records. Open source must not provide the original server's credentials or authority; its public network exposure still needs separate protection.

Before creating a public Git repository, also review its history, attachments and CI settings. Packaging a clean source snapshot does not clean an existing repository's history.

References:

- https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
- https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
- https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html
