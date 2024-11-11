package auth

import (
	"fmt"
	lg "gotthlogviewer/cmd/logger"
	auth "gotthlogviewer/internal/auth/views"
	"net/http"
	"time"
)

type Handler struct {
	storedHash []byte // The pre-stored hash to compare against
}

func NewHandler(storedHash []byte) *Handler {
	return &Handler{
		storedHash: storedHash,
	}
}

func (h *Handler) ServeLogin(w http.ResponseWriter, r *http.Request) error {
	lg.Debug(fmt.Sprintf("Handling login request: %s", r.Method))

	// Serve login page for GET requests
	if r.Method == http.MethodGet {
		lg.Debug("serving login page")
		return auth.LoginPage().Render(r.Context(), w)
	}

	// Handle login attempts for POST requests
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		passphrase := r.FormValue("passphrase")

		lg.Debug(fmt.Sprintf("login attempt for username : %s", username))

		// Generate hash from submitted credentials
		hash, err := HashCredentials(username, passphrase)
		if err != nil {
			lg.Error("hash generation failed", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return err
		}

		// Compare with stored hash
		if !SecureCompare(hash, h.storedHash) {
			// Return to login page on failure
			// Note: It's good practice to not specify whether username or password was incorrect
			lg.Debug("login failed - invalid credential")
			//w.Header().Set("HX-Redirect", "/auth/login")
			return auth.LoginPage().Render(r.Context(), w)
		}

		lg.Debug("login successful - setting cookie and redirect")

		// Set secure session cookie on success
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    username, // In production, use a secure session token instead
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   int(24 * time.Hour.Seconds()),
		})

		// Redirect to main page
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		return nil
	}
	lg.Debug("invalid request method")
	return nil
}
