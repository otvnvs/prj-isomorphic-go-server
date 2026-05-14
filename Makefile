# ---------------------------------------------------------------------------
# Makefile — build and run the Go auth server
#
#   make / make build    → build both native binary and WASM module
#   make native          → build native binary only  (./myapp)
#   make wasm            → build WASM module only     (./dist/a.wasm)
#   make run             → build & run the native server
#   make run-wasm        → build WASM & serve dist/ in the browser
#   make clean           → remove build artefacts
#   make help            → list targets
# ---------------------------------------------------------------------------

SRC_DIR            := ./src
ASSETS_EMBEDDED    := ./assets_embedded
ASSETS_STATIC      := ./assets_static
DIST_DIR           := ./dist

BINARY   := myapp
WASM_OUT := $(DIST_DIR)/a.wasm

VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/-dirty//' || echo dev)

BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

GO        := go
DARKHTTPD := darkhttpd
MIMETYPES := mimetypes
PORT      := 8000

CORS_FLAGS := \
  --header "Cross-Origin-Opener-Policy: same-origin" \
  --header "Cross-Origin-Embedder-Policy: require-corp"

.PHONY: all build native wasm run run-wasm clean help

all: build

## build: compile both native binary and WASM module
build: native wasm

## native: compile the native server binary (./myapp)
native:
	@echo "→ Building native binary: $(BINARY) ($(VERSION) @ $(BUILD_TIME))"
	$(GO) build -C $(SRC_DIR) $(LDFLAGS) -o ../$(BINARY) .
	@echo "✓ $(BINARY) ready"

## wasm: compile the WASM module into dist/ alongside assets
wasm: $(DIST_DIR)
	@echo "→ Building WASM module: $(WASM_OUT) ($(VERSION) @ $(BUILD_TIME))"
	GOOS=js GOARCH=wasm $(GO) build -C $(SRC_DIR) $(LDFLAGS) -o ../$(WASM_OUT) .
	@echo "→ Copying assets to $(DIST_DIR)/"
	cp $(ASSETS_EMBEDDED)/wasm_loader.html $(DIST_DIR)/index.html
	cp $(ASSETS_STATIC)/sw.js             $(DIST_DIR)/
	rm -rf $(DIST_DIR)/sqlite-wasm-3530100
	cp -r  $(ASSETS_STATIC)/sqlite-wasm-3530100 $(DIST_DIR)/
	rm -rf $(DIST_DIR)/cdn.jsdelivr.net
	cp -r  $(ASSETS_STATIC)/cdn.jsdelivr.net $(DIST_DIR)/
	@echo "✓ WASM build ready in $(DIST_DIR)/"

## run: build and start the native server on :8080
run: native
	@echo "→ Starting native server on http://localhost:8080"
	./$(BINARY)

## run-wasm: build WASM and serve dist/ with darkhttpd
run-wasm: wasm
	@echo "→ Serving WASM app on http://localhost:$(PORT)"
	$(DARKHTTPD) $(DIST_DIR) --port $(PORT) $(CORS_FLAGS) --mimetypes $(MIMETYPES)

## clean: remove compiled artefacts
clean:
	@echo "→ Cleaning build artefacts"
	rm -f $(BINARY)
	rm -rf $(DIST_DIR)
	@echo "✓ Clean"

## help: list available targets
help:
	@grep -E '^##' Makefile | sed 's/## /  /'

$(DIST_DIR):
	@mkdir -p $(DIST_DIR)
