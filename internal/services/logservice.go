package services

import (
	"bufio"
	"encoding/json"
	"fmt"
	lg "gotthlogviewer/cmd/logger"
	"gotthlogviewer/internal/types"
	"os"
	"strings"
	"sync"
	"time"
)

type LogService struct {
	filepath    string
	cache       []types.LogEntry
	lastRead    time.Time
	lastModTime time.Time
	mu          sync.RWMutex
	onChange    func([]types.LogEntry)
}

func NewLogService(filepath string) *LogService {
	ls := &LogService{
		filepath: filepath,
	}

	go ls.watch()

	return ls
}

func (ls *LogService) watch() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	ls.lastModTime = time.Now() // Initialize with current time

	for range ticker.C {
		lg.Debug("checking for log changes")

		stat, err := os.Stat(ls.filepath)
		if err != nil {
			lg.Error("failed to stat log file", err)
			continue
		}

		// Check if file has been modified
		if stat.ModTime().After(ls.lastModTime) {
			lg.Debug(fmt.Sprintln("detected log file changes ",
				"lastMod ", ls.lastModTime,
				"newMod ", stat.ModTime()))

			// Update timestamp first to avoid missing updates
			ls.lastModTime = stat.ModTime()

			// Read new logs
			logs, err := ls.ReadLogs()
			if err != nil {
				lg.Error("failed to read logs", err)
				continue
			}

			if ls.onChange != nil {
				lg.Debug(fmt.Sprintln("broadcasting log updates ", "logCount:", len(logs)))
				ls.onChange(logs)
			}
		}
	}
}

func (ls *LogService) SetOnChange(fn func([]types.LogEntry)) {
	ls.onChange = fn
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
		lg.Debug(fmt.Sprintln("returning cached logs ", "count:", len(logs)))
		return logs, nil
	}
	ls.mu.RUnlock()

	ls.mu.Lock()
	defer ls.mu.Unlock()

	// Double-check cache after getting write lock
	if ls.cache != nil && time.Since(ls.lastRead) < 5*time.Second {
		lg.Debug(fmt.Sprintln("returning cached logs after double-check ", "count:", len(ls.cache)))
		return ls.cache, nil
	}

	lg.Debug("reading logs from file")
	file, err := os.Open(ls.filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
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
		return nil, fmt.Errorf("error scanning log file: %w", err)
	}

	ls.cache = logs
	ls.lastRead = time.Now()

	lg.Debug(fmt.Sprintln("read new logs from file ", "count:", len(logs)))
	return logs, nil
}
