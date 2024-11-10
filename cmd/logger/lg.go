package lg

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
)

var (
	logger      *slog.Logger
	isDebugOn   bool
	debugFlag   = flag.Bool("debug", false, "Enable debug logging")
	programName string // Will store the module name
	timeFormat  = "2006-01-02 15:04:05.000"
)

// Initialize sets up the logger with the given log file path
func Initialize(logPath string, progName string) error {
	flag.Parse()
	isDebugOn = *debugFlag

	programName = progName

	// Create or open log file
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// Create JSON handler with options
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(time.Now().Format(timeFormat))
			}
			return a
		},
	}

	if isDebugOn {
		opts.Level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(file, opts)
	logger = slog.New(handler.WithAttrs([]slog.Attr{
		slog.String("program", programName),
	}))
	slog.SetDefault(logger)

	return nil
}

// getCallerLocation returns the file:line location of the caller
func getCallerLocation() string {
	_, file, line, _ := runtime.Caller(2) // Skip 2 frames to get the actual caller
	return fmt.Sprintf("%s:%d", file, line)
}

// Info logs an info level message
func Info(message string) {
	logger.Info(message)
}

// Warn logs a warning level message with error details but without location
func Warn(message string, err error) {
	logger.Warn(message,
		"error", err,
	)
}

// Error logs an error level message with error details and stack trace
func Error(message string, err error) {
	logger.Error(message,
		"error", err,
		"location", getCallerLocation(),
	)
}

// Debug logs a debug message if debug mode is enabled
func Debug(message string) {
	if isDebugOn {
		logger.Debug(message,
			"location", getCallerLocation(),
		)
	}
}

// Fatal logs a fatal error and exits the program
func Fatal(message string, err error) {
	logger.Error(message,
		"error", err,
		"location", getCallerLocation(),
		"level", "FATAL",
	)
	os.Exit(1)
}
