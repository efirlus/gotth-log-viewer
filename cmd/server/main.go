package main

import (
	"gotthlogviewer/internal/handlers"
	"gotthlogviewer/internal/services"
	"gotthlogviewer/internal/shared"

	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		logger.Error("error loading .env file", "error", err)
		os.Exit(1)
	}

	logService := services.NewLogService("test.app.log")
	handler := handlers.NewLogHandler(logService)

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	mux.HandleFunc("/", shared.Make(handler.HandleIndex))
	mux.HandleFunc("GET /api/logs/partial", shared.Make(handler.HandleLogsPartial))

	// Start server
	listenAddr := os.Getenv("LISTEN_ADDR")
	logger.Info("HTTP server started", "listenAddr", listenAddr)
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

/*
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Program   string `json:"program"`
	Location  string `json:"location,omitempty"`
}
	// Static files
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	// API routes
	mux.HandleFunc("GET /api/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// TODO: Implement log reading logic
	})

	// SSE endpoint for real-time updates
	mux.HandleFunc("GET /api/logs/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		// TODO: Implement streaming logic
	})

	// Main page
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement templ component rendering
		// component := views.LogViewerPage()
		// component.Render(r.Context(), w)
	})
*/
