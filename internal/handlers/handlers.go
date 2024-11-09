package handlers

import (
	"gotthlogviewer/internal/services"
	"gotthlogviewer/internal/shared"
	"gotthlogviewer/internal/types"
	"gotthlogviewer/internal/view/components"
	"log/slog"
	"net/http"
)

type LogHandler struct {
	logService *services.LogService
}

func NewLogHandler(logService *services.LogService) *LogHandler {
	return &LogHandler{
		logService: logService,
	}
}

func (h *LogHandler) handleLogs(r *http.Request) ([]types.LogEntry, types.LogFilters, error) {
	logs, err := h.logService.ReadLogs()
	if err != nil {
		return nil, types.LogFilters{}, err
	}

	filters := types.LogFilters{
		Search:  r.URL.Query().Get("search"),
		Level:   r.URL.Query().Get("level"),
		Program: r.URL.Query().Get("program"),
	}

	// Log for debugging
	slog.Info("received filter params",
		"program", filters.Program,
		"search", filters.Search,
		"level", filters.Level,
	)
	return logs, filters, nil
}

// getFiltersFromRequest extracts all filters from the request
func getFiltersFromRequest(r *http.Request) types.LogFilters {
	if err := r.ParseForm(); err != nil {
		slog.Error("failed to parse form", "error", err)
	}

	return types.LogFilters{
		Search:  r.FormValue("search"),
		Level:   r.FormValue("level"),
		Program: r.FormValue("program"),
	}
}

// HandleIndex renders the full page
func (h *LogHandler) HandleIndex(w http.ResponseWriter, r *http.Request) error {
	logs, err := h.logService.ReadLogs()
	if err != nil {
		return err
	}

	filters := types.LogFilters{
		Program: r.FormValue("program"),
		Level:   r.FormValue("level"),
		Search:  r.FormValue("search"),
	}

	return shared.Render(w, r, components.LogViewer(logs, filters))
}

// HandleLogsPartial handles all partial updates (both polling and filter changes)
func (h *LogHandler) HandleLogsPartial(w http.ResponseWriter, r *http.Request) error {
	logs, err := h.logService.ReadLogs()
	if err != nil {
		return err
	}

	filters := types.LogFilters{
		Program: r.FormValue("program"),
		Level:   r.FormValue("level"),
		Search:  r.FormValue("search"),
	}

	slog.Info("handling logs partial",
		"program_filter", filters.Program,
		"level_filter", filters.Level,
		"search_filter", filters.Search)

	// Make sure we're rendering LogList with the current filters
	return shared.Render(w, r, components.LogList(logs, filters))
}
