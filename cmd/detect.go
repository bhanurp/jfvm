package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

type ProjectInfo struct {
	Type         string           `json:"type"`
	Language     string           `json:"language"`
	BuildTool    string           `json:"build_tool"`
	Dependencies []string         `json:"dependencies"`
	Indicators   []string         `json:"indicators"`
	Recommended  RecommendedSetup `json:"recommended"`
}

type RecommendedSetup struct {
	JFVersion     string   `json:"jf_version"`
	Repositories  []string `json:"repositories"`
	BuildCommands []string `json:"build_commands"`
	UploadPaths   []string `json:"upload_paths"`
}

var Detect = &cli.Command{
	Name:        "detect",
	Usage:       "Detect project type and recommend JF CLI setup",
	Description: "Analyzes the current project and provides intelligent recommendations for JF CLI configuration.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "path",
			Usage: "Path to analyze (default: current directory)",
			Value: ".",
		},
		&cli.BoolFlag{
			Name:  "json",
			Usage: "Output in JSON format",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "verbose",
			Usage: "Show detailed analysis",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "setup",
			Usage: "Apply recommended setup automatically",
			Value: false,
		},
	},
	Action: func(c *cli.Context) error {
		projectInfo := analyzeProject(c.String("path"), c.Bool("verbose"))

		if c.Bool("json") {
			return outputJSON(projectInfo)
		}

		displayProjectInfo(projectInfo, c.Bool("verbose"))

		if c.Bool("setup") {
			return applyRecommendedSetup(projectInfo)
		}

		return nil
	},
}

func analyzeProject(projectPath string, verbose bool) ProjectInfo {
	if verbose {
		fmt.Printf("🔍 Analyzing project...\n")
		fmt.Printf("📁 Path: %s\n\n", projectPath)
	}

	info := ProjectInfo{
		Type:         "unknown",
		Language:     "unknown",
		BuildTool:    "unknown",
		Indicators:   []string{},
		Dependencies: []string{},
	}

	// Detect project type and language
	detectors := []func(string, *ProjectInfo, bool){
		detectJava,
		detectNode,
		detectDocker,
		detectGo,
		detectPython,
		detectPHP,
		detectRuby,
		detectRust,
		detectDotNet,
	}

	for _, detector := range detectors {
		detector(projectPath, &info, verbose)
		if info.Type != "unknown" {
			break
		}
	}

	// Generate recommendations
	info.Recommended = generateRecommendations(info)

	if verbose {
		fmt.Printf("✅ Analysis complete!\n\n")
	}

	return info
}

func detectJava(projectPath string, info *ProjectInfo, verbose bool) {
	javaFiles := []string{"pom.xml", "build.gradle", "build.gradle.kts", "gradle.properties"}

	for _, file := range javaFiles {
		if fileExists(filepath.Join(projectPath, file)) {
			info.Type = "java"
			info.Language = "Java"
			info.Indicators = append(info.Indicators, file)

			if strings.Contains(file, "pom.xml") {
				info.BuildTool = "Maven"
				analyzeMavenProject(projectPath, info, verbose)
			} else if strings.Contains(file, "gradle") {
				info.BuildTool = "Gradle"
				analyzeGradleProject(projectPath, info, verbose)
			}
			return
		}
	}
}

func detectNode(projectPath string, info *ProjectInfo, verbose bool) {
	nodeFiles := []string{"package.json", "yarn.lock", "package-lock.json"}

	for _, file := range nodeFiles {
		if fileExists(filepath.Join(projectPath, file)) {
			info.Type = "node"
			info.Language = "JavaScript/TypeScript"
			info.Indicators = append(info.Indicators, file)

			if file == "package.json" {
				analyzePackageJSON(projectPath, info, verbose)
			}

			if file == "yarn.lock" {
				info.BuildTool = "Yarn"
			} else if file == "package-lock.json" {
				info.BuildTool = "NPM"
			}
			return
		}
	}
}

func detectDocker(projectPath string, info *ProjectInfo, verbose bool) {
	dockerFiles := []string{"Dockerfile", "docker-compose.yml", "docker-compose.yaml"}

	for _, file := range dockerFiles {
		if fileExists(filepath.Join(projectPath, file)) {
			info.Type = "docker"
			info.BuildTool = "Docker"
			info.Indicators = append(info.Indicators, file)

			if verbose {
				fmt.Printf("🐳 Found Docker configuration: %s\n", file)
			}
			return
		}
	}
}

func detectGo(projectPath string, info *ProjectInfo, verbose bool) {
	goFiles := []string{"go.mod", "go.sum", "Gopkg.toml"}

	for _, file := range goFiles {
		if fileExists(filepath.Join(projectPath, file)) {
			info.Type = "go"
			info.Language = "Go"
			info.BuildTool = "Go Modules"
			info.Indicators = append(info.Indicators, file)

			if verbose {
				fmt.Printf("🔷 Found Go project: %s\n", file)
			}
			return
		}
	}
}

