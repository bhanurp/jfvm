package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/bhanurp/jfvm/cmd/descriptions"
	"github.com/bhanurp/jfvm/cmd/utils"
	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

type ExecutionResult struct {
	Version   string
	Command   string
	Output    string
	ErrorMsg  string
	ExitCode  int
	Duration  time.Duration
	StartTime time.Time
}

var Compare = &cli.Command{
	Name:        "compare",
	Usage:       descriptions.Compare.Usage,
	ArgsUsage:   "<version1> <version2> -- <jf-command> [args...]",
	Description: descriptions.Compare.Format(),
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "unified",
			Usage: "Show unified diff format instead of side-by-side",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "no-color",
			Usage: "Disable colored output",
			Value: false,
		},
		&cli.IntFlag{
			Name:  "timeout",
			Usage: "Command timeout in seconds",
			Value: 30,
		},
		&cli.BoolFlag{
			Name:  "timing",
			Usage: "Show execution timing information",
			Value: true,
		},
	},
	Action: func(c *cli.Context) error {
		args := c.Args().Slice()
		if len(args) < 3 {
			return cli.Exit("Usage: jfvm compare <version1> <version2> -- <jf-command> [args...]", 1)
		}

		// Find the separator "--"
		separatorIndex := -1
		for i, arg := range args {
			if arg == "--" {
				separatorIndex = i
				break
			}
		}

		if separatorIndex == -1 || separatorIndex != 2 {
			return cli.Exit("Missing '--' separator. Usage: jfvm compare <version1> <version2> -- <jf-command> [args...]", 1)
		}

		version1 := args[0]
		version2 := args[1]
		jfCommand := args[3:]

		if len(jfCommand) == 0 {
			return cli.Exit("No JFrog CLI command specified after '--'", 1)
		}

		// Resolve aliases if needed
		resolved1, err := utils.ResolveVersionOrAlias(version1)
		if err != nil {
			resolved1 = version1
		}
		resolved2, err := utils.ResolveVersionOrAlias(version2)
		if err != nil {
			resolved2 = version2
		}

		// Check if versions exist
		if err := utils.CheckVersionExists(resolved1); err != nil {
			return fmt.Errorf("version %s (%s) not found: %w", version1, resolved1, err)
		}
		if err := utils.CheckVersionExists(resolved2); err != nil {
			return fmt.Errorf("version %s (%s) not found: %w", version2, resolved2, err)
		}

		fmt.Printf("🔄 Comparing JFrog CLI versions: %s vs %s\n", version1, version2)
		fmt.Printf("📝 Command: jf %s\n\n", strings.Join(jfCommand, " "))

		// Execute commands in parallel
		results := make([]ExecutionResult, 2)
		g, ctx := errgroup.WithContext(context.Background())

		timeout := time.Duration(c.Int("timeout")) * time.Second
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		g.Go(func() error {
			result, err := executeJFCommand(timeoutCtx, resolved1, jfCommand)
			results[0] = result
			return err
		})

		g.Go(func() error {
			result, err := executeJFCommand(timeoutCtx, resolved2, jfCommand)
			results[1] = result
			return err
		})

		if err := g.Wait(); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: %v\n\n", err)
		}

		// Display results
		displayComparison(results[0], results[1], c.Bool("unified"), c.Bool("no-color"), c.Bool("timing"))

		return nil
	},
}

func executeJFCommand(ctx context.Context, version string, jfCommand []string) (ExecutionResult, error) {
	result := ExecutionResult{
		Version:   version,
		Command:   strings.Join(jfCommand, " "),
		StartTime: time.Now(),
	}

	binPath := filepath.Join(utils.JfvmVersions, version, utils.BinaryName)

	cmd := exec.CommandContext(ctx, binPath, jfCommand...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.Duration = time.Since(result.StartTime)

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = 1
		}
		result.ErrorMsg = stderr.String()
	} else {
		// When command succeeds, combine stdout and stderr for output comparison
		// Many CLI tools write informational messages to stderr even on success
		stdoutStr := stdout.String()
		stderrStr := stderr.String()

		if stdoutStr != "" && stderrStr != "" {
			result.Output = stdoutStr + "\n" + stderrStr
		} else if stdoutStr != "" {
			result.Output = stdoutStr
		} else {
			result.Output = stderrStr
		}
	}

	// Keep original stdout for cases where only stdout is needed
	if result.ExitCode == 0 && result.Output == "" {
		result.Output = stdout.String()
	}

	return result, nil
}

