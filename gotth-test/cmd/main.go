package main

import (
	"gotthtest/handlers"
	"log/slog"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", public())
	mux.Get("/", handlers.Make(handlers.HandleHome))
	mux.Get("/login", handlers.Make(handlers.HandleLoginIndex))

	mux.HandleFunc()

	listenAddr := "5174"
	slog.Info("HTTP server started", "listenAddr", listenAddr)
	http.ListenAndServe(listenAddr, mux)
}