func detectPython(projectPath string, info *ProjectInfo, verbose bool) {
	pythonFiles := []string{"requirements.txt", "setup.py", "pyproject.toml", "Pipfile"}

	for _, file := range pythonFiles {
		if fileExists(filepath.Join(projectPath, file)) {
			info.Type = "python"
			info.Language = "Python"
			info.Indicators = append(info.Indicators, file)

			if file == "Pipfile" {
				info.BuildTool = "Pipenv"
			} else if file == "pyproject.toml" {
				info.BuildTool = "Poetry"
			} else {
				info.BuildTool = "pip"
			}

			if verbose {
				fmt.Printf("🐍 Found Python project: %s\n", file)
			}
			return
		}
	}
}

func detectPHP(projectPath string, info *ProjectInfo, verbose bool) {
	if fileExists(filepath.Join(projectPath, "composer.json")) {
		info.Type = "php"
		info.Language = "PHP"
		info.BuildTool = "Composer"
		info.Indicators = append(info.Indicators, "composer.json")

		if verbose {
			fmt.Printf("🐘 Found PHP project: composer.json\n")
		}
	}
}

func detectRuby(projectPath string, info *ProjectInfo, verbose bool) {
	if fileExists(filepath.Join(projectPath, "Gemfile")) {
		info.Type = "ruby"
		info.Language = "Ruby"
		info.BuildTool = "Bundler"
		info.Indicators = append(info.Indicators, "Gemfile")

		if verbose {
			fmt.Printf("💎 Found Ruby project: Gemfile\n")
		}
	}
}

func detectRust(projectPath string, info *ProjectInfo, verbose bool) {
	if fileExists(filepath.Join(projectPath, "Cargo.toml")) {
		info.Type = "rust"
		info.Language = "Rust"
		info.BuildTool = "Cargo"
		info.Indicators = append(info.Indicators, "Cargo.toml")

		if verbose {
			fmt.Printf("🦀 Found Rust project: Cargo.toml\n")
		}
	}
}

func detectDotNet(projectPath string, info *ProjectInfo, verbose bool) {
	dotnetFiles := []string{"*.csproj", "*.fsproj", "*.vbproj", "*.sln"}

	for _, pattern := range dotnetFiles {
		matches, _ := filepath.Glob(filepath.Join(projectPath, pattern))
		if len(matches) > 0 {
			info.Type = "dotnet"
			info.Language = ".NET"
			info.BuildTool = "MSBuild"
			info.Indicators = append(info.Indicators, pattern)

			if verbose {
				fmt.Printf("🔵 Found .NET project: %s\n", pattern)
			}
			return
		}
	}
}

func analyzeMavenProject(projectPath string, info *ProjectInfo, verbose bool) {
	pomPath := filepath.Join(projectPath, "pom.xml")
	if content, err := os.ReadFile(pomPath); err == nil {
		pomContent := string(content)

		// Look for common dependencies
		dependencies := []string{}
		if strings.Contains(pomContent, "spring") {
			dependencies = append(dependencies, "Spring")
		}
		if strings.Contains(pomContent, "junit") {
			dependencies = append(dependencies, "JUnit")
		}
		if strings.Contains(pomContent, "jackson") {
			dependencies = append(dependencies, "Jackson")
		}

		info.Dependencies = dependencies

		if verbose && len(dependencies) > 0 {
			fmt.Printf("📦 Maven dependencies detected: %s\n", strings.Join(dependencies, ", "))
		}
	}
}

func analyzeGradleProject(projectPath string, info *ProjectInfo, verbose bool) {
	gradleFiles := []string{"build.gradle", "build.gradle.kts"}

	for _, file := range gradleFiles {
		if content, err := os.ReadFile(filepath.Join(projectPath, file)); err == nil {
			gradleContent := string(content)

			dependencies := []string{}
			if strings.Contains(gradleContent, "spring") {
				dependencies = append(dependencies, "Spring")
			}
			if strings.Contains(gradleContent, "junit") {
				dependencies = append(dependencies, "JUnit")
			}

			info.Dependencies = dependencies

			if verbose && len(dependencies) > 0 {
				fmt.Printf("📦 Gradle dependencies detected: %s\n", strings.Join(dependencies, ", "))
			}
			break
		}
	}
}

