# local-clipboard

A Go app to share clipboard text between your Linux laptop and your phone over local network.

## Features

- Local HTTP server with responsive mobile-friendly web UI.
- Linux clipboard watcher that automatically uploads new copy events.
- Mobile web form to manually push text from phone.
- SQLite-backed clipboard history (`/api/history`) for persistence and browsing.
- Optional remote-to-local clipboard apply when clipboard write command is available.

## Run

### 1) Start server on laptop

```bash
go run . server -addr :8080 -db clipboard.db
```

### 2) Start Linux clipboard watcher client

```bash
go run . client -server http://127.0.0.1:8080 -interval 1s
```

### 3) Open from phone

Use your laptop LAN IP (example `192.168.1.20`):

```text
http://192.168.1.20:8080
```

## API

- `POST /api/clipboard` with `{ "text": "...", "source": "mobile-web" }`
- `GET /api/clipboard` for latest value
- `GET /api/history?limit=25` for recent history

## Linux dependencies

Install one clipboard tool:

- Wayland: `wl-clipboard` (`wl-paste` and `wl-copy`)
- X11: `xclip` or `xsel`
