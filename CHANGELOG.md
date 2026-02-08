# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-02-08

### Added
- Initial release of WhenIdle daemon for macOS
- CPU monitoring with configurable threshold and idle duration
- Automatic task launching during idle periods
- Task suspension/resume with SIGSTOP/SIGCONT when CPU becomes active/idle
- Process group signaling for full process tree management
- Graceful shutdown with SIGTERM → SIGKILL timeout (5 seconds)
- JSON configuration file support
- Launch Agent macOS integration for auto-start at login
- Installation and uninstallation scripts
- Comprehensive test suite (13 unit tests, 60% coverage)
- Complete documentation (README, examples, FAQ)

### Technical Details
- Built with Go 1.18+
- Single dependency: gopsutil v4 for CPU monitoring
- Binary size: ~3.4 MB
- Memory footprint: ~5-10 MB at runtime
- Non-blocking CPU measurement (0 second interval)

[Unreleased]: https://github.com/user/whenidle/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/user/whenidle/releases/tag/v1.0.0
