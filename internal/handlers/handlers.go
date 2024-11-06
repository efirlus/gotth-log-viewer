package handlers

import (
	"gotthlogviewer/internal/services"
	"gotthlogviewer/internal/shared"
	"gotthlogviewer/internal/view/components"
	"net/http"
)

func HandleIndex(w http.ResponseWriter, r *http.Request) error {
	logReader := services.NewLogReader("test.app.log")

	logs, err := logReader.ReadLogs()
	if err != nil {
		return err
	}

	return shared.Render(w, r, components.LogViewer(logs))
}
