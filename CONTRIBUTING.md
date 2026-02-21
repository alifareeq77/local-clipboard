# Contributing to local-clipboard

Thanks for your interest in contributing. This guide will help you get set up and submit changes.

## Development setup

### Prerequisites

- **Go** (1.21+)
- **Node.js** (18+) and npm (for the Vue UI in `web/`)
- **SQLite** (used by the server; no extra install needed on most systems)
- **Clipboard tools** (for the Linux client):
  - Wayland: `wl-clipboard` (`wl-paste`, `wl-copy`)
  - X11: `xclip` or `xsel`

### Clone and build

```bash
git clone https://github.com/alifareeq77/local-clipboard.git
cd local-clipboard
go build -o local-clipboard .
```

### Run locally

1. **Build the web UI** (recommended):

   ```bash
   cd web && npm install && npm run build && cd ..
   ```

2. **Run server and client together**:

   ```bash
   ./local-clipboard run -addr :8080 -db clipboard.db
   ```

   Or run them separately: `./local-clipboard server ...` and `./local-clipboard client -server http://127.0.0.1:8080`.

3. **Web UI development** (hot reload):

   ```bash
   cd web && npm run dev
   ```

   Then run the server with `-static ""` to use the embedded fallback, or point your browser at the Vite dev server. See [README](README.md) for full run options.

## Project structure

- `main.go` — CLI entrypoint (server / client / run)
- `internal/` — Go packages (server, client, clipboard, models)
- `web/` — Vue 3 SPA (Vite); build output in `web/dist`
- `docs/` — Additional documentation

## Submitting changes

1. **Fork** the repo and create a branch from `master`:
   ```bash
   git checkout -b feature/your-feature
   # or
   git checkout -b fix/your-fix
   ```

2. **Make your changes**
   - Keep commits focused and messages clear (e.g. `Add X`, `Fix Y`).
   - If you touch the web UI, run `npm run build` in `web/` and ensure the app still works.

3. **Run tests**
   ```bash
   go test ./...
   ```
   To verify the Docker image: `docker compose up -d --build` then open http://localhost:8080.

4. **Push** your branch and open a **Pull Request** against `master`.
   - Describe what you changed and why.
   - Reference any related issues.

5. **Code style**
   - Go: follow standard `gofmt` / `go vet`; the project uses typical Go style.
   - Vue/JS: keep the existing style in `web/` (Vue 3 Composition API, existing patterns).

## Reporting issues

- Use the GitHub issue tracker.
- Include: OS, Go/Node versions, steps to reproduce, and what you expected vs what happened.
- For feature ideas, describe the use case and how you’d expect it to work.

## Questions

Open an issue with the question label, or reach out via the links in the [README](README.md).

---

Thank you for contributing.
