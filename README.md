# local-clipboard

A compact Go app to sync clipboard text between Linux and phone on your local network.

## Features

- **Vue 3 UI** in `web/`: modern SPA (Plus Jakarta Sans, teal accent, dark theme) with send form, latest clipboard, searchable history, pin, and copy. Build with `cd web && npm run build`; Go serves `web/dist` by default.
- Mobile send form + live latest clipboard view.
- Searchable history with pin-to-top behavior.
- One-tap copy button on each history card.
- SQLite-backed persistent history (`clipboard.db`).
- Linux clipboard watcher client (Wayland/X11 tools).

## Run

### 1) Build the Vue UI (recommended)

```bash
cd web && npm install && npm run build && cd ..
```

The server serves the built app from `web/dist` by default.

### 2) Start server

```bash
go run . server -addr :8080 -db clipboard.db
```

Use `-static ""` to skip the Vue app and use the embedded fallback HTML. Use `-static web/dist` (default) to serve the Vue SPA.

### 3) Start clipboard watcher on Linux

```bash
go run . client -server http://127.0.0.1:8080 -interval 1s
```

### 4) Open from phone

```text
http://<your-laptop-lan-ip>:8080
```

## API

- `POST /api/clipboard` → save latest clipboard
- `GET /api/clipboard` → get latest clipboard
- `GET /api/history?limit=80&q=keyword` → list/search history (pinned first)
- `POST /api/history/pin` with `{ "id": 4, "pinned": true }`

## Linux dependencies

Install one clipboard tool:

- Wayland: `wl-paste` + `wl-copy` (`wl-clipboard` package)
- X11: `xclip` or `xsel`