func analyzePackageJSON(projectPath string, info *ProjectInfo, verbose bool) {
	packagePath := filepath.Join(projectPath, "package.json")
	if content, err := os.ReadFile(packagePath); err == nil {
		var pkg map[string]interface{}
		if json.Unmarshal(content, &pkg) == nil {
			dependencies := []string{}

			// Check dependencies
			if deps, ok := pkg["dependencies"].(map[string]interface{}); ok {
				for dep := range deps {
					if strings.Contains(dep, "react") {
						dependencies = append(dependencies, "React")
					} else if strings.Contains(dep, "vue") {
						dependencies = append(dependencies, "Vue")
					} else if strings.Contains(dep, "angular") {
						dependencies = append(dependencies, "Angular")
					} else if strings.Contains(dep, "express") {
						dependencies = append(dependencies, "Express")
					}
				}
			}

			info.Dependencies = dependencies

			if verbose && len(dependencies) > 0 {
				fmt.Printf("📦 NPM dependencies detected: %s\n", strings.Join(dependencies, ", "))
			}
		}
	}
}

func generateRecommendations(info ProjectInfo) RecommendedSetup {
	setup := RecommendedSetup{
		JFVersion:     "2.76.1", // Default to latest
		Repositories:  []string{},
		BuildCommands: []string{},
		UploadPaths:   []string{},
	}

	switch info.Type {
	case "java":
		setup.Repositories = []string{"libs-release-local", "libs-snapshot-local"}
		if info.BuildTool == "Maven" {
			setup.BuildCommands = []string{"mvn clean package"}
			setup.UploadPaths = []string{"target/*.jar", "target/*.war"}
		} else if info.BuildTool == "Gradle" {
			setup.BuildCommands = []string{"./gradlew build"}
			setup.UploadPaths = []string{"build/libs/*.jar"}
		}

	case "node":
		setup.Repositories = []string{"npm-local"}
		setup.BuildCommands = []string{"npm run build"}
		setup.UploadPaths = []string{"dist/", "build/"}

	case "docker":
		setup.Repositories = []string{"docker-local"}
		setup.BuildCommands = []string{"docker build -t myapp ."}

	case "go":
		setup.Repositories = []string{"go-local"}
		setup.BuildCommands = []string{"go build"}

	case "python":
		setup.Repositories = []string{"pypi-local"}
		setup.BuildCommands = []string{"python setup.py sdist bdist_wheel"}
		setup.UploadPaths = []string{"dist/*.whl", "dist/*.tar.gz"}
	}

	return setup
}

func displayProjectInfo(info ProjectInfo, verbose bool) {
	fmt.Printf("🔍 PROJECT ANALYSIS RESULTS\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════════\n\n")

	fmt.Printf("📋 Project Information:\n")
	fmt.Printf("   Type: %s\n", info.Type)
	fmt.Printf("   Language: %s\n", info.Language)
	fmt.Printf("   Build Tool: %s\n", info.BuildTool)

	if len(info.Dependencies) > 0 {
		fmt.Printf("   Dependencies: %s\n", strings.Join(info.Dependencies, ", "))
	}

	if verbose && len(info.Indicators) > 0 {
		fmt.Printf("   Detected Files: %s\n", strings.Join(info.Indicators, ", "))
	}

	fmt.Printf("\n💡 Recommendations:\n")
	fmt.Printf("   JF CLI Version: %s\n", info.Recommended.JFVersion)

	if len(info.Recommended.Repositories) > 0 {
		fmt.Printf("   Suggested Repositories: %s\n", strings.Join(info.Recommended.Repositories, ", "))
	}

	if len(info.Recommended.BuildCommands) > 0 {
		fmt.Printf("   Build Commands: %s\n", strings.Join(info.Recommended.BuildCommands, " && "))
	}

	if len(info.Recommended.UploadPaths) > 0 {
		fmt.Printf("   Upload Paths: %s\n", strings.Join(info.Recommended.UploadPaths, ", "))
	}

	fmt.Printf("\n🚀 To apply these recommendations, run: jfvm detect --setup\n")
}

func outputJSON(info ProjectInfo) error {
	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}

func applyRecommendedSetup(info ProjectInfo) error {
	fmt.Printf("🚀 Applying recommended setup...\n")

	// Create .jfrog-version file
	if err := createJFrogVersionFile(info.Recommended.JFVersion); err != nil {
		fmt.Printf("⚠️  Failed to create .jfrog-version: %v\n", err)
	} else {
		fmt.Printf("✅ Created .jfrog-version with version %s\n", info.Recommended.JFVersion)
	}

	// Install recommended version
	fmt.Printf("📥 Installing JF CLI version %s...\n", info.Recommended.JFVersion)
	if err := installVersionQuiet(info.Recommended.JFVersion); err != nil {
		return fmt.Errorf("failed to install version: %w", err)
	}

	fmt.Printf("✅ Setup completed!\n")
	return nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
