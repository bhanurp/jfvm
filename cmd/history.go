package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bhanurp/jfvm/cmd/descriptions"
	"github.com/bhanurp/jfvm/cmd/utils"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
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

type VersionStats struct {
	Version   string
	Count     int
	FirstUsed time.Time
	LastUsed  time.Time
	TotalTime time.Duration
	Commands  map[string]int
}

var History = &cli.Command{
	Name:        "history",
	Usage:       descriptions.History.Usage,
	Description: descriptions.History.Format(),
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "limit",
			Usage: "Limit number of entries to show",
			Value: 50,
		},
		&cli.BoolFlag{
			Name:  "stats",
			Usage: "Show aggregated statistics",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "version",
			Usage: "Filter by specific version",
		},
		&cli.BoolFlag{
			Name:  "no-color",
			Usage: "Disable colored output",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "format",
			Usage: "Output format: table, json",
			Value: "table",
		},
		&cli.BoolFlag{
			Name:  "clear",
			Usage: "Clear history (cannot be undone)",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "show-output",
			Usage: "Show command output in history entries",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "command",
			Usage: "Filter by command pattern (case-insensitive)",
		},
		&cli.BoolFlag{
			Name:  "failures-only",
			Usage: "Show only failed commands (exit code != 0)",
			Value: false,
		},
	},
	Action: func(c *cli.Context) error {
		if c.Bool("clear") {
			return clearHistory()
		}

		historyFile := filepath.Join(utils.JfvmRoot, "history.json")

		entries, err := loadHistory(historyFile)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to load history: %w", err)
		}

		// Filter by version if specified
		if version := c.String("version"); version != "" {
			filtered := []HistoryEntry{}
			for _, entry := range entries {
				if entry.Version == version {
					filtered = append(filtered, entry)
				}
			}
			entries = filtered
		}

		// Filter by command pattern if specified
		if cmdPattern := c.String("command"); cmdPattern != "" {
			filtered := []HistoryEntry{}
			for _, entry := range entries {
				if strings.Contains(strings.ToLower(entry.Command), strings.ToLower(cmdPattern)) {
					filtered = append(filtered, entry)
				}
			}
			entries = filtered
		}

		// Filter failures only if specified
		if c.Bool("failures-only") {
			filtered := []HistoryEntry{}
			for _, entry := range entries {
				if entry.ExitCode != 0 {
					filtered = append(filtered, entry)
				}
			}
			entries = filtered
		}

		if len(entries) == 0 {
			fmt.Println("📭 No history entries found.")
			return nil
		}

		if c.Bool("stats") {
			displayHistoryStats(entries, c.Bool("no-color"))
		} else {
			displayHistory(entries, c.Int("limit"), c.String("format"), c.Bool("no-color"), c.Bool("show-output"))
		}

		return nil
	},
}

func loadHistory(historyFile string) ([]HistoryEntry, error) {
	data, err := os.ReadFile(historyFile)
	if err != nil {
		return nil, err
	}

	var entries []HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func saveHistory(historyFile string, entries []HistoryEntry) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(historyFile), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyFile, data, 0644)
}

func AddHistoryEntry(version, command string, duration time.Duration, exitCode int, stdout, stderr string) {
	historyFile := filepath.Join(utils.JfvmRoot, "history.json")

	entries, err := loadHistory(historyFile)
	if err != nil && !os.IsNotExist(err) {
		return
	}

	// Truncate output to prevent huge history files
	const maxOutputSize = 5000
	if len(stdout) > maxOutputSize {
		stdout = stdout[:maxOutputSize] + "\n... (truncated)"
	}
	if len(stderr) > maxOutputSize {
		stderr = stderr[:maxOutputSize] + "\n... (truncated)"
	}

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

	saveHistory(historyFile, entries)
}

