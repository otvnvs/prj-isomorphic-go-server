# ---------------------------------------------------------------------------
# Makefile — build, run, and clean the Go auth server
#
#  make build         → build both native binary and WASM module
#  make native        → build native binary only  (./my-server)
#  make wasm          → build WASM module only     (./dist/a.wasm)
#  make run-native    → build & run the native server
#  make run-wasm      → build WASM & start a local file server for the browser
#  make clean         → remove build artefacts
#  make help          → list available targets
#  make backup        → backs up project
# ---------------------------------------------------------------------------

# ---- Paths -----------------------------------------------------------------
SRC_DIR    := ./src
DIST_DIR   := ./dist
BIN_DIR    := ./bin
ASSETS_DIR := ./assets_unminified
MIN_DIR    := ./assets

# ---- Output names ----------------------------------------------------------
BINARY     := $(BIN_DIR)/a.out
WASM_OUT   := $(DIST_DIR)/a.wasm

# ---- Version stamping ------------------------------------------------------
# VERSION: use the nearest git tag, or 'dev' if not in a git repo.
# BUILD_TIME: current UTC time in RFC3339 format.
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

TAGS       := 

# ---- Tools -----------------------------------------------------------------
MINIFY     := minify
GO         := go
DARKHTTPD  := darkhttpd
PORT       := 8000

# CORS headers required for SharedArrayBuffer / WASM threads
CORS_FLAGS := \
  --header "Cross-Origin-Opener-Policy: same-origin" \
  --header "Cross-Origin-Embedder-Policy: require-corp"

# ---- Phony targets ---------------------------------------------------------
.PHONY: all build native wasm run-native run-wasm clean help backup

all: build

## build: compile both native binary and WASM module
build: native wasm

## native: compile the native server binary
native: $(BIN_DIR) minify
	@echo "→ Building native binary: $(BINARY) ($(VERSION) @ $(BUILD_TIME))"
	@$(GO) build -C $(SRC_DIR) $(TAGS) $(LDFLAGS) -o ../$(BINARY) .
	@echo "✓ $(BINARY) ready"

## wasm: compile the WASM module into dist/ alongside assets
wasm: $(DIST_DIR) minify
	@echo "→ Building WASM module: $(WASM_OUT) ($(VERSION) @ $(BUILD_TIME))"
	@GOOS=js GOARCH=wasm $(GO) build -C $(SRC_DIR) $(TAGS) $(LDFLAGS) -o ../$(WASM_OUT) .
	@echo "→ Copying assets to $(DIST_DIR)/"
	@cp $(MIN_DIR)/sw.js            $(DIST_DIR)/sw.js
	@cp $(MIN_DIR)/wasm_loader.html $(DIST_DIR)/index.html
	@cp $(MIN_DIR)/favicon.ico $(DIST_DIR)/favicon.ico
	@echo "✓ WASM build ready in $(DIST_DIR)/"

## minify: shrink all assets recursively into assets.min
minify: $(MIN_DIR)
	@echo "→ Minifying assets recursively to $(MIN_DIR)..."
	@$(MINIFY) -r -s -o $(MIN_DIR) $(ASSETS_DIR)/
	@echo "✓ Minification complete"
	@echo "→ Injecting 'prod' build tag into minified assets..."
	@# This replaces '!prod' with 'prod' in the minified version only
	@sed -i 's/!prod/prod/g' $(MIN_DIR)/assets.go
	@echo "✓ Minification and tag injection complete"

## run-native: build and start the native HTTP server on :8080
run-native: native
	@echo "→ Starting native server on http://localhost:8080"
	@./$(BINARY)

## run-wasm: build WASM and serve the dist/ directory with darkhttpd
run-wasm: wasm
	@echo "→ Serving WASM app on http://localhost:$(PORT)"
	@$(DARKHTTPD) $(DIST_DIR) --port $(PORT) $(CORS_FLAGS)

## clean: remove compiled artefacts
clean:
	@echo "→ Cleaning build artefacts"
	@rm -f $(BINARY)
	@rm -rf $(DIST_DIR)
	find $(MIN_DIR) -maxdepth 1 -type f ! -name '*.go' -delete
	@echo "✓ Clean"

## help: list available targets
help:
	@grep -E '^##' Makefile | sed 's/## /  /'

# ---- Ensure dist/ exists ---------------------------------------------------
$(DIST_DIR):
	@mkdir -p $(DIST_DIR)

# ---- Ensure bin/ exists ----------------------------------------------------
$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

# ---- Ensure assets.min exists -----------------------------------------------
$(MIN_DIR):
	@mkdir -p $(MIN_DIR)

# ---- Favicon ---------------------------------------------------------------
favicon:
	@convert -size 256x256 xc:none -fill grey \
		-draw "circle 128,128 128,40"\
		-define icon:auto-resize=16,32,48 ./assets/favicon.ico
	

# ---- Backup ----------------------------------------------------------------
TIMESTAMP  := $(shell date +%Y%m%d%H%M%S%3N)
BAK_DIR     = ./bak
BACKUP_PATH = $(BAK_DIR)/$(TIMESTAMP)
## backup: backup up project to ./bak/YYYYMMDDHHMMSSsss
backup:
	@mkdir -p $(BACKUP_PATH)
	@cp ./README.md $(BACKUP_PATH)
	@cp ./Makefile $(BACKUP_PATH)
	@cp ./go.sum $(BACKUP_PATH)
	@cp ./go.sum $(BACKUP_PATH)
	@cp ./a.db $(BACKUP_PATH)
	@cp -r ./src $(BACKUP_PATH)
	@cp -r ./assets $(BACKUP_PATH)
	@cp Makefile $(BACKUP_PATH)
	@echo "✓ backed up to $(BACKUP_PATH)"
