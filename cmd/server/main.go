package main

import (
	"crypto/tls"
	"encoding/hex"
	"fmt"
	lg "gotthlogviewer/cmd/logger"
	"gotthlogviewer/internal/auth"
	"gotthlogviewer/internal/handlers"
	"gotthlogviewer/internal/services"
	"gotthlogviewer/internal/shared"

	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Print directly to stderr for maximum visibility
	os.Stderr.WriteString("Starting application...\n")

	// environmental variable get
	if err := loadEnvironment(); err != nil {
		fmt.Fprintf(os.Stderr, "environment initialization failed: %v\n", err)
		os.Exit(1)
	}
	os.Stderr.WriteString("Environment loaded successfully\n")

	// credential hash load
	storedHash, err := loadCredentialsHash()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load credential hash: %v\n", err)
		os.Exit(1)
	}
	os.Stderr.WriteString("Credential loaded successfully\n")

	// Check certificate files
	certFile := os.Getenv("CERT_FILE")
	keyFile := os.Getenv("KEY_FILE")

	// Try to read cert file
	_, err = os.ReadFile(certFile)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to read certificate file: %v\n", err))
		os.Exit(1)
	}
	os.Stderr.WriteString("Successfully read certificate file\n")

	// Try to read key file
	_, err = os.ReadFile(keyFile)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to read key file: %v\n", err))
		os.Exit(1)
	}
	os.Stderr.WriteString("Successfully read key file\n")

	// Initialize structured logging
	if err := lg.Initialize(os.Getenv("LOG_PATH"), "gotth-log-viewer"); err != nil {
		fmt.Fprintf(os.Stderr, "logger initialization failed: %v\n", err)
		os.Exit(1)
	}
	os.Stderr.WriteString("Logger initialized successfully\n")

	logService := services.NewLogService(os.Getenv("LOG_PATH"))
	handler := handlers.NewLogHandler(logService)

	authHandler := auth.NewHandler(storedHash)

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Auth routes
	mux.HandleFunc("/auth/login", shared.Make(func(w http.ResponseWriter, r *http.Request) error {
		lg.Debug("received request to /auth/login")
		return authHandler.ServeLogin(w, r)
	}))

	// Protected routes wrapped with auth check
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lg.Debug(fmt.Sprintf("Received request to protected route: %s", r.URL.Path))

		// Check for valid session cookie
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value == "" {
			lg.Debug("No valid session cookie found, redirecting to login")
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		lg.Debug("Valid session found, proceeding with request")

		// Route to appropriate handler based on path
		switch r.URL.Path {
		case "/":
			shared.Make(handler.HandleIndex)(w, r)
		case "/api/logs/partial":
			if r.Method == "GET" {
				shared.Make(handler.HandleLogsPartial)(w, r)
			}
		default:
			http.NotFound(w, r)
		}
	})

	// Add the protected handler to the mux
	mux.Handle("/", protectedHandler)
	mux.Handle("/api/logs/partial", protectedHandler)

	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
	}

	server := &http.Server{
		Addr:      os.Getenv("LISTEN_ADDR"),
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	os.Stderr.WriteString(fmt.Sprintf("Starting HTTPS server on %s\n", server.Addr))
	if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Server error: %v\n", err))
		os.Exit(1)
	}
}

func loadEnvironment() error {

	// Try to load .env file, but don't error if it doesn't exist
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	// Validate required environment variables
	required := []string{"LISTEN_ADDR", "LOG_PATH", "CERT_FILE", "KEY_FILE", "CRED_HASH"}
	missing := []string{}

	for _, env := range required {
		if os.Getenv(env) == "" {
			missing = append(missing, env)

		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("required environment variables not set: %v", missing)
	}

	return nil
}

func loadCredentialsHash() ([]byte, error) {
	// Get hash from environment
	hashStr := os.Getenv("CRED_HASH")
	if hashStr == "" {
		return nil, fmt.Errorf("CRED_HASH not set in environment")
	}

	// Decode hex string to bytes
	return hex.DecodeString(hashStr)
}