func displayHistory(entries []HistoryEntry, limit int, format string, noColor, showOutput bool) {
	if noColor {
		color.NoColor = true
	}

	// Sort by timestamp (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	if limit > 0 && limit < len(entries) {
		entries = entries[:limit]
	}

	switch format {
	case "json":
		displayHistoryJSON(entries)
	default:
		displayHistoryTable(entries, showOutput)
	}
}

func displayHistoryTable(entries []HistoryEntry, showOutput bool) {
	var (
		blueColor   = color.New(color.FgBlue)
		greenColor  = color.New(color.FgGreen)
		yellowColor = color.New(color.FgYellow)
		redColor    = color.New(color.FgRed)
	)

	fmt.Printf("📊 JFVM USAGE HISTORY\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════════\n\n")

	if showOutput {
		fmt.Printf("%-20s %-15s %-12s %-8s %-30s\n", "TIMESTAMP", "VERSION", "DURATION", "EXIT", "COMMAND")
	} else {
		fmt.Printf("%-20s %-15s %-12s %-30s\n", "TIMESTAMP", "VERSION", "DURATION", "COMMAND")
	}
	fmt.Printf("─────────────────────────────────────────────────────────────────────────────────────\n")

	for _, entry := range entries {
		timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
		duration := ""
		if entry.Duration > 0 {
			duration = fmt.Sprintf("%.0fms", float64(entry.Duration))
		}

		command := entry.Command
		if !showOutput && len(command) > 28 {
			command = command[:25] + "..."
		}

		var durationColor *color.Color = greenColor
		if entry.Duration > 5000 {
			durationColor = yellowColor
		}

		exitCodeColor := greenColor
		exitCodeText := "0"
		if entry.ExitCode != 0 {
			exitCodeColor = redColor
			exitCodeText = fmt.Sprintf("%d", entry.ExitCode)
		}

		if showOutput {
			fmt.Printf("%-20s %-15s %-12s %-8s %-30s\n",
				blueColor.Sprint(timestamp),
				greenColor.Sprint(entry.Version),
				durationColor.Sprint(duration),
				exitCodeColor.Sprint(exitCodeText),
				command)

			if entry.Stdout != "" {
				fmt.Printf("  📤 STDOUT:\n%s\n", entry.Stdout)
			}
			if entry.Stderr != "" {
				fmt.Printf("  📥 STDERR:\n%s\n", redColor.Sprint(entry.Stderr))
			}
			if entry.Stdout != "" || entry.Stderr != "" {
				fmt.Println()
			}
		} else {
			fmt.Printf("%-20s %-15s %-12s %-30s\n",
				blueColor.Sprint(timestamp),
				greenColor.Sprint(entry.Version),
				durationColor.Sprint(duration),
				command)
		}
	}

	fmt.Printf("\n📈 Total entries: %d\n", len(entries))
}

