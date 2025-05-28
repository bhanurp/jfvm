package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bhanurp/jfvm/cmd/descriptions"
	"github.com/bhanurp/jfvm/cmd/utils"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

type BenchmarkResult struct {
	Version     string
	Iterations  int
	TotalTime   time.Duration
	AverageTime time.Duration
	MinTime     time.Duration
	MaxTime     time.Duration
	SuccessRate float64
	Executions  []ExecutionResult
}

var Benchmark = &cli.Command{
	Name:        "benchmark",
	Usage:       descriptions.Benchmark.Usage,
	ArgsUsage:   "<version1,version2,...> -- <jf-command> [args...]",
	Description: descriptions.Benchmark.Format(),
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "iterations",
			Usage: "Number of iterations per version",
			Value: 5,
		},
		&cli.IntFlag{
			Name:  "timeout",
			Usage: "Command timeout in seconds",
			Value: 30,
		},
		&cli.BoolFlag{
			Name:  "no-color",
			Usage: "Disable colored output",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "detailed",
			Usage: "Show detailed execution logs",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "format",
			Usage: "Output format: table, json, csv",
			Value: "table",
		},
	},
	Action: func(c *cli.Context) error {
		// Parse and validate arguments
		versions, jfCommand, err := parseArguments(c.Args().Slice())
		if err != nil {
			return err
		}

		// Validate versions exist
		resolvedVersions, err := validateVersions(versions)
		if err != nil {
			return err
		}

		// Extract configuration
		config := extractBenchmarkConfig(c)

		// Run benchmarks
		results, err := runBenchmarks(resolvedVersions, jfCommand, config)
		if err != nil && config.Format == "table" {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: %v\n\n", err)
		}

		// Display results
		displayBenchmarkResults(results, config.Format, config.NoColor, config.Detailed)

		return nil
	},
}

type BenchmarkConfig struct {
	Iterations int
	Timeout    time.Duration
	Format     string
	NoColor    bool
	Detailed   bool
}

func parseArguments(args []string) (versions []string, jfCommand []string, err error) {
	if len(args) < 2 {
		return nil, nil, cli.Exit("Usage: jfvm benchmark <version1,version2,...> -- <jf-command> [args...]", 1)
	}

	// Find the separator "--"
	separatorIndex := -1
	for i, arg := range args {
		if arg == "--" {
			separatorIndex = i
			break
		}
	}

	if separatorIndex == -1 {
		return nil, nil, cli.Exit("Missing '--' separator. Usage: jfvm benchmark <versions> -- <jf-command> [args...]", 1)
	}

	if separatorIndex == 0 {
		return nil, nil, cli.Exit("No versions specified. Usage: jfvm benchmark <versions> -- <jf-command> [args...]", 1)
	}

	versionsStr := args[0]
	jfCommand = args[separatorIndex+1:]
	versions = strings.Split(versionsStr, ",")

	if len(jfCommand) == 0 {
		return nil, nil, cli.Exit("No JFrog CLI command specified after '--'", 1)
	}

	return versions, jfCommand, nil
}

func validateVersions(versions []string) ([]string, error) {
	resolvedVersions := make([]string, len(versions))
	for i, version := range versions {
		version = strings.TrimSpace(version)
		resolved, err := utils.ResolveVersionOrAlias(version)
		if err != nil {
			resolved = version
		}
		if err := utils.CheckVersionExists(resolved); err != nil {
			return nil, fmt.Errorf("version %s (%s) not found: %w", version, resolved, err)
		}
		resolvedVersions[i] = resolved
	}
	return resolvedVersions, nil
}

func extractBenchmarkConfig(c *cli.Context) BenchmarkConfig {
	return BenchmarkConfig{
		Iterations: c.Int("iterations"),
		Timeout:    time.Duration(c.Int("timeout")) * time.Second,
		Format:     c.String("format"),
		NoColor:    c.Bool("no-color"),
		Detailed:   c.Bool("detailed"),
	}
}

func runBenchmarks(versions []string, jfCommand []string, config BenchmarkConfig) ([]BenchmarkResult, error) {
	// Only show headers for table format
	if config.Format == "table" {
		fmt.Printf("🏁 Benchmarking JFrog CLI versions: %s\n", strings.Join(versions, ", "))
		fmt.Printf("📝 Command: jf %s\n", strings.Join(jfCommand, " "))
		fmt.Printf("🔄 Iterations: %d per version\n\n", config.Iterations)
	}

	// Run benchmarks
	results := make([]BenchmarkResult, len(versions))
	g, ctx := errgroup.WithContext(context.Background())

	for i, version := range versions {
		i, version := i, version
		g.Go(func() error {
			result, err := runBenchmark(ctx, version, jfCommand, config.Iterations, config.Timeout)
			results[i] = result
			return err
		})
	}

	return results, g.Wait()
}

func runBenchmark(ctx context.Context, version string, jfCommand []string, iterations int, timeout time.Duration) (BenchmarkResult, error) {
	result := BenchmarkResult{
		Version:    version,
		Iterations: iterations,
		MinTime:    time.Hour,
		Executions: make([]ExecutionResult, iterations),
	}

	successCount := 0

	for i := 0; i < iterations; i++ {
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		exec, err := executeJFCommand(timeoutCtx, version, jfCommand)
		cancel()

		result.Executions[i] = exec
		result.TotalTime += exec.Duration

		if exec.ExitCode == 0 {
			successCount++
		}

		if exec.Duration < result.MinTime {
			result.MinTime = exec.Duration
		}
		if exec.Duration > result.MaxTime {
			result.MaxTime = exec.Duration
		}

		if err != nil {
			fmt.Printf("⚠️  Iteration %d for %s failed: %v\n", i+1, version, err)
		}
	}

	result.AverageTime = result.TotalTime / time.Duration(iterations)
	result.SuccessRate = float64(successCount) / float64(iterations) * 100

	return result, nil
}

