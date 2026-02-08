# WhenIdle Roadmap

## v1.0.0 (Released)

### Core Features
- [x] CPU monitoring with configurable threshold
- [x] Automatic task launching during idle periods
- [x] Task suspension/resume with SIGSTOP/SIGCONT
- [x] Graceful shutdown with SIGTERM → SIGKILL timeout
- [x] JSON configuration support
- [x] Launch Agent macOS integration
- [x] Comprehensive test suite (13 unit tests)
- [x] Complete documentation (README, CHANGELOG, inline comments)
- [x] Installation/uninstallation scripts

**Status**: ✅ DONE
**Completed**: 2026-02-08
**PR**: Feature branch feature/when-idle-daemon ready for merge

---

## v1.1.0 (In Development)

### Enhancements
- [x] GUI system tray interface with configuration editor
- [x] Dynamic icons indicating state (idle/running/paused/disabled)
- [x] Live log viewer in GUI
- [x] Ring buffer for log capturing (500 lines)
- [ ] Support for multiple concurrent tasks
- [ ] Statistics and metrics dashboard
- [ ] Support for `--strategy=stop-restart` (alternative to SIGSTOP/SIGCONT)
- [ ] Linux support (in addition to macOS)

### Quality
- [x] Unit tests: 24 passing (65% coverage)
- [x] Code review completed
- [x] Documentation updated (README, CHANGELOG)
- [ ] Code coverage > 80%
- [ ] Performance benchmarking
- [ ] Security audit

**Status**: ✅ DONE (GUI feature complete)
**Completed**: 2026-02-08
**PR**: feature/gui-system-tray ready for merge
**Commits**: 6 commits (320a21e → 64c0183)

---

## v2.0.0 (Planned - Long-term)

### Major Features
- [ ] Scheduling with cron-like expressions
- [ ] Web API for remote control
- [ ] Task priority and queueing
- [ ] Advanced resource limits (CPU%, memory%)
- [ ] Checkpoint/restore support

---

## Known Limitations (v1.0.0)

| Limitation | Impact | Mitigation |
|-----------|--------|------------|
| Single task per daemon | Need multiple daemons for multiple tasks | Create multiple configs + Launch Agents |
| Memory occupied during suspension | Resource usage for long suspensions | Monitor suspension duration in logs |
| Long suspension may cause network timeouts | Rare case for typical workloads | Avoid network-dependent tasks; restart instead |

---

## Testing Status

| Category | Status | v1.0.0 | v1.1.0 | Notes |
|----------|--------|--------|--------|-------|
| Unit tests | ✅ | 13/13 | 24/24 | 60% → 65% coverage |
| Integration tests | ✅ | Verified | Verified | Manual CPU load testing |
| GUI tests | ⚠️ | N/A | Manual | Fyne difficult to mock, manual testing OK |
| macOS compatibility | ✅ | Darwin 23.5.0 | Darwin 23.5.0 | System tray, Dock hiding verified |
| Edge cases | ✅ | Handled | Handled | Process group signaling, LogBuffer ring buffer |

---

## Release History

### v1.0.0 - 2026-02-07
Initial release
- Feature: Complete core functionality
- Quality: 13 unit tests passing
- Docs: Full README, CHANGELOG, inline comments
- Status: Ready for production

### v1.1.0 - 2026-02-08 (In Development)
GUI enhancement release
- Feature: System tray GUI with config editor, log viewer
- Quality: 24 unit tests passing (65% coverage)
- Architecture: Fyne.io, LogBuffer ring buffer, HideFromDock CGo
- Docs: Updated README, CHANGELOG, ADR-002, learnings documented
- Status: Ready for code review and merge

**Development Metrics**:
- Duration: ~9-10 hours
- Lines added: +905
- Commits: 6 (320a21e → 64c0183)
- Branch: feature/gui-system-tray

---

## Next Steps (Post-v1.1.0)

- [ ] Code review approval → Merge to main
- [ ] User feedback collection on GUI
- [ ] Performance monitoring in production
- [ ] Plan v1.2 features based on feedback
- [ ] Consider Linux support for v2.0
