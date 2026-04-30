package main

import (
	"net/http"
	"example/assets"
)

func main() {
	http.HandleFunc("/api/version", logger(versionInfo))
	http.HandleFunc("/api/signup", logger(signup))
	http.HandleFunc("/api/login", logger(login))
	http.HandleFunc("/api/me", logger(checkAuth))
	http.Handle("/", http.FileServerFS(assets.FS))
	startServer()
}
