package main

import (
	"fmt"
	lg "gotthlogviewer/cmd/logger"
	"gotthlogviewer/internal/handlers"
	"gotthlogviewer/internal/services"
	"gotthlogviewer/internal/shared"

	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// environmental variable get
	loadEnvironment()

	// Initialize structured logging
	if err := lg.Initialize(os.Getenv("LOG_PATH"), "gotth-log-viewer"); err != nil {
		os.Stderr.WriteString("Failed to initialize logger: " + err.Error() + "\n")
		os.Exit(1)
	}

	logService := services.NewLogService(os.Getenv("LOG_PATH"))
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
	lg.Info(fmt.Sprintf("HTTP server started at port %v", listenAddr))
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		lg.Error("server error", err)
		os.Exit(1)
	}
}

func loadEnvironment() {
	// Try to load .env file, but don't error if it doesn't exist
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	// Validate required environment variables
	required := []string{"LISTEN_ADDR", "LOG_PATH"}
	for _, env := range required {
		if os.Getenv(env) == "" {
			fmt.Println("required environment variable not set - no environment variable found")
			os.Exit(1)
		}
	}
}
