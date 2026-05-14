package main

import (
	"net/http"
	"myapp/assets_embedded"
)

func main() {
	http.HandleFunc("/api/version", logger(versionInfo))
	http.HandleFunc("/api/signup", logger(signup))
	http.HandleFunc("/api/login", logger(login))
	http.HandleFunc("/api/me", logger(checkAuth))
	http.HandleFunc("/api/sampleRequiringToken", logger(requireAuth(sampleRequiringToken)))
	http.HandleFunc("/api/sampleNotRequiringToken", logger(sampleNotRequiringToken))
	http.Handle("/", http.FileServerFS(assets_embedded.FS))
	startServer()
}
