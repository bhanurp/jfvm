# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **📤 Command Output Capture**: History now captures stdout, stderr, and exit codes for comprehensive debugging and analysis
- **🔍 Enhanced History Display**: New `--show-output` flag to view captured command outputs in history
- **📊 Exit Code Tracking**: Track command success/failure rates in usage analytics

### Changed
- Enhanced HistoryEntry struct to include output capture fields
- Improved history display with exit code indicators and output viewing
- Added output size limits (5KB max per command) to prevent bloated history files

## [0.0.2] - 2024-12-XX

### Added
- **🔍 Version Comparison**: Side-by-side command output comparison with git-like diffs
- **⚡ Performance Benchmarking**: Multi-version performance analysis with statistics
- **📊 Usage Analytics**: Comprehensive history tracking and usage insights
- **🚀 Enhanced Shim**: Automatic command timing and history recording
- **🎨 Rich Output**: Colored terminal output with multiple format options
- **⚙️ Advanced Options**: Configurable timeouts, iterations, and output formats

### Changed
- Refactored benchmark command for better maintainability

### Fixed
- Various bug fixes and performance improvements

## [0.0.1] - Initial Release

### Added
- Basic version management functionality
- Version switching and installation
- Alias support for version management 