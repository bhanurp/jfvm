# jfvm - JFrog CLI Version Manager

[![CI](https://github.com/bhanurp/jfvm/actions/workflows/release.yml/badge.svg)](https://github.com/bhanurp/jfvm/actions/workflows/release.yml)
[![Latest Release](https://img.shields.io/github/v/release/bhanurp/jfvm)](https://github.com/bhanurp/jfvm/releases)
[![License](https://img.shields.io/github/license/bhanurp/jfvm)](https://github.com/bhanurp/jfvm/blob/main/LICENSE)
[![homebrew installs](https://img.shields.io/badge/homebrew-installs-brightgreen?logo=homebrew)](https://github.com/bhanurp/homebrew-jfvm)

**jfvm** is a powerful CLI tool that helps you manage multiple versions of the [JFrog CLI](https://jfrog.com/getcli/) on your system. It supports auto-installation, version switching, project-specific defaults, local binary linking, aliasing, parallel version comparison, performance benchmarking, and usage analytics — all inspired by tools like `nvm`, `sdkman`, and `volta`.

## 🎥 Demo

![jfvm demo](https://user-images.githubusercontent.com/your-username/demo.gif)

> 📸 Replace the above link with a real GIF after recording a demo.

## 🚀 Why jfvm?

Managing different versions of the JFrog CLI across multiple projects and environments can be challenging. `jfvm` simplifies this by:

- Installing any released version of the `jf` binary
- Allowing you to link locally built versions
- Automatically switching versions based on a `.jfrog-version` file
- Letting you define named aliases (e.g., `prod`, `dev`)
- Providing a smooth `jf` shim for command redirection
- Parallel command comparison between versions with diff visualization
- Performance benchmarking across multiple versions
- Usage history tracking and analytics

No more symlink hacking or hardcoded paths.

---

## 🛠️ Installation

### Via Homebrew (with tap):
```bash
brew tap bhanurp/jfvm
brew install jfvm
```

### Via one-liner:
```bash
brew install https://raw.githubusercontent.com/bhanureddy/homebrew-jfvm/main/jfvm.rb
```

### Or Build From Source:
```bash
git clone https://github.com/bhanurp/jfvm.git
cd jfvm
make install
```

---

## 📦 Commands

### Core Version Management

#### `jfvm install <version>`
Installs the specified version of JFrog CLI (`jf`) from JFrog's public release server.
```bash
jfvm install 2.74.0
```

#### `jfvm use <version or alias>`
Activates the given version or alias. If `.jfrog-version` exists in the current directory, that will be used if no argument is passed.
```bash
jfvm use 2.74.0
jfvm use prod
```

#### `jfvm list`
Shows all installed versions and the currently active one.
```bash
jfvm list
```

#### `jfvm remove <version>`
Removes a specific version of `jf`.
```bash
jfvm remove 2.72.1
```

#### `jfvm clear`
Removes **all** installed versions.
```bash
jfvm clear
```

#### `jfvm alias <n> <version>`
Defines an alias for a specific version.
```bash
jfvm alias dev 2.74.0
```

#### `jfvm link --from <path> --name <n>`
Links a **locally built `jf` binary** to be used via `jfvm`.
```bash
jfvm link --from /Users/bhanu/go/bin/jf --name local-dev
jfvm use local-dev
```

### Advanced Features

#### `jfvm compare <version1> <version2> -- <command>`
Compare JFrog CLI command output between two versions in parallel with git-like diff visualization.

```bash
# Compare version output
jfvm compare 2.74.0 2.73.0 -- --version

# Compare command outputs with side-by-side diff
jfvm compare prod dev -- rt ping

# Show unified diff format
jfvm compare 2.74.0 2.73.0 -- config show --unified

# Disable colored output and timing
jfvm compare old new -- rt search "*.jar" --no-color --timing=false
```

**Features:**
- Parallel execution for faster results
- Side-by-side and unified diff formats
- Colored output highlighting differences
- Execution timing comparison
- Exit code and error output comparison

#### `jfvm benchmark <versions> -- <command>`
Run performance benchmarks across multiple JFrog CLI versions with detailed statistics.

```bash
# Benchmark across multiple versions
jfvm benchmark 2.74.0,2.73.0,2.72.0 -- --version

# Custom iterations and detailed output
jfvm benchmark prod,dev,latest -- rt ping --iterations 10 --detailed

# Export results as JSON or CSV
jfvm benchmark 2.74.0,2.73.0 -- config show --format json
jfvm benchmark 2.74.0,2.73.0 -- rt search "*.jar" --format csv
```

**Features:**
- Configurable iteration counts
- Statistical analysis (min, max, average, success rate)
- Multiple output formats (table, JSON, CSV)
- Parallel execution across versions
- Detailed execution logs
- Performance ranking and speed comparisons

#### `jfvm history`
Track and analyze version usage patterns with comprehensive statistics.

```bash
# Show recent usage history
jfvm history

# Show detailed statistics
jfvm history --stats

# Filter by specific version
jfvm history --version 2.74.0

# Limit number of entries
jfvm history --limit 20

# Export as JSON
jfvm history --format json

# Clear history (cannot be undone)
jfvm history --clear
```

**Features:**
- Automatic usage tracking through the shim
- Command execution timing
- Most used versions and commands
- Usage trends and timeline analysis
- Configurable history limits

---

## 📁 Project-specific Version

Add a `.jfrog-version` file to your repo:
```bash
echo "2.74.0" > .jfrog-version
```
Then run:
```bash
jfvm use
```

---

## ⚙️ Shell Integration
Add this to your shell profile (`.zshrc`, `.bashrc`, etc.):
```bash
export PATH="$HOME/.jfvm/shim:$PATH"
```
This allows the shimmed `jf` command to delegate to the correct version transparently.

### Debug Mode
Set `JFVM_DEBUG=1` to see detailed shim execution information:
```bash
export JFVM_DEBUG=1
# Will show which version is being executed
jf --version
```

---

## 🧪 Advanced Examples

### Comparing Configuration Changes
```bash
# Compare configuration differences between versions
jfvm compare 2.74.0 2.73.0 -- config show --format json

# Check if a specific feature works across versions
jfvm compare old new -- rt search "libs-release-local/*.jar" --limit 5
```

### Performance Analysis
```bash
# Benchmark search performance across versions
jfvm benchmark 2.74.0,2.73.0,2.72.0 -- rt search "*" --limit 100 --iterations 3

# Test upload performance
jfvm benchmark prod,dev -- rt upload test.txt my-repo/ --iterations 5 --detailed
```

### Usage Analytics
```bash
# See your most used JFrog CLI commands
jfvm history --stats

# Track version adoption over time
jfvm history --version 2.74.0
```

### Automation and CI/CD
```bash
# Export benchmark results for CI analysis
jfvm benchmark $OLD_VERSION,$NEW_VERSION -- rt ping --format json > performance.json

# Compare outputs in automated testing
jfvm compare baseline canary -- rt search "*.jar" --unified --no-color
```

---

## 🧼 Uninstall
```bash
rm -rf ~/.jfvm
 # if installed via Homebrew
brew uninstall jfvm
```

---

## 🔧 Advanced Configuration

### History Management
- History is automatically tracked in `~/.jfvm/history.json`
- Limited to 1000 entries to prevent unlimited growth
- Includes command execution timing and metadata

### Performance Optimization
- Commands run in parallel when possible
- Configurable timeouts for long-running operations
- Efficient diff algorithms for large outputs

---

## 📝 Use Cases

### Development Teams
- **Version Testing**: Compare behavior across JFrog CLI versions before upgrading
- **Performance Monitoring**: Track performance regressions between releases
- **Usage Analytics**: Understand which commands and versions are used most

### DevOps Engineers
- **CI/CD Integration**: Automate version comparison in deployment pipelines
- **Performance Benchmarks**: Ensure new versions meet performance requirements
- **Migration Planning**: Analyze compatibility before major version upgrades

### Enterprise Environments
- **Compliance Tracking**: Monitor which versions are being used across teams
- **Performance Optimization**: Identify and optimize slow operations
- **Training Insights**: Understand which commands teams use most frequently

---

## 📬 Feedback / Contributions
PRs and issues welcome! Open source, MIT licensed.

**GitHub:** https://github.com/bhanurp/jfvm
