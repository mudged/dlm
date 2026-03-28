# dlm

Domestic Light & Magic — **one Go binary** serves the **JSON API** and the **Next.js static export** (Tailwind) embedded at build time. See `docs/architecture.md` and `AGENTS.md`.

## Prerequisites

- **Go 1.22+** (module `go` directive); the devcontainer uses **`mcr.microsoft.com/devcontainers/go:1-bookworm`** (current **Go 1.x** on rebuild).
- **Node.js** — **Active LTS** in the devcontainer (`features/node` with **`version: "lts"`**); only needed to build `web/` (not on the Pi at runtime).

## Production-style run (single process)

Bake the UI into the binary, then run:

```bash
cd web && npm run release:sync
cd ../backend && go run ./cmd/server
```

Open [http://127.0.0.1:8080/](http://127.0.0.1:8080/) — HTML and `/api/v1/...` are served by the same program.

Cross-compile for Raspberry Pi (from `backend/`):

```text
GOOS=linux GOARCH=arm64 go build -o bin/dlm-arm64 ./cmd/server
```

## Development (two processes, not shipped)

1. **Go API + embedded placeholder UI** (or re-sync after `release:sync`):

   ```bash
   cd backend
   go run ./cmd/server
   ```

2. **Next dev** (hot reload; browser uses rewrites to `:8080` for `/api/v1`):

   ```bash
   cd web
   npm run dev
   ```

   Open [http://localhost:3000](http://localhost:3000). Set `CORS_ALLOWED_ORIGINS=http://localhost:3000` on the Go process if you call the API from the browser without rewrites.

## Commands

| Area | Command |
|------|---------|
| Sync export → embed tree | `cd web && npm run release:sync` |
| Go tests | `cd backend && go test ./...` |
| Web lint | `cd web && npm run lint` |
| Web export only | `cd web && npm run build` (output in `web/out/`) |

## License

See [LICENSE](LICENSE).
