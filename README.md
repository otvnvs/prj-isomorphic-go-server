# WASM Auth Server

A Go HTTP server that compiles to both a **native binary** and a **WebAssembly module** served via a Service Worker.

Uses [go-wasm-http-server](https://github.com/nlepage/go-wasm-http-server).

---

## Project layout

```
.
├── Makefile
├── go.mod / go.sum
├── assets_unminified/    ← unminified assets
│   ├── index.html        ← auth UI
│   ├── style.css         ← stylesheet
│   ├── sw.js             ← Service Worker: loads WASM + IndexedDB helpers
│   └── wasm_loader.html  ← entry page that registers the Service Worker
├── assets/               ← assets_unminified minified into this directory
│   └── assets.go         ← embed package — recursively embeds all files in this directory
└── src/
    ├── main.go           ← route registration
    ├── handlers.go       ← API handlers
    ├── middleware.go     ← logging middleware
    ├── store.go          ← thread-safe in-memory user/session store
    ├── serve_native.go   ← native build: disk persistence + net/http listener
    └── serve_wasm.go     ← WASM build:  IndexedDB persistence + wasmhttp.Serve
```

---

## Quick start

```bash
# Build everything
make build

# Run the native server  →  http://localhost:8080
make run-native

# Run the WASM version   →  http://localhost:8000
make run-wasm            # requires darkhttpd
```

### Available make targets

| Target | Description |
|---|---|
| `make build` | Build both native binary and WASM module |
| `make native` | Build `./my-server` only |
| `make wasm` | Build `./dist/a.wasm` + copy assets to `./dist/` |
| `make run-native` | Build + start the native server on `:8080` |
| `make run-wasm` | Build WASM + serve `./dist/` with darkhttpd on `:8000` |
| `make clean` | Remove `./my-server` and `./dist/` |
| `make backup`| Backup project to `./bak` |

---

## Adding a new asset

Drop a file into `assets/` and rebuild. That's it.

```
assets/
└── my-new-file.js   ← add it here
```

```bash
make build
```

The file is immediately available at `/<filename>` — no route registration, no
changes to any Go file required.

### How it works

`assets/assets.go` uses a `*` glob to embed every file in the directory:

```go
package assets

import "embed"

//go:embed *
var FS embed.FS
```

`//go:embed *` is a compiler directive: at build time Go walks the `assets/`
directory, reads every file it finds, and stores them inside the binary. At
runtime the originals don't need to exist — the data is in memory.

`src/main.go` mounts the entire `embed.FS` as a static file server on `/`:

```go
http.Handle("/", http.FileServerFS(assets.FS))
```

`http.FileServerFS` maps any URL path directly to a filename in the FS, so
`/style.css` → `assets.FS["style.css"]`, `/my-new-file.js` →
`assets.FS["my-new-file.js"]`, and so on — automatically, for every file in
the directory. Content-Type detection, ETags, `Last-Modified`, and range
requests are all handled for free.

The API routes (`/api/*`) are registered before the catch-all and take
priority, so there is no conflict between static files and API endpoints.

### Subdirectories

Subdirectories are embedded recursively. A file at `assets/fonts/Inter.woff2`
is served at `/fonts/Inter.woff2` with no extra configuration.

### Files excluded from embedding

`//go:embed *` skips files and directories whose names begin with `.` or `_`
(e.g. `.DS_Store`, `_notes/`). To include those as well, use `all:*` instead:

```go
//go:embed all:*
var FS embed.FS
```

---

## Architecture notes

### Build tags

`serve_native.go` and `serve_wasm.go` are selected at compile time via Go build tags:

| File | Tag | What it provides |
|---|---|---|
| `serve_native.go` | `!js \|\| !wasm` | `startServer()`, `persistUser()` (disk/JSON) |
| `serve_wasm.go` | `js && wasm` | `startServer()`, `persistUser()` (IndexedDB via JS interop) |

### Persistence

- **Native**: users are stored as JSON in `a.db` (loaded on startup, written on every signup).
- **WASM**: users are stored in **IndexedDB** inside the browser. The Service Worker exposes `saveUserToDB` and `loadUsersFromDB` globals that Go calls via `syscall/js`.

### Module layout

| File | Responsibility |
|---|---|
| `main.go` | Route wiring only |
| `handlers.go` | API handlers |
| `middleware.go` | `logger` middleware |
| `store.go` | All shared state behind a typed, mutex-safe API |
| `serve_*.go` | Platform-specific startup and persistence |
| `assets/assets.go` | Recursively embeds all static files into the binary |
