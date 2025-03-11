package util

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// LogLevel specifies the log level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// LogWriter is an interface for writing logs
type LogWriter interface {
	Write(message string) error
}

// StderrLogger is a logger that writes to stderr
type StderrLogger struct{}

// Write implements LogWriter
func (l StderrLogger) Write(message string) error {
	_, err := fmt.Fprintln(os.Stderr, message)
	return err
}

// FileLogger is a logger that writes to a file
type FileLogger struct {
	File *os.File
}

// Write implements LogWriter
func (l FileLogger) Write(message string) error {
	_, err := fmt.Fprintln(l.File, message)
	return err
}

// Logger is a simple logger
type Logger struct {
	writer    LogWriter
	logLevel  LogLevel
	showLevel bool
	showTime  bool
}

// NewLogger creates a new logger
func NewLogger(writer LogWriter) *Logger {
	return &Logger{
		writer:    writer,
		logLevel:  LogLevelInfo, // Default log level
		showLevel: true,
		showTime:  true,
	}
}

// SetLogLevel sets the log level
func (l *Logger) SetLogLevel(level LogLevel) {
	l.logLevel = level
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.logLevel <= LogLevelDebug {
		l.log("DEBUG", format, args...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.logLevel <= LogLevelInfo {
		l.log("INFO", format, args...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.logLevel <= LogLevelWarn {
		l.log("WARN", format, args...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.logLevel <= LogLevelError {
		l.log("ERROR", format, args...)
	}
}

// log formats and writes a log message
func (l *Logger) log(level, format string, args ...interface{}) {
	var message string

	// Add timestamp
	if l.showTime {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		message = fmt.Sprintf("[%s] ", timestamp)
	}

	// Add log level
	if l.showLevel {
		message += fmt.Sprintf("[%s] ", level)
	}

	// Add message
	message += fmt.Sprintf(format, args...)

	// Write to the log
	if err := l.writer.Write(message); err != nil {
		fmt.Fprintf(os.Stderr, "Logger error: %v\n", err)
	}
}

// ParseInt parses a string into an integer
func ParseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// StdioReadWriteCloser is a ReadWriteCloser for stdin/stdout
type StdioReadWriteCloser struct{}

// Read implements io.Reader
func (StdioReadWriteCloser) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

// Write implements io.Writer
func (StdioReadWriteCloser) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

// Close implements io.Closer
func (StdioReadWriteCloser) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}

// NewStdioReadWriteCloser creates a new StdioReadWriteCloser
func NewStdioReadWriteCloser() *StdioReadWriteCloser {
	return &StdioReadWriteCloser{}
}
