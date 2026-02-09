// Package logger provides a simple wrapper around slog for the toolbox CLI.
package logger

import (
	"io"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger with additional convenience methods.
type Logger struct {
	*slog.Logger
	verbose bool
}

// New creates a new Logger with the given verbosity level.
func New(verbose bool) *Logger {
	var level slog.Level
	if verbose {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	return &Logger{
		Logger:  slog.New(handler),
		verbose: verbose,
	}
}

// NewWithWriter creates a new Logger that writes to the given io.Writer.
func NewWithWriter(w io.Writer, verbose bool) *Logger {
	var level slog.Level
	if verbose {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(w, opts)
	return &Logger{
		Logger:  slog.New(handler),
		verbose: verbose,
	}
}

// IsVerbose returns whether verbose logging is enabled.
func (l *Logger) IsVerbose() bool {
	return l.verbose
}

// Success logs a success message at Info level.
func (l *Logger) Success(msg string, args ...any) {
	l.Info(msg, args...)
}

// Warning logs a warning message at Warn level.
func (l *Logger) Warning(msg string, args ...any) {
	l.Warn(msg, args...)
}

// Verbose logs a message only if verbose mode is enabled.
func (l *Logger) Verbose(msg string, args ...any) {
	if l.verbose {
		l.Debug(msg, args...)
	}
}