func displayComparison(result1, result2 ExecutionResult, unified, noColor, showTiming bool) {
	// Setup colors
	var (
		redColor   = color.New(color.FgRed)
		greenColor = color.New(color.FgGreen)
		blueColor  = color.New(color.FgBlue)
	)

	if noColor {
		color.NoColor = true
	}

	// Display headers
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("🔍 COMPARISON RESULTS\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════════\n\n")

	// Display timing information
	if showTiming {
		fmt.Printf("⏱️  EXECUTION TIMING:\n")
		fmt.Printf("   Version %s: %v\n", blueColor.Sprint(result1.Version), result1.Duration)
		fmt.Printf("   Version %s: %v\n", blueColor.Sprint(result2.Version), result2.Duration)
		fmt.Printf("\n")
	}

	// Display exit codes if different
	if result1.ExitCode != result2.ExitCode {
		fmt.Printf("🚨 EXIT CODE DIFFERENCE:\n")
		if result1.ExitCode == 0 {
			fmt.Printf("   %s: %s\n", result1.Version, greenColor.Sprint("✓ 0"))
		} else {
			fmt.Printf("   %s: %s\n", result1.Version, redColor.Sprintf("✗ %d", result1.ExitCode))
		}
		if result2.ExitCode == 0 {
			fmt.Printf("   %s: %s\n", result2.Version, greenColor.Sprint("✓ 0"))
		} else {
			fmt.Printf("   %s: %s\n", result2.Version, redColor.Sprintf("✗ %d", result2.ExitCode))
		}
		fmt.Printf("\n")
	}

	// Display errors if any
	if result1.ErrorMsg != "" || result2.ErrorMsg != "" {
		fmt.Printf("🚨 ERROR OUTPUT:\n")
		if result1.ErrorMsg != "" {
			fmt.Printf("   %s ERROR:\n%s\n", redColor.Sprint(result1.Version), result1.ErrorMsg)
		}
		if result2.ErrorMsg != "" {
			fmt.Printf("   %s ERROR:\n%s\n", redColor.Sprint(result2.Version), result2.ErrorMsg)
		}
		fmt.Printf("\n")
	}

	// Compare outputs
	output1 := strings.TrimSpace(result1.Output)
	output2 := strings.TrimSpace(result2.Output)

	// Commands with different exit codes should never be considered identical
	// Even if their stdout happens to be the same, they represent different execution results
	if output1 == output2 && result1.ExitCode == result2.ExitCode && result1.ErrorMsg == result2.ErrorMsg {
		fmt.Printf("✅ OUTPUTS ARE IDENTICAL\n")
		fmt.Printf("📄 Output (%d lines):\n", len(strings.Split(output1, "\n")))
		fmt.Printf("─────────────────────────────────────────────────────────────────────────────────────\n")
		fmt.Printf("%s\n", output1)
		return
	}

	fmt.Printf("📊 OUTPUT DIFFERENCES:\n")

	if unified {
		displayUnifiedDiff(output1, output2, result1.Version, result2.Version, noColor)
	} else {
		displaySideBySideDiff(output1, output2, result1.Version, result2.Version, noColor)
	}
}

func displayUnifiedDiff(output1, output2, version1, version2 string, noColor bool) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(output1, output2, false)

	var (
		redColor   = color.New(color.FgRed)
		greenColor = color.New(color.FgGreen)
	)

	fmt.Printf("─────────────────────────────────────────────────────────────────────────────────────\n")
	fmt.Printf("%s %s\n", redColor.Sprint("---"), version1)
	fmt.Printf("%s %s\n", greenColor.Sprint("+++"), version2)
	fmt.Printf("─────────────────────────────────────────────────────────────────────────────────────\n")

	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffDelete:
			if noColor {
				fmt.Printf("- %s", diff.Text)
			} else {
				redColor.Printf("- %s", diff.Text)
			}
		case diffmatchpatch.DiffInsert:
			if noColor {
				fmt.Printf("+ %s", diff.Text)
			} else {
				greenColor.Printf("+ %s", diff.Text)
			}
		case diffmatchpatch.DiffEqual:
			fmt.Printf("  %s", diff.Text)
		}
	}
}

func displaySideBySideDiff(output1, output2, version1, version2 string, noColor bool) {
	lines1 := strings.Split(output1, "\n")
	lines2 := strings.Split(output2, "\n")

	maxLines := len(lines1)
	if len(lines2) > maxLines {
		maxLines = len(lines2)
	}

	var (
		blueColor  = color.New(color.FgBlue)
		redColor   = color.New(color.FgRed, color.Bold)
		greenColor = color.New(color.FgGreen, color.Bold)
	)

	// Header
	fmt.Printf("─────────────────────────────────────────────────────────────────────────────────────\n")
	fmt.Printf("%-40s │ %-40s\n", blueColor.Sprintf("%s", version1), blueColor.Sprintf("%s", version2))
	fmt.Printf("─────────────────────────────────────────────────────────────────────────────────────\n")

	for i := 0; i < maxLines; i++ {
		line1 := ""
		line2 := ""

		if i < len(lines1) {
			line1 = lines1[i]
		}
		if i < len(lines2) {
			line2 = lines2[i]
		}

		// Truncate long lines for display
		if len(line1) > 38 {
			line1 = line1[:35] + "..."
		}
		if len(line2) > 38 {
			line2 = line2[:35] + "..."
		}

		marker1 := " "
		marker2 := " "

		if line1 != line2 {
			if line1 != "" && line2 == "" {
				marker1 = "-"
				if !noColor {
					line1 = redColor.Sprint(line1)
				}
			} else if line1 == "" && line2 != "" {
				marker2 = "+"
				if !noColor {
					line2 = greenColor.Sprint(line2)
				}
			} else {
				marker1 = "~"
				marker2 = "~"
				if !noColor {
					line1 = redColor.Sprint(line1)
					line2 = greenColor.Sprint(line2)
				}
			}
		}

		fmt.Printf("%s%-39s │ %s%-39s\n", marker1, line1, marker2, line2)
	}
}
