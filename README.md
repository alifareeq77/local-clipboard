# local-clipboard

A tiny Go app that helps you share clipboard text between your Linux laptop and phone over your local network.

## What it does

- Starts a local HTTP server with a mobile-friendly web page.
- Accepts clipboard text from your phone browser.
- Runs a Linux clipboard watcher client that detects local copy changes and pushes them to the server automatically.
- Optionally writes remote clipboard changes back to Linux clipboard when write command is available.

## Run

### 1) Start server on laptop

```bash
go run . server -addr :8080
```

### 2) Start Linux clipboard watcher

```bash
go run . client -server http://127.0.0.1:8080 -interval 1s
```

### 3) Open from phone

Find your laptop LAN IP (for example `192.168.1.20`) and open:

```text
http://192.168.1.20:8080
```

Then paste text in the page and tap **Send to server**.

## Linux dependencies

Install one clipboard tool:

- Wayland: `wl-clipboard` (`wl-paste` / `wl-copy`)
- X11: `xclip` or `xsel`

If write support is unavailable, the client still uploads local copy events to the server.
