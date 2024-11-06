package types

type LogFilters struct {
	Program *string `json:"program"`
	Level   *string `json:"level"`
	Search  string  `json:"search"`
}

type LogEntry struct {
	Timestamp string  `json:"timestamp"`
	Level     string  `json:"level"`
	Program   string  `json:"program"`
	Message   string  `json:"message"`
	Location  *string `json:"location,omitempty"`
	Raw       any     `json:"raw"`
}
