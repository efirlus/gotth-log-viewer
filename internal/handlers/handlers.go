package handlers

import (
	"fmt"
	lg "gotthlogviewer/cmd/logger"
	"gotthlogviewer/internal/services"
	"gotthlogviewer/internal/shared"
	"gotthlogviewer/internal/types"
	"gotthlogviewer/internal/view/components"
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

	lg.Debug(fmt.Sprintln("handling logs partial ",
		"program_filter ", filters.Program,
		"level_filter ", filters.Level,
		"search_filter ", filters.Search))

	// Make sure we're rendering LogList with the current filters
	return shared.Render(w, r, components.LogList(logs, filters))
}
