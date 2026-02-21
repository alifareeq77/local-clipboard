# King AI: Developer-Focused Prototype Prompt

Use this prompt with King AI (or any AI coding assistant) to generate a **concrete, runnable prototype** for an app idea.

**See also:** [Runway ML prompts](RUNWAY_ML_PROTOTYPE_PROMPT.md) for *visual* prototypes (UI mockups, concept video) that developers can implement from. Copy and fill the bracketed sections, then paste into the AI.

---

## The Prompt (copy below)

```
Create a working prototype for this app idea, with implementation details suitable for developers.

## App idea (one sentence)
[Describe the app in one clear sentence, e.g. "A small tool that syncs clipboard text between my Linux machine and phone on the local network."]

## Target users
Developers and technical users who will run it locally (CLI, single binary, or simple server).

## Required deliverables

1. **Backend**
   - Language: [e.g. Go preferred; or Node/Python if it fits better]
   - Single binary or single entrypoint where possible (e.g. `./app server`, `./app run`).
   - Clear API surface: list the HTTP endpoints (method, path, request/response shape) or equivalent (e.g. gRPC, stdio).
   - Persistent storage: [e.g. SQLite for local data; path/config for DB file].
   - No external SaaS or cloud dependencies unless strictly required; prefer local-first.

2. **Frontend (if the app has a UI)**
   - [e.g. Vue 3 SPA, or simple HTML/JS, or TUI in Go.]
   - Build steps: e.g. `cd web && npm install && npm run build`; backend serves built assets (e.g. from `web/dist`).
   - One primary layout: [e.g. dark theme, single accent color, readable font].
   - Mobile-friendly if users will open it from phone/tablet (e.g. responsive, touch-friendly).

3. **Data model**
   - Define the main entities and their fields (e.g. id, created_at, source, content).
   - Mention indexes or constraints that matter for the core use case (e.g. unique, pinned flag).

4. **CLI / runtime**
   - Subcommands and flags (e.g. `server -addr :8080 -db path`, `client -server URL`, `run` to run both).
   - How to run locally: exact commands to build, run server, run client (if any), and open the UI.
   - Any env vars or config files (path, format) if needed.

5. **README**
   - Short description, features list, build/run instructions, and a minimal "API" section listing endpoints and payloads.
   - Optional: one paragraph on "how to extend" (e.g. add new API route, add new UI page).

## Constraints
- No placeholder or "TODO" implementations for core flows: the prototype must run end-to-end (e.g. submit form → API → DB → show in UI).
- Prefer standard libraries and minimal dependencies; avoid frameworks that require heavy setup.
- Code should be in a flat or shallow structure (e.g. `internal/` for backend, `web/` for frontend) with clear separation.

## Output format
- Provide full file contents for the main entrypoint, server, handlers, DB layer, and at least one representative frontend component/page.
- Use concrete names (e.g. `clipboard.db`, `POST /api/clipboard`) rather than "your-db" or "your-endpoint".
- Include a `package.json` (if frontend) and `go.mod` (if Go) with exact versions so a developer can `go build` / `npm run build` and run immediately.
```

---

## Example: Filled prompt for a “local clipboard sync” idea

```
Create a working prototype for this app idea, with implementation details suitable for developers.

## App idea (one sentence)
A compact app to sync clipboard text between Linux and phone on the local network.

## Target users
Developers and technical users who will run it locally (CLI, single binary, or simple server).

## Required deliverables

1. **Backend**
   - Language: Go.
   - Single binary: `./app server`, `./app client`, `./app run` (server + client in one process).
   - API: POST/GET /api/clipboard, GET /api/history?limit=&q=, POST /api/history/pin. JSON request/response.
   - SQLite for history (e.g. clipboard.db); configurable path.

2. **Frontend**
   - Vue 3 SPA in `web/`. Build: `cd web && npm run build`; Go serves `web/dist`.
   - Dark theme, one accent color (e.g. teal), readable font (e.g. Plus Jakarta Sans). Mobile-friendly.

3. **Data model**
   - Clipboard: latest text, source, updated_at. History: id, text, source, created_at, pinned (bool). Index on pinned, created_at.

4. **CLI**
   - server: -addr :8080 -db clipboard.db -static web/dist
   - client: -server http://127.0.0.1:8080 -interval 1s
   - run: same flags, starts server + client; print LAN URL for phone.

5. **README**
   - Description, features, build/run, API section with endpoints and example payloads.

## Constraints
- No placeholders for core flows; must run end-to-end.
- Minimal deps; standard library where possible.
- Structure: internal/server, internal/client, internal/history, web/.

## Output format
- Full contents for main.go, server, handlers, SQLite layer, and at least one Vue page (e.g. send form + history list).
- Concrete names: clipboard.db, /api/clipboard, etc.
- go.mod and web/package.json with versions for immediate go build / npm run build.
```

---

## Tips for developers

- **Narrow the idea**: One sentence keeps the prototype scoped; you can add “Phase 2” items in a follow-up.
- **Reuse this repo as reference**: Point the AI at your `local-clipboard` structure (main.go, internal/server, web/) and say “same style and layout as this project.”
- **Lock versions**: Ask for explicit versions in go.mod and package.json to avoid “works on my machine” issues.
- **API-first**: Defining endpoints and payloads first helps the AI generate consistent backend and frontend code.
