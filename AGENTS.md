# Project Operating Rules

- Publish only the card game. The complete host website, account platform, provider relay, QQ system and production deployment configuration are outside this repository's scope.
- Keep the backend release allowlist limited to game models/controllers and integration examples. Authentication, wallet and database services are external host dependencies, not bundled platform source.
- Preserve existing local and remote changes. Back up and compare hashes before deployment.
- Keep gameplay, rewards, and account data server-authoritative. Never trust local saves.
- Static game CSS/HTML/JS changes do not require a backend image rebuild.
- Build and CPU-heavy test jobs must share at most half of the server's available CPU capacity.
- Use `sh scripts/build-half-cpu.sh [docker build arguments]` on the build server.
- On a shared build host, set `GACHA_BUILD_WRAPPER` to the absolute path of its shared wrapper. Keep host-specific paths in private environment configuration, not in published source.
- Use the same wrapper with `--run` for foreground test containers. It serializes jobs with the build lock and applies the same hard CPU quota.
- Example: `sh scripts/build-half-cpu.sh --run --rm -e GOMAXPROCS=1 TEST_IMAGE go test -p 1 -parallel 1 ./...`.
- Keep Go builds at `GOMAXPROCS=1 go build -p 1`; keep Go tests at `GOMAXPROCS=1 go test -p 1 -parallel 1` inside the capped container.
- Do not run other builds or CPU-heavy tests concurrently outside this wrapper. Coordinate all agents before starting one.
- `nice`, Go parallelism flags, and BuildKit flags alone are not CPU hard limits. The wrapper intentionally uses the supported classic Docker builder.
- Never apply these build limits to the production service or change global Docker settings.
- Never print credentials, session tokens, or secret-bearing configuration.
