package descriptions

import "fmt"

type CommandDescription struct {
	Usage       string
	Description string
	Examples    []Example
}

type Example struct {
	Command     string
	Description string
}

func (cd CommandDescription) Format() string {
	result := cd.Description

	if len(cd.Examples) > 0 {
		result += "\n\nExamples:\n"
		for _, example := range cd.Examples {
			if example.Description != "" {
				result += fmt.Sprintf("  # %s\n", example.Description)
			}
			result += fmt.Sprintf("  %s\n", example.Command)
			if example.Description != "" {
				result += "\n"
			}
		}
	}

	return result
}

var Install = CommandDescription{
	Usage:       "Install a specific JFrog CLI version",
	Description: "Downloads and installs the specified version of JFrog CLI from JFrog's public release server.",
	Examples: []Example{
		{
			Command:     "jfvm install 2.74.0",
			Description: "Install JFrog CLI version 2.74.0",
		},
		{
			Command:     "jfvm install latest",
			Description: "Install the latest available version",
		},
	},
}

var Use = CommandDescription{
	Usage:       "Set a specific JFrog CLI version as active",
	Description: "Activates the given version or alias. If .jfrog-version exists in the current directory, that will be used if no argument is passed.",
	Examples: []Example{
		{
			Command:     "jfvm use 2.74.0",
			Description: "Switch to JFrog CLI version 2.74.0",
		},
		{
			Command:     "jfvm use prod",
			Description: "Switch to the version aliased as 'prod'",
		},
		{
			Command:     "jfvm use",
			Description: "Use version from .jfrog-version file",
		},
	},
}

var List = CommandDescription{
	Usage:       "List all installed JFrog CLI versions",
	Description: "Shows all installed versions and highlights the currently active one.",
	Examples: []Example{
		{
			Command:     "jfvm list",
			Description: "Show all installed versions",
		},
	},
}

var Remove = CommandDescription{
	Usage:       "Remove a specific JFrog CLI version",
	Description: "Removes a specific version of JFrog CLI from your system.",
	Examples: []Example{
		{
			Command:     "jfvm remove 2.72.1",
			Description: "Remove JFrog CLI version 2.72.1",
		},
		{
			Command:     "jfvm remove old-dev",
			Description: "Remove a linked version named 'old-dev'",
		},
	},
}

var Clear = CommandDescription{
	Usage:       "Remove all installed JFrog CLI versions",
	Description: "Removes all installed versions of JFrog CLI. This action cannot be undone.",
	Examples: []Example{
		{
			Command:     "jfvm clear",
			Description: "Remove all installed versions",
		},
	},
}

var Alias = CommandDescription{
	Usage:       "Create or manage version aliases",
	Description: "Defines an alias for a specific version, making it easier to reference commonly used versions.",
	Examples: []Example{
		{
			Command:     "jfvm alias dev 2.74.0",
			Description: "Create alias 'dev' pointing to version 2.74.0",
		},
		{
			Command:     "jfvm alias prod 2.73.0",
			Description: "Create alias 'prod' pointing to version 2.73.0",
		},
		{
			Command:     "jfvm alias staging latest",
			Description: "Create alias 'staging' pointing to latest version",
		},
	},
}

var Link = CommandDescription{
	Usage:       "Link a locally built JFrog CLI binary",
	Description: "Links a locally built jf binary to be used via jfvm. Useful for development and testing custom builds.",
	Examples: []Example{
		{
			Command:     "jfvm link --from /Users/dev/go/bin/jf --name local-dev",
			Description: "Link a local binary as 'local-dev'",
		},
		{
			Command:     "jfvm link --from ./jf --name custom-build",
			Description: "Link relative path binary as 'custom-build'",
		},
	},
}

var Compare = CommandDescription{
	Usage:       "Compare JFrog CLI command output between versions",
	Description: "Compare JFrog CLI command output between two versions in parallel with git-like diff visualization. Measures execution time, success rate, and highlights differences.",
	Examples: []Example{
		{
			Command:     "jfvm compare 2.74.0 2.73.0 -- --version",
			Description: "Compare version output between two releases",
		},
		{
			Command:     "jfvm compare prod dev -- rt ping",
			Description: "Compare command outputs using aliases",
		},
		{
			Command:     "jfvm compare 2.74.0 2.73.0 -- config show --unified",
			Description: "Show unified diff format",
		},
		{
			Command:     "jfvm compare old new -- rt search \"*.jar\" --no-color --timing=false",
			Description: "Disable colored output and timing",
		},
	},
}

var Benchmark = CommandDescription{
	Usage:       "Benchmark JFrog CLI command performance across versions",
	Description: "Run performance benchmarks for JFrog CLI commands across multiple versions. Measures execution time, success rate, and provides statistical analysis.",
	Examples: []Example{
		{
			Command:     "jfvm benchmark 2.74.0,2.73.0,2.72.0 -- --version",
			Description: "Benchmark across multiple versions",
		},
		{
			Command:     "jfvm benchmark prod,dev,latest -- rt ping --iterations 10 --detailed",
			Description: "Custom iterations with detailed output",
		},
		{
			Command:     "jfvm benchmark 2.74.0,2.73.0 -- config show --format json",
			Description: "Export results as JSON",
		},
		{
			Command:     "jfvm benchmark 2.74.0,2.73.0 -- rt search \"*.jar\" --format csv",
			Description: "Export results as CSV",
		},
	},
}

var History = CommandDescription{
	Usage:       "Show version usage history and statistics",
	Description: "Display historical usage patterns for JFrog CLI versions. Tracks when versions were used, most common commands, usage trends, and command outputs.",
	Examples: []Example{
		{
			Command:     "jfvm history",
			Description: "Show recent usage history",
		},
		{
			Command:     "jfvm history --limit 20",
			Description: "Show last 20 entries",
		},
		{
			Command:     "jfvm history --stats",
			Description: "Show detailed statistics",
		},
		{
			Command:     "jfvm history --version 2.74.0",
			Description: "Filter by specific version",
		},
		{
			Command:     "jfvm history --show-output",
			Description: "Show captured command outputs",
		},
		{
			Command:     "jfvm history --format json",
			Description: "Export as JSON",
		},
		{
			Command:     "jfvm history --clear",
			Description: "Clear history (cannot be undone)",
		},
	},
}
