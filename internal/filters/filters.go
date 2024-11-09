package filters

import (
	"gotthlogviewer/internal/types"
	"slices"
	"strings"
)

// GetUniquePrograms returns a sorted slice of unique program names from logs
func GetUniquePrograms(logs []types.LogEntry) []string {
	programMap := make(map[string]bool)
	for _, log := range logs {
		if log.Program != "" {
			programMap[log.Program] = true
		}
	}

	programs := make([]string, 0, len(programMap))
	for program := range programMap {
		programs = append(programs, program)
	}
	slices.Sort(programs)
	return programs
}

// SortLogs sorts log entries by timestamp in descending order (newest first)
func SortLogs(logs []types.LogEntry) []types.LogEntry {
	sorted := make([]types.LogEntry, len(logs))
	copy(sorted, logs)

	slices.SortFunc(sorted, func(a, b types.LogEntry) int {
		switch {
		case a.Timestamp > b.Timestamp:
			return -1
		case a.Timestamp < b.Timestamp:
			return 1
		default:
			return 0
		}
	})

	return sorted
}

// ApplyFilters filters log entries based on the provided filters
func ApplyFilters(logs []types.LogEntry, filters types.LogFilters) []types.LogEntry {
	if !hasActiveFilters(filters) {
		return logs
	}

	filtered := slices.Clone(logs)

	if filters.Search != "" {
		search := strings.ToLower(filters.Search)
		filtered = slices.DeleteFunc(filtered, func(log types.LogEntry) bool {
			return !matchesSearch(log, search)
		})
	}

	if filters.Level != "" {
		filtered = slices.DeleteFunc(filtered, func(log types.LogEntry) bool {
			return log.Level != filters.Level
		})
	}

	if filters.Program != "" {
		filtered = slices.DeleteFunc(filtered, func(log types.LogEntry) bool {
			return log.Program != filters.Program
		})
	}

	return filtered
}

// hasActiveFilters checks if any filters are active
func hasActiveFilters(filters types.LogFilters) bool {
	return filters.Search != "" || filters.Level != "" || filters.Program != ""
}

// matchesSearch checks if a log entry matches the search term
func matchesSearch(log types.LogEntry, search string) bool {
	return strings.Contains(strings.ToLower(log.Message), search) ||
		strings.Contains(strings.ToLower(log.Program), search) ||
		strings.Contains(strings.ToLower(log.Level), search)
}
