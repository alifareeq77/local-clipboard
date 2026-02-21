# local-clipboard

A compact Go app to sync clipboard text between Linux and phone on your local network.

## Features

- **Vue 3 UI** in `web/`: modern SPA (Plus Jakarta Sans, teal accent, dark theme) with send form, latest clipboard, searchable history, pin, and copy. Build with `cd web && npm run build`; Go serves `web/dist` by default.
- Mobile send form + live latest clipboard view.
- Searchable history with pin-to-top behavior.
- One-tap copy button on each history card.
- SQLite-backed persistent history (`clipboard.db`).
- Linux clipboard watcher client (Wayland/X11 tools).

## Build (single binary)

Compile the server and client into one executable:

```bash
go build -o local-clipboard .
```

Run the server: `./local-clipboard server -addr :8080 -db clipboard.db`  
Run the client: `./local-clipboard client -server http://127.0.0.1:8080`

**Run both in one process (single binary):**

```bash
./local-clipboard run -addr :8080 -db clipboard.db
```

This starts the web server and the clipboard client in the same process. No need to run two terminals. The client automatically connects to the server (e.g. `http://127.0.0.1:8080`). Optional flags: `-interval 1s`, `-source my-pc`, `-no-build`.

When the server starts, it prints the LAN URLs (e.g. `open from phone: http://192.168.1.5:8080`). The web UI also shows **Open from phone:** with copyable URLs in the header.

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

You can set the port via the **PORT** environment variable (e.g. in a `.env` file; see `.env.example`). Use `-static ""` to skip the Vue app and use the embedded fallback HTML. Use `-static web/dist` (default) to serve the Vue SPA.

### 3) Start clipboard watcher on Linux

```bash
go run . client -server http://127.0.0.1:8080 -interval 1s
```

### 4) Open from phone

```text
http://<your-laptop-lan-ip>:8080
```

## Docker

The image runs **only the server** (no clipboard watcher; use the client on the host or send from phone).

**Build and run with Docker Compose (recommended):**

Copy `.env.example` to `.env` and set `PORT` if you want a different port (default 8080):

```bash
cp .env.example .env
# edit .env to set PORT=3000 etc.
docker compose up -d --build
```

Then open `http://localhost:8080` (or your chosen port) or `http://<your-lan-ip>:8080` from your phone. The SQLite DB is stored in a named volume `clipboard-data`.

**Plain Docker:**

```bash
docker build -t local-clipboard .
docker run -d -p 8080:8080 -v clipboard-data:/data --name local-clipboard local-clipboard
```

To override the DB path or listen address, override the entrypoint, e.g.:

```bash
docker run -d -p 8080:8080 -v clipboard-data:/data local-clipboard \
  /app/local-clipboard server -addr :8080 -db /data/clipboard.db -static /app/static -no-build
```

## iOS Shortcut: send clipboard (e.g. with Back Tap)

You can send the current clipboard from your iPhone using a Shortcut, and trigger it with **Back Tap** (double- or triple-tap the back of the phone).

### 1) Create the shortcut

1. Open the **Shortcuts** app.
2. Tap **+** to create a new shortcut.
3. Add **Get Clipboard**.
4. Add **Get Contents of URL**:
   - **URL:** `http://<your-laptop-lan-ip>:8080/api/clipboard` (same IP as in step 4 above).
   - **Method:** **POST**.
   - **Headers:** add `Content-Type` = `application/json`.
   - **Request Body:** **JSON** with key `text` = output of **Get Clipboard**, and optionally `source` = `iOS`.
5. Name the shortcut (e.g. **Send Clipboard**) and save.

### 2) Assign Back Tap

1. Open **Settings** → **Accessibility** → **Touch**.
2. Scroll to **Back Tap** and tap it.
3. Under **Double Tap** or **Triple Tap**, choose your shortcut (e.g. **Send Clipboard**).

### 3) Use it

Copy something, then double-tap (or triple-tap) the **back of your iPhone** to send the clipboard to the server.

**Note:** Back Tap requires iPhone 8 or later (iOS 14+). Your phone must be on the same Wi‑Fi as the machine running the server.

## API

- `POST /api/clipboard` → save latest clipboard
- `GET /api/clipboard` → get latest clipboard
- `GET /api/history?limit=80&q=keyword` → list/search history (pinned first)
- `POST /api/history/pin` with `{ "id": 4, "pinned": true }`

## Linux dependencies

Install one clipboard tool:

- Wayland: `wl-paste` + `wl-copy` (`wl-clipboard` package)
- X11: `xclip` or `xsel`

---

Created with ❤️ by [alifareeq](https://github.com/alifareeq77) · [LinkedIn](https://www.linkedin.com/in/ali-fareeq-1390351b0/)
