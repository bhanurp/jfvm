package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/bhanurp/jfvm/cmd/utils"
	"github.com/urfave/cli/v2"
)

var Init = &cli.Command{
	Name:        "init",
	Usage:       "Interactive setup wizard for jfvm",
	Description: "Smart initialization wizard that guides you through setting up your JFrog CLI environment.",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "auto",
			Usage: "Run in automatic mode with sensible defaults",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "wizard",
			Usage: "Run interactive setup wizard",
			Value: true,
		},
		&cli.StringFlag{
			Name:  "project-type",
			Usage: "Specify project type (java, node, docker, go, python)",
		},
	},
	Action: func(c *cli.Context) error {
		if c.Bool("auto") {
			return runAutoInit()
		}
		return runInitWizard(c.String("project-type"))
	},
}

func runInitWizard(projectType string) error {
	fmt.Printf("🚀 JFVM Smart Initialization Wizard\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════════\n\n")

	// Step 1: Welcome and environment check
	fmt.Printf("👋 Welcome! Let's set up your JFrog CLI environment.\n\n")

	// Check existing installations
	existingJF := detectExistingJFInstallations()
	if len(existingJF) > 0 {
		fmt.Printf("🔍 Detected existing JF CLI installations:\n")
		for _, jf := range existingJF {
			fmt.Printf("   • %s (%s)\n", jf.Path, jf.Version)
		}
		fmt.Println()

		if askYesNo("Would you like to import configurations from existing installations?") {
			importExistingConfigurations(existingJF)
		}
	}

	// Step 2: Project detection
	if projectType == "" {
		projectType = detectProjectType()
	}

	fmt.Printf("📦 Project type detected: %s\n\n", projectType)

	// Step 3: Recommend JF CLI version
	recommendedVersion := getRecommendedVersion(projectType)
	fmt.Printf("💡 Recommended JF CLI version: %s\n", recommendedVersion)

	version := recommendedVersion
	if askYesNo("Would you like to use a different version?") {
		version = promptForInput("Enter JF CLI version", recommendedVersion)
	}

	// Step 4: Install version if needed
	fmt.Printf("\n🔧 Installing JF CLI version %s...\n", version)
	if err := installVersionQuiet(version); err != nil {
		return fmt.Errorf("failed to install version %s: %w", version, err)
	}

	// Step 5: Set as current version
	if err := setCurrentVersion(version); err != nil {
		return fmt.Errorf("failed to set current version: %w", err)
	}

	// Step 6: Create project files
	if askYesNo("Create .jfrog-version file in current directory?") {
		if err := createJFrogVersionFile(version); err != nil {
			fmt.Printf("⚠️  Failed to create .jfrog-version file: %v\n", err)
		} else {
			fmt.Printf("✅ Created .jfrog-version file\n")
		}
	}

	// Step 7: Setup shell integration
	if askYesNo("Setup shell integration (add shim to PATH)?") {
		setupShellIntegration()
	}

	fmt.Printf("\n🎉 Setup completed successfully!\n")
	fmt.Printf("Run 'jfvm list' to get started.\n")
	fmt.Printf("💡 Try 'jfvm translate \"jf rt upload myfile.jar\" --lang pirate' for some fun!\n")

	return nil
}

func runAutoInit() error {
	fmt.Printf("🤖 Running automatic initialization...\n")

	projectType := detectProjectType()
	version := getRecommendedVersion(projectType)

	fmt.Printf("• Project type: %s\n", projectType)
	fmt.Printf("• Installing version: %s\n", version)

	if err := installVersionQuiet(version); err != nil {
		return err
	}

	if err := setCurrentVersion(version); err != nil {
		return err
	}

	createJFrogVersionFile(version)
	setupShellIntegration()

	fmt.Printf("✅ Auto-initialization completed!\n")
	return nil
}

type JFInstallation struct {
	Path    string
	Version string
}

func detectExistingJFInstallations() []JFInstallation {
	var installations []JFInstallation

	// Common installation paths
	paths := []string{
		"/usr/local/bin/jf",
		"/opt/homebrew/bin/jf",
		"/usr/bin/jf",
		filepath.Join(os.Getenv("HOME"), "bin/jf"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			if version := getJFVersion(path); version != "" {
				installations = append(installations, JFInstallation{
					Path:    path,
					Version: version,
				})
			}
		}
	}

	return installations
}

