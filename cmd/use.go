package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bhanurp/jfvm/cmd/descriptions"
	"github.com/bhanurp/jfvm/cmd/utils"
	"github.com/bhanurp/jfvm/internal"
	"github.com/urfave/cli/v2"
)

var Use = &cli.Command{
	Name:        "use",
	Usage:       descriptions.Use.Usage,
	ArgsUsage:   "[version or alias] (optional if .jfrog-version exists)",
	Description: descriptions.Use.Format(),
	Action: func(c *cli.Context) error {
		fmt.Println("Executing 'jfvm use' command...")
		var version string

		if c.Args().Len() == 1 {
			v := c.Args().Get(0)
			fmt.Printf("Received argument: %s\n", v)

			// Try to resolve alias (silently fallback if not found)
			resolved, err := utils.ResolveAlias(v)
			if err == nil {
				version = strings.TrimSpace(resolved)
				fmt.Printf("Using alias '%s' resolved to version: %s\n", v, version)
			} else {
				// don't log anything — just fallback silently
				version = v
			}
		} else {
			v, err := utils.GetVersionFromProjectFile()
			if err != nil {
				return cli.Exit("No version provided and no .jfrog-version file found", 1)
			}
			version = v
			fmt.Printf("Using version from .jfrog-version: %s\n", version)
		}

		binPath := filepath.Join(utils.JfvmVersions, version, utils.BinaryName)
		fmt.Printf("Checking if binary exists at: %s\n", binPath)

		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			fmt.Printf("Version %s not found locally. Installing...\n", version)
			if err := internal.DownloadAndInstall(version); err != nil {
				return fmt.Errorf("auto-install failed: %w", err)
			}
		}

		fmt.Printf("Writing selected version '%s' to config file: %s\n", version, utils.JfvmConfig)
		return os.WriteFile(utils.JfvmConfig, []byte(version), 0644)
	},
}
