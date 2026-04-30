package main

import (
	"net/http"
)

// These variables are set at build time via -ldflags:
//
//	-X main.version=1.2.3
//	-X main.buildTime=2024-01-01T00:00:00Z
//
// If the binary is built without those flags (e.g. plain `go run`),
// they fall back to the defaults below.
var (
	version   = "dev"
	buildTime = "unknown"
)

func versionInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"version":    version,
		"build_time": buildTime,
	})
}