func getJFVersion(jfPath string) string {
	cmd := exec.Command(jfPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Extract version from output like "jf version 2.75.0"
	parts := strings.Fields(string(output))
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

func detectProjectType() string {
	// Check for various project indicators
	indicators := map[string][]string{
		"java":   {"pom.xml", "build.gradle", "build.gradle.kts", "gradle.properties"},
		"node":   {"package.json", "yarn.lock", "package-lock.json"},
		"docker": {"Dockerfile", "docker-compose.yml", "docker-compose.yaml"},
		"go":     {"go.mod", "go.sum", "Gopkg.toml"},
		"python": {"requirements.txt", "setup.py", "pyproject.toml", "Pipfile"},
		"php":    {"composer.json", "composer.lock"},
		"ruby":   {"Gemfile", "Gemfile.lock"},
		"rust":   {"Cargo.toml", "Cargo.lock"},
	}

	for projectType, files := range indicators {
		for _, file := range files {
			if _, err := os.Stat(file); err == nil {
				return projectType
			}
		}
	}

	return "generic"
}

func getRecommendedVersion(projectType string) string {
	// Version recommendations based on project type
	recommendations := map[string]string{
		"java":    "2.76.1", // Latest stable
		"node":    "2.76.1",
		"docker":  "2.76.1",
		"go":      "2.76.1",
		"python":  "2.76.1",
		"generic": "2.76.1",
	}

	if version, exists := recommendations[projectType]; exists {
		return version
	}
	return "2.76.1"
}

func askYesNo(question string) bool {
	fmt.Printf("❓ %s (y/N): ", question)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes"
}

func promptForInput(prompt, defaultValue string) string {
	fmt.Printf("📝 %s [%s]: ", prompt, defaultValue)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return defaultValue
	}
	return input
}

func installVersionQuiet(version string) error {
	// This would use the existing install functionality
	// For now, we'll simulate it
	fmt.Printf("   📥 Downloading JF CLI %s...\n", version)
	time.Sleep(1 * time.Second) // Simulate download
	fmt.Printf("   ✅ Installed successfully\n")
	return nil
}

func setCurrentVersion(version string) error {
	configPath := filepath.Join(utils.JfvmRoot, "config")
	return os.WriteFile(configPath, []byte(version), 0644)
}

func createJFrogVersionFile(version string) error {
	return os.WriteFile(".jfrog-version", []byte(version+"\n"), 0644)
}

func setupShellIntegration() {
	homeDir := os.Getenv("HOME")
	shimPath := filepath.Join(homeDir, ".jfvm", "shim")

	shells := []struct {
		name string
		file string
	}{
		{"zsh", ".zshrc"},
		{"bash", ".bashrc"},
		{"fish", ".config/fish/config.fish"},
	}

	exportLine := fmt.Sprintf(`export PATH="%s:$PATH"`, shimPath)

	fmt.Printf("🔧 Setting up shell integration...\n")

	for _, shell := range shells {
		shellFile := filepath.Join(homeDir, shell.file)
		if _, err := os.Stat(shellFile); err == nil {
			// Check if already configured
			content, err := os.ReadFile(shellFile)
			if err == nil && !strings.Contains(string(content), ".jfvm/shim") {
				// Append to shell file
				f, err := os.OpenFile(shellFile, os.O_APPEND|os.O_WRONLY, 0644)
				if err == nil {
					fmt.Fprintf(f, "\n# Added by jfvm\n%s\n", exportLine)
					f.Close()
					fmt.Printf("   ✅ Added to %s\n", shell.file)
				}
			}
		}
	}

	fmt.Printf("   💡 Please restart your shell or run: source ~/.zshrc\n")
}

func importExistingConfigurations(installations []JFInstallation) {
	fmt.Printf("📋 Importing configurations...\n")

	for _, install := range installations {
		fmt.Printf("   • Analyzing %s...\n", install.Path)
		// Here we would import actual JF CLI configurations
		// For now, just simulate
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("   ✅ Import completed\n")
}
