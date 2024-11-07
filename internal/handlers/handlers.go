package handlers

import (
	"gotthlogviewer/internal/services"
	"gotthlogviewer/internal/shared"
	"gotthlogviewer/internal/types"
	"gotthlogviewer/internal/view/components"
	"log/slog"
	"net/http"
)

type Handlers struct {
	logService *services.LogService
}

func LogHandler(logService *services.LogService) *Handlers {
	return &Handlers{
		logService: logService,
	}
}

func (h *Handlers) handleLogs(r *http.Request) ([]types.LogEntry, components.ViewFilters, error) {
	logs, err := h.logService.ReadLogs()
	if err != nil {
		return nil, components.ViewFilters{}, err
	}

	program := r.URL.Query().Get("program")

	// Log for debugging
	slog.Info("received filter params",
		"program", program,
		"search", r.URL.Query().Get("search"),
		"level", r.URL.Query().Get("level"),
	)

	filters := components.ViewFilters{
		Search:  r.URL.Query().Get("search"),
		Level:   r.URL.Query().Get("level"),
		Program: r.URL.Query().Get("program"),
	}

	return logs, filters, nil
}

func (h *Handlers) HandleIndex(w http.ResponseWriter, r *http.Request) error {

	logs, filters, err := h.handleLogs(r)
	if err != nil {
		return err
	}

	return shared.Render(w, r, components.LogViewer(logs, filters))
}

// Handler for HTMX partial updates
func (h *Handlers) HandleLogsSearch(w http.ResponseWriter, r *http.Request) error {
	logs, filters, err := h.handleLogs(r)
	if err != nil {
		return err
	}

	// Only render the log list component for HTMX requests
	return shared.Render(w, r, components.LogList(logs, filters))
}
