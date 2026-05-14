# myapp

A minimal Go HTTP server that compiles to both a **native binary** and a **WebAssembly module** served via a Service Worker. Intended as a clean starting point for future projects.

Uses [go-wasm-http-server](https://github.com/nlepage/go-wasm-http-server) and [sqlite-wasm](https://sqlite.org/wasm/doc/trunk/index.md).

---

## Project layout

```
myapp/
├── src/
│   ├── main.go           ← route registration
│   ├── handlers.go       ← API handlers + writeJSON helper
│   ├── middleware.go     ← logger, requireAuth middleware
│   ├── store.go          ← in-memory session store (mutex-safe, crypto/rand tokens)
│   ├── db.go             ← shared DB logic (bcrypt passwords)
│   ├── db_native.go      ← native: opens SQLite file via go-sqlite3
│   ├── db_wasm.go        ← WASM: registers the JS-backed SQL driver
│   ├── driver_wasm.go    ← database/sql driver that calls self.sqlExec in sw.js
│   ├── serve_native.go   ← native: openDB + initDB + net/http listener
│   ├── serve_wasm.go     ← WASM: openDB + initDB + wasmhttp.Serve
│   └── version.go        ← /api/version handler, version/buildTime vars
│
├── assets_embedded/      ← embedded into the native binary via //go:embed *
│   ├── assets.go         ← embed package declaration
│   ├── index.html        ← main app UI (login/signup)
│   ├── style.css         ← shared stylesheet
│   └── wasm_loader.html  ← WASM entry page: registers the Service Worker
│
├── assets_static/        ← never embedded; copied to dist/ by `make wasm`
│   ├── sw.js             ← Service Worker: loads Go WASM + exposes self.sqlExec
│   ├── sqlite-wasm-*/    ← vendored sqlite-wasm JS/WASM
│   └── cdn.jsdelivr.net/ ← vendored wasm_exec.js + go-wasm-http-server SW helper
│
├── go.mod / go.sum
├── Makefile
└── mimetypes             ← MIME type overrides for darkhttpd (WASM content-type)
```

---

## Quick start

```bash
# Build everything (native binary + WASM module)
make

# Run native server → http://localhost:8080
make run

# Run WASM version  → http://localhost:8000  (requires darkhttpd)
make run-wasm
```

### Make targets

| Target | Description |
|---|---|
| `make` / `make build` | Build both native binary and WASM module |
| `make native` | Build `./myapp` only |
| `make wasm` | Build `./dist/a.wasm` + copy assets to `./dist/` |
| `make run` | Build + start native server on `:8080` |
| `make run-wasm` | Build WASM + serve `./dist/` with darkhttpd on `:8000` |
| `make clean` | Remove `./myapp` and `./dist/` |
| `make help` | List all targets |

---

## Build targets

### Native

A standard Go binary. SQLite is accessed via [go-sqlite3](https://github.com/mattn/go-sqlite3) (CGO). The database file is `a.db` in the working directory.

```bash
make native   # produces ./myapp
make run      # build + start
```

### WASM

The Go program compiles to `a.wasm` and runs inside a Service Worker in the browser. HTTP requests are intercepted by [go-wasm-http-server](https://github.com/nlepage/go-wasm-http-server) and routed to the Go handlers without ever leaving the browser.

SQLite runs via [sqlite-wasm](https://sqlite.org/wasm/doc/trunk/index.md) — `self.sqlExec` is exposed as a global in `sw.js` and called by Go through `syscall/js`. The database is currently **in-memory** and is lost when the Service Worker is terminated. See `sw.js` for the OPFS TODO.

```bash
make wasm       # produces dist/
make run-wasm   # build + serve
```

> The WASM build requires `darkhttpd` with `Cross-Origin-Opener-Policy` and `Cross-Origin-Embedder-Policy` headers — these are set automatically by `make run-wasm`. These headers are required for `SharedArrayBuffer`, which sqlite-wasm depends on.

---

## API

| Method | Path | Auth required | Description |
|---|---|---|---|
| `GET` | `/api/version` | No | Build version and timestamp |
| `POST` | `/api/signup` | No | Create a new user |
| `POST` | `/api/login` | No | Authenticate, receive a session token |
| `GET` | `/api/me` | No* | Returns `logged_in` + `user` if token is valid |

\* `/api/me` accepts an optional Bearer token and returns `logged_in: false` rather than 401 when it is missing or invalid. It is used as a session probe on page load.

Passwords are hashed with bcrypt. Session tokens are 128-bit random hex strings stored in an in-memory map — they are lost on server restart.

---

## Frontend

The frontend is plain HTML + JS with no build step or framework.

### Shared client (`app.js`)

```js
api(path, method, body)   // fetch wrapper — attaches Bearer token if present
requireLogin()            // redirects to /index.html if not logged in; returns { user }
getToken() / setToken() / clearToken()
```

### Adding a new page

1. Create `assets_embedded/mypage.html`.
2. Include `app.js` and call `requireLogin()` if the page requires auth.
3. Rebuild — the file is served at `/mypage.html` automatically, no Go changes needed.

```html
<script src="/app.js"></script>
<script>
    requireLogin().then(({ user }) => {
        // page code here
    });
</script>
```

### Adding a public page

Omit `requireLogin()` — `api()` still attaches the token if one exists.

---

## Adding a new API route

**1. Add a handler in `src/` (new file per feature area):**

```go
// src/notes.go
func listNotes(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]any{"notes": []string{}})
}
```

**2. Register the route in `src/main.go`:**

```go
// Public:
http.HandleFunc("/api/notes", logger(listNotes))

// Protected (requires valid session token):
http.HandleFunc("/api/notes", logger(requireAuth(listNotes)))
```

**3. Add a DB table if needed — in `src/db.go`'s `initDB()`:**

```go
if _, err := sqlDB.Exec(`
    CREATE TABLE IF NOT EXISTS notes (
        id       INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL,
        body     TEXT NOT NULL
    )
`); err != nil {
    log.Fatalf("initDB: CREATE TABLE notes: %v", err)
}
```

Add the SQL helper functions in a new `src/db_notes.go` file — keep `db.go` to schema only.

---

## Notes

**Sessions are in-memory.** On server restart, all session tokens are invalidated. The browser will send the stale token from `localStorage`, `/api/me` will return `logged_in: false`, and the login screen will be shown. The token is overwritten on next login.

**WASM transactions are not supported.** `wasmConn.Begin()` returns an error. Avoid ORMs or patterns that wrap operations in transactions when targeting WASM.

**Version stamping.** The build version and timestamp are injected at compile time via `-ldflags`. When running with `go run` directly (bypassing the Makefile) they fall back to `"dev"` and `"unknown"`.
