package lg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Log levels
const (
	LevelInfo    = "info"
	LevelWarning = "warning"
	LevelError   = "error"
	LevelFatal   = "fatal"
	LevelPanic   = "panic"
	LevelDebug   = "debug"
	logFilePath  = "/home/efirlus/goproject/Logs/app.log"
)

// Logger struct to hold information like program name and logger instance
type Logger struct {
	programName string
	file        *os.File
	mu          sync.Mutex
}

var logger *Logger

// LogEntry structure to be written in the log file
type LogEntry struct {
	Time        string `json:"time"`
	Level       string `json:"loglevel"`
	ProgramName string `json:"programname"`
	Message     string `json:"logstring"`
	Location    string `json:"location,omitempty"`
}

// NewLogger initializes and returns a new Logger instance
func NewLogger(programName string) {
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Could not open log file:", err)
		return
	}
	logger = &Logger{
		programName: programName,
		file:        file, // No need for flags since we handle format
	}
}

// Info logs an informational message
func Info(message string) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	log(LevelInfo, message, "")
}

// Warn logs a warning message
func Warn(message string) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	log(LevelInfo, message, "")
}

// Err logs an error message along with the error object and stack trace
func Err(message string, err error) {
	logger.mu.Lock()
	defer logger.mu.Unlock()

	stackTrace := getStackTrace()
	fullMessage := fmt.Sprintf("%s: %v", message, err)
	log(LevelError, fullMessage, stackTrace)
}

// Fatal logs a fatal error and exits
func Fatal(message string, err error) {
	stackTrace := getStackTrace()
	fullMessage := fmt.Sprintf("%s: %v", message, err)
	log(LevelFatal, fullMessage, stackTrace)
	os.Exit(1)
}

// Panic logs a panic message
func Panic(message string, err error) {
	stackTrace := getStackTrace()
	fullMessage := fmt.Sprintf("%s: %v", message, err)
	log(LevelPanic, fullMessage, stackTrace)
	panic(fullMessage)
}

func Debug(elem ...any) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	var message strings.Builder
	message.WriteString("DEBUG: \n")
	for _, l := range elem {
		message.WriteString(fmt.Sprintf("%v\n", l))
	}
	log(LevelInfo, message.String(), "")
}

// log is a helper function to write logs in JSON format
func log(level, message, location string) {
	entry := LogEntry{
		Time:        time.Now().Format("2006-01-02 15:04:05.000"),
		Level:       level,
		ProgramName: logger.programName,
		Message:     message,
		Location:    location,
	}
	logJSON, err := json.Marshal(entry)
	if err != nil {
		fmt.Println("Could not marshal log entry:", err)
	}
	_, err = logger.file.WriteString(string(logJSON) + "\n")
	if err != nil {
		fmt.Println("Could not write log:", err)
	}
}

// getStackTrace retrieves the stack trace for error-level logs
func getStackTrace() string {
	stack := ""
	basePath := getBasePath()
	for i := 2; i < 10; i++ { // Skipping the first two calls (runtime, log function)
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		if !strings.HasPrefix(file, basePath) {
			continue
		}
		fn := runtime.FuncForPC(pc)
		stack += fmt.Sprintf("%s:%d %s\n", file, line, fn.Name())
	}
	return stack
}

// getBasePath retrieves the base directory path of the running program
func getBasePath() string {
	exePath, err := os.Executable() // Get the executable's path
	if err != nil {
		return ""
	}
	return filepath.Dir(exePath) // Get the directory where the executable is located
}