func displayBenchmarkResults(results []BenchmarkResult, format string, noColor, detailed bool) {
	if noColor {
		color.NoColor = true
	}

	var (
		greenColor  = color.New(color.FgGreen, color.Bold)
		redColor    = color.New(color.FgRed, color.Bold)
		blueColor   = color.New(color.FgBlue, color.Bold)
		yellowColor = color.New(color.FgYellow, color.Bold)
	)

	switch format {
	case "json":
		displayBenchmarkJSON(results)
	case "csv":
		displayBenchmarkCSV(results)
	default:
		displayBenchmarkTable(results, greenColor, redColor, blueColor, yellowColor, detailed)
	}
}

func displayBenchmarkTable(results []BenchmarkResult, greenColor, redColor, blueColor, yellowColor *color.Color, detailed bool) {
	fmt.Printf("📊 BENCHMARK RESULTS\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════════\n\n")

	// Sort by average time
	sort.Slice(results, func(i, j int) bool {
		return results[i].AverageTime < results[j].AverageTime
	})

	fmt.Printf("%-15s %-12s %-12s %-12s %-12s %-10s\n",
		"VERSION", "AVG TIME", "MIN TIME", "MAX TIME", "TOTAL TIME", "SUCCESS")
	fmt.Printf("─────────────────────────────────────────────────────────────────────────────────────\n")

	fastest := results[0].AverageTime

	for i, result := range results {
		var versionColor *color.Color
		if i == 0 {
			versionColor = greenColor
		} else if i == len(results)-1 {
			versionColor = redColor
		} else {
			versionColor = blueColor
		}

		speedup := float64(result.AverageTime) / float64(fastest)
		successColor := greenColor
		if result.SuccessRate < 100 {
			successColor = yellowColor
		}
		if result.SuccessRate < 80 {
			successColor = redColor
		}

		fmt.Printf("%-15s %-12s %-12s %-12s %-12s %s\n",
			versionColor.Sprint(result.Version),
			formatDuration(result.AverageTime),
			formatDuration(result.MinTime),
			formatDuration(result.MaxTime),
			formatDuration(result.TotalTime),
			successColor.Sprintf("%.1f%%", result.SuccessRate))

		if i > 0 {
			fmt.Printf("%-15s %s\n", "", yellowColor.Sprintf("↳ %.2fx slower", speedup))
		}
	}

	fmt.Printf("\n🏆 Performance Summary:\n")
	fmt.Printf("   Fastest: %s (%s avg)\n",
		greenColor.Sprint(results[0].Version),
		formatDuration(results[0].AverageTime))

	if len(results) > 1 {
		slowest := results[len(results)-1]
		speedDiff := float64(slowest.AverageTime) / float64(fastest)
		fmt.Printf("   Slowest: %s (%s avg, %.2fx slower)\n",
			redColor.Sprint(slowest.Version),
			formatDuration(slowest.AverageTime),
			speedDiff)
	}

	if detailed {
		fmt.Printf("\n📝 Detailed Execution Log:\n")
		for _, result := range results {
			fmt.Printf("\n%s:\n", blueColor.Sprint(result.Version))
			for i, exec := range result.Executions {
				status := greenColor.Sprint("✓")
				if exec.ExitCode != 0 {
					status = redColor.Sprint("✗")
				}
				fmt.Printf("  #%d: %s %s", i+1, status, formatDuration(exec.Duration))
				if exec.ExitCode != 0 {
					fmt.Printf(" (exit %d)", exec.ExitCode)
				}
				fmt.Printf("\n")
			}
		}
	}
}

func displayBenchmarkJSON(results []BenchmarkResult) {
	fmt.Printf("{\n")
	fmt.Printf("  \"benchmark_results\": [\n")
	for i, result := range results {
		fmt.Printf("    {\n")
		fmt.Printf("      \"version\": \"%s\",\n", result.Version)
		fmt.Printf("      \"iterations\": %d,\n", result.Iterations)
		fmt.Printf("      \"total_time_ms\": %.2f,\n", float64(result.TotalTime.Nanoseconds())/1e6)
		fmt.Printf("      \"average_time_ms\": %.2f,\n", float64(result.AverageTime.Nanoseconds())/1e6)
		fmt.Printf("      \"min_time_ms\": %.2f,\n", float64(result.MinTime.Nanoseconds())/1e6)
		fmt.Printf("      \"max_time_ms\": %.2f,\n", float64(result.MaxTime.Nanoseconds())/1e6)
		fmt.Printf("      \"success_rate\": %.2f\n", result.SuccessRate)
		if i < len(results)-1 {
			fmt.Printf("    },\n")
		} else {
			fmt.Printf("    }\n")
		}
	}
	fmt.Printf("  ]\n")
	fmt.Printf("}\n")
}

func displayBenchmarkCSV(results []BenchmarkResult) {
	fmt.Printf("version,iterations,total_time_ms,average_time_ms,min_time_ms,max_time_ms,success_rate\n")
	for _, result := range results {
		fmt.Printf("%s,%d,%.2f,%.2f,%.2f,%.2f,%.2f\n",
			result.Version,
			result.Iterations,
			float64(result.TotalTime.Nanoseconds())/1e6,
			float64(result.AverageTime.Nanoseconds())/1e6,
			float64(result.MinTime.Nanoseconds())/1e6,
			float64(result.MaxTime.Nanoseconds())/1e6,
			result.SuccessRate)
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fμs", float64(d.Nanoseconds())/1000)
	} else if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1e6)
	} else {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}
