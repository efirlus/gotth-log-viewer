package services

import (
	"bufio"
	"encoding/json"
	"gotthlogviewer/internal/types"
	"os"
	"strings"
	"sync"
	"time"
)

type LogService struct {
	filepath string
	cache    []types.LogEntry
	lastRead time.Time
	mu       sync.RWMutex
}

func NewLogService(filepath string) *LogService {
	return &LogService{
		filepath: filepath,
	}
}

// getStringValue tries multiple field names and returns the first non-empty value
func getStringValue(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := data[key]; ok {
			if str, ok := val.(string); ok && str != "" {
				return str
			}
		}
	}
	return keys[len(keys)-1] // Return last key as default value
}

// getStringPointerValue returns nil if no value found
func getStringPointerValue(data map[string]interface{}, keys ...string) *string {
	for _, key := range keys {
		if val, ok := data[key]; ok {
			if str, ok := val.(string); ok && str != "" {
				return &str
			}
		}
	}
	return nil
}

func parseLogLine(line []byte) types.LogEntry {
	var rawLog map[string]interface{}
	if err := json.Unmarshal(line, &rawLog); err != nil {
		// If not JSON, return basic log entry
		return types.LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Level:     "info",
			Program:   "unknown",
			Message:   string(line),
		}
	}

	// Try multiple field names for each field
	timestamp := getStringValue(rawLog,
		"time",
		"timestamp",
		"@timestamp",
		time.Now().Format(time.RFC3339),
	)

	level := strings.ToLower(getStringValue(rawLog,
		"loglevel",
		"level",
		"severity",
		"log_level",
		"info",
	))

	program := getStringValue(rawLog,
		"programname",
		"program",
		"service",
		"app",
		"application",
		"unknown",
	)

	message := getStringValue(rawLog,
		"message",
		"msg",
		"text",
		"logstring",
		"log",
		"content",
		"",
	)

	var location *string
	if loc := getStringPointerValue(rawLog,
		"location",
		"source",
		"file",
		"caller",
	); loc != nil {
		location = loc
	}

	// Store the original log entry for debugging or detailed view
	originalJSON, _ := json.Marshal(rawLog)

	return types.LogEntry{
		Timestamp: timestamp,
		Level:     normalizeLogLevel(level),
		Program:   program,
		Message:   message,
		Location:  location,
		Raw:       string(originalJSON), // Keep original for debugging
	}
}

// normalizeLogLevel standardizes log levels
func normalizeLogLevel(level string) string {
	level = strings.ToLower(level)
	switch level {
	case "error", "err", "fatal", "panic":
		return "error"
	case "warn", "warning":
		return "warn"
	case "info", "information", "notice":
		return "info"
	case "debug", "trace":
		return "debug"
	default:
		return "info"
	}
}

func (ls *LogService) ReadLogs() ([]types.LogEntry, error) {
	ls.mu.RLock()

	if ls.cache != nil && time.Since(ls.lastRead) < 5*time.Second {
		logs := ls.cache
		ls.mu.RUnlock()
		return logs, nil
	}
	ls.mu.RUnlock()

	ls.mu.Lock()
	defer ls.mu.Unlock()

	// Double-check cache after getting write lock
	if ls.cache != nil && time.Since(ls.lastRead) < 5*time.Second {
		return ls.cache, nil
	}

	file, err := os.Open(ls.filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var logs []types.LogEntry
	scanner := bufio.NewScanner(file)
	// Increase buffer size for large lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		entry := parseLogLine(line)
		logs = append(logs, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	ls.cache = logs
	ls.lastRead = time.Now()

	return logs, nil
}

/*
type LogService interface {
	FetchLogs(filters model.LogFilters) ([]model.LogEntry, error)
	GetUniquePrograms() ([]string, error)
}

// Implementation will handle:
// - Reading log file
// - Parsing logs
// - Filtering
// - Timestamp tracking
type logService struct {
    logPath string
    cache   *logCache  // Optionally cache parsed logs
}

func (s *logService) FetchLogs(filters model.LogFilters) ([]model.LogEntry, error) {
    rawLogs := s.readLogFile()
    parsedLogs := s.parseLogs(rawLogs)
    return s.applyFilters(parsedLogs, filters), nil
}

var lastTimestamp *string = nil
var logFilePath = "/home/efirlus/goproject/Logs/app.log"
var logMutex sync.Mutex

func fetchLogs() ([]model.LogEntry, error) {
	baseURL := "/api/logs"
	var apiUrl string

	if lastTimestamp != nil {
		apiUrl = fmt.Sprintf("%s?since=%s", baseURL, url.QueryEscape(*lastTimestamp))
	} else {
		apiUrl = baseURL
	}

	resp, err := http.Get(apiUrl)
	if err != nil {
		return nil, fmt.Errorf("error fetching logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error! status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var data []map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	var logs []model.LogEntry
	for _, log := range data {
		timestamp := getStringValue(log, "time", "timestamp", time.Now().Format(time.RFC3339))
		level := strings.ToLower(getStringValue(log, "loglevel", "level", "severity", "info"))
		program := getStringValue(log, "programname", "program", "service", "unknown")
		message := getStringValue(log, "message", "msg", "text", "logstring")
		location := getStringPointerValue(log, "location")

		logEntry := model.LogEntry{
			Timestamp: timestamp,
			Level:     level,
			Program:   program,
			Message:   message,
			Location:  location,
			Raw:       log, // Store the original log entry for debugging
		}
		logs = append(logs, logEntry)
	}

	if len(logs) > 0 {
		lastTimestamp = &logs[len(logs)-1].Timestamp
	}

	return logs, nil
}

func resetLastTimestamp() {
	lastTimestamp = nil
}

func getStringValue(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := data[key]; ok {
			if strVal, valid := val.(string); valid {
				return strVal
			}
		}
	}
	return ""
}

func getStringPointerValue(data map[string]interface{}, keys ...string) *string {
	for _, key := range keys {
		if val, ok := data[key]; ok {
			if strVal, valid := val.(string); valid {
				return &strVal
			}
		}
	}
	return nil
}

// -------------  로그를 파일에서 읽는 부분 까지 --------------------- //

// SSEClient holds the channel for sending events
type SSEClient struct {
	channel chan []model.LogEntry
}

var clients = make(map[*SSEClient]bool)
var clientsMutex sync.Mutex

// addClient registers an SSE client
func addClient(client *SSEClient) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	clients[client] = true
}

// removeClient unregisters an SSE client
func removeClient(client *SSEClient) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	delete(clients, client)
	close(client.channel)
}

// broadcastLogs sends new log entries to all SSE clients
func broadcastLogs(entries []model.LogEntry) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	for client := range clients {
		client.channel <- entries
	}
}
*/
