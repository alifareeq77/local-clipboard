# local-clipboard

A compact Go app to sync clipboard text between Linux and phone on your local network.

## Features

- Responsive green-themed UI tuned for small phones (including iPhone-sized screens).
- Mobile send form + live latest clipboard view.
- Searchable history with pin-to-top behavior.
- One-tap copy button on each history card.
- SQLite-backed persistent history (`clipboard.db`).
- Linux clipboard watcher client (Wayland/X11 tools).

## Run

### 1) Start server

```bash
go run . server -addr :8080 -db clipboard.db
```

### 2) Start clipboard watcher on Linux

```bash
go run . client -server http://127.0.0.1:8080 -interval 1s
```

### 3) Open from phone

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
