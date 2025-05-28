package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type HistoryEntry struct {
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Command   string    `json:"command,omitempty"`
	Duration  int64     `json:"duration_ms,omitempty"`
	ExitCode  int       `json:"exit_code,omitempty"`
	Stdout    string    `json:"stdout,omitempty"`
	Stderr    string    `json:"stderr,omitempty"`
}

func main() {
	home := os.Getenv("HOME")
	configPath := filepath.Join(home, ".jfvm", "config")
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "No current version set. Run `jfvm use <version>` first.\n")
		os.Exit(1)
	}

	version := strings.TrimSpace(string(data))
	bin := filepath.Join(home, ".jfvm", "versions", version, "jf")

	// Only print debug info if JFVM_DEBUG is set
	if os.Getenv("JFVM_DEBUG") != "" {
		fmt.Printf("[shim] Executing version: %s\n", version)
		fmt.Printf("[shim] Full binary path: %s\n", bin)
	}

	// Record start time for history
	startTime := time.Now()
	command := strings.Join(os.Args[1:], " ")

	// Create buffers to capture output
	var stdout, stderr bytes.Buffer

	cmd := exec.Command(bin, os.Args[1:]...)
	cmd.Stdin = os.Stdin

	// Capture output while also writing to the original streams
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	duration := time.Since(startTime)

	// Write captured output to original streams
	os.Stdout.Write(stdout.Bytes())
	os.Stderr.Write(stderr.Bytes())

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	// Record history entry (silently fail if there's an issue)
	addHistoryEntry(home, version, command, duration, exitCode, stdout.String(), stderr.String())

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "[shim] Error executing binary: %v\n", err)
		os.Exit(1)
	}
}

func addHistoryEntry(home, version, command string, duration time.Duration, exitCode int, stdout, stderr string) {
	historyFile := filepath.Join(home, ".jfvm", "history.json")

	// Load existing history
	var entries []HistoryEntry
	if data, err := os.ReadFile(historyFile); err == nil {
		// Ignore errors, start fresh if corrupted
		json.Unmarshal(data, &entries)
	}

	// Truncate output to prevent huge history files
	const maxOutputSize = 5000
	if len(stdout) > maxOutputSize {
		stdout = stdout[:maxOutputSize] + "\n... (truncated)"
	}
	if len(stderr) > maxOutputSize {
		stderr = stderr[:maxOutputSize] + "\n... (truncated)"
	}

	// Add new entry
	entry := HistoryEntry{
		Version:   version,
		Timestamp: time.Now(),
		Command:   command,
		Duration:  duration.Milliseconds(),
		ExitCode:  exitCode,
		Stdout:    stdout,
		Stderr:    stderr,
	}
	entries = append(entries, entry)

	// Keep only last 1000 entries to prevent unlimited growth
	if len(entries) > 1000 {
		entries = entries[len(entries)-1000:]
	}

	// Save back (silently fail on errors to avoid disrupting normal operation)
	if data, err := json.MarshalIndent(entries, "", "  "); err == nil {
		os.MkdirAll(filepath.Dir(historyFile), 0755)
		os.WriteFile(historyFile, data, 0644)
	}
}