func displayHistoryJSON(entries []HistoryEntry) {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func displayHistoryStats(entries []HistoryEntry, noColor bool) {
	if noColor {
		color.NoColor = true
	}

	var (
		greenColor  = color.New(color.FgGreen, color.Bold)
		blueColor   = color.New(color.FgBlue, color.Bold)
		yellowColor = color.New(color.FgYellow, color.Bold)
	)

	// Aggregate statistics
	stats := make(map[string]*VersionStats)
	totalCommands := make(map[string]int)

	for _, entry := range entries {
		if stats[entry.Version] == nil {
			stats[entry.Version] = &VersionStats{
				Version:   entry.Version,
				FirstUsed: entry.Timestamp,
				LastUsed:  entry.Timestamp,
				Commands:  make(map[string]int),
			}
		}

		s := stats[entry.Version]
		s.Count++
		s.TotalTime += time.Duration(entry.Duration) * time.Millisecond

		if entry.Timestamp.Before(s.FirstUsed) {
			s.FirstUsed = entry.Timestamp
		}
		if entry.Timestamp.After(s.LastUsed) {
			s.LastUsed = entry.Timestamp
		}

		if entry.Command != "" {
			s.Commands[entry.Command]++
			totalCommands[entry.Command]++
		}
	}

	fmt.Printf("📊 JFVM USAGE STATISTICS\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════════\n\n")

	// Version usage
	fmt.Printf("🔢 VERSION USAGE:\n")
	fmt.Printf("─────────────────────────────────────────────────────────────────────────────────────\n")

	// Sort versions by usage count
	versions := make([]*VersionStats, 0, len(stats))
	for _, stat := range stats {
		versions = append(versions, stat)
	}
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Count > versions[j].Count
	})

	fmt.Printf("%-15s %-8s %-12s %-20s %-20s\n", "VERSION", "COUNT", "TOTAL TIME", "FIRST USED", "LAST USED")
	fmt.Printf("─────────────────────────────────────────────────────────────────────────────────────\n")

	for i, stat := range versions {
		var versionColor *color.Color = blueColor
		if i == 0 {
			versionColor = greenColor
		}

		fmt.Printf("%-15s %-8s %-12s %-20s %-20s\n",
			versionColor.Sprint(stat.Version),
			yellowColor.Sprintf("%d", stat.Count),
			formatDuration(stat.TotalTime),
			stat.FirstUsed.Format("2006-01-02 15:04"),
			stat.LastUsed.Format("2006-01-02 15:04"))
	}

	// Most common commands
	if len(totalCommands) > 0 {
		fmt.Printf("\n🚀 MOST COMMON COMMANDS:\n")
		fmt.Printf("─────────────────────────────────────────────────────────────────────────────────────\n")

		// Sort commands by frequency
		type commandStat struct {
			command string
			count   int
		}
		commands := make([]commandStat, 0, len(totalCommands))
		for cmd, count := range totalCommands {
			commands = append(commands, commandStat{cmd, count})
		}
		sort.Slice(commands, func(i, j int) bool {
			return commands[i].count > commands[j].count
		})

		maxShow := 10
		if len(commands) < maxShow {
			maxShow = len(commands)
		}

		for i, cmd := range commands[:maxShow] {
			var color *color.Color = blueColor
			if i == 0 {
				color = greenColor
			}
			fmt.Printf("%-50s %s\n", cmd.command, color.Sprintf("(%d times)", cmd.count))
		}
	}

	// Timeline
	fmt.Printf("\n📅 USAGE TIMELINE:\n")
	fmt.Printf("─────────────────────────────────────────────────────────────────────────────────────\n")

	if len(entries) > 0 {
		oldest := entries[0].Timestamp
		newest := entries[0].Timestamp
		for _, entry := range entries {
			if entry.Timestamp.Before(oldest) {
				oldest = entry.Timestamp
			}
			if entry.Timestamp.After(newest) {
				newest = entry.Timestamp
			}
		}

		duration := newest.Sub(oldest)
		avgPerDay := float64(len(entries)) / (duration.Hours() / 24)

		fmt.Printf("First usage: %s\n", greenColor.Sprint(oldest.Format("2006-01-02 15:04:05")))
		fmt.Printf("Latest usage: %s\n", greenColor.Sprint(newest.Format("2006-01-02 15:04:05")))
		fmt.Printf("Total period: %s\n", yellowColor.Sprint(formatDuration(duration)))
		fmt.Printf("Total entries: %s\n", yellowColor.Sprintf("%d", len(entries)))
		if duration.Hours() > 24 {
			fmt.Printf("Average per day: %s\n", yellowColor.Sprintf("%.1f", avgPerDay))
		}
	}
}

func clearHistory() error {
	historyFile := filepath.Join(utils.JfvmRoot, "history.json")

	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		fmt.Println("📭 No history file found.")
		return nil
	}

	if err := os.Remove(historyFile); err != nil {
		return fmt.Errorf("failed to clear history: %w", err)
	}

	fmt.Println("🗑️  History cleared successfully.")
	return nil
}
