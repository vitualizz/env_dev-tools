package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// InstallLogger writes installation logs to a file and provides
// the log path for user inspection.
type InstallLogger struct {
	file    *os.File
	mu      sync.Mutex
	logPath string
}

// NewInstallLogger creates a logger that writes to ~/.vitualizz-devstack/install.log.
func NewInstallLogger() (*InstallLogger, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	logDir := filepath.Join(home, ".vitualizz-devstack")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}

	logPath := filepath.Join(logDir, "install.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	// Write session header
	fmt.Fprintf(f, "\n=== Session started: %s ===\n", time.Now().Format("2006-01-02 15:04:05"))

	return &InstallLogger{file: f, logPath: logPath}, nil
}

// LogPath returns the path to the log file.
func (l *InstallLogger) LogPath() string {
	return l.logPath
}

// LogCommand logs a command before execution.
func (l *InstallLogger) LogCommand(toolName, command string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.file, "[%s] CMD [%s]: %s\n", timestamp(), toolName, command)
	_ = l.file.Sync()
}

// LogSuccess logs a successful command with duration.
func (l *InstallLogger) LogSuccess(toolName string, duration time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.file, "[%s] OK   [%s] (%s)\n", timestamp(), toolName, duration)
	_ = l.file.Sync()
}

// LogError logs a failed command with error and output.
func (l *InstallLogger) LogError(toolName, command string, err error, output string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.file, "[%s] ERR  [%s]\n", timestamp(), toolName)
	fmt.Fprintf(l.file, "       Command: %s\n", command)
	fmt.Fprintf(l.file, "       Error:   %v\n", err)
	if output != "" {
		// Truncate output if too long
		out := truncate(output, 2000)
		fmt.Fprintf(l.file, "       Output:\n%s\n", indent(out, "       "))
	}
	fmt.Fprintln(l.file)
	_ = l.file.Sync()
}

// LogInfo logs a general info message.
func (l *InstallLogger) LogInfo(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.file, "[%s] INFO: %s\n", timestamp(), msg)
	_ = l.file.Sync()
}

// Close closes the log file.
func (l *InstallLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func timestamp() string {
	return time.Now().Format("15:04:05")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "... [truncated]"
}

func indent(s, prefix string) string {
	return prefix + strings.ReplaceAll(s, "\n", "\n"+prefix)
}
