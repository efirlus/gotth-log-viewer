package handlers

import (
	"gotthlogviewer/internal/shared"
	"gotthlogviewer/internal/view/components"
	"net/http"
)

func HandleRoot(w http.ResponseWriter, r *http.Request) error {
	return shared.Render(w, r, components.TestLayout())
}

// HTTP endpoints:
// GET /logs              -> Full page load
// GET /api/logs         -> JSON for initial load
// GET /api/logs/poll    -> HTMX partial update
// POST /api/logs/filter -> HTMX filtered results

/*
// handler/logs.go
type LogHandler struct {
	service services.LogService
	views   *components.Templates
}

func (h *LogHandler) ServeLogViewer(w http.ResponseWriter, r *http.Request) {
	logs, _ := h.service.FetchLogs(model.LogFilters{})
	programs, _ := h.service.GetUniquePrograms()
	h.views.LogViewer(logs, programs, model.LogFilters{}).Render(w)
}

func (h *LogHandler) HandleLogPolling(w http.ResponseWriter, r *http.Request) {
	since := r.URL.Query().Get("since")
	logs, _ := h.service.FetchLogs(model.LogFilters{Since: &since})
	h.views.LogEntries(logs).Render(w)
}
*/
