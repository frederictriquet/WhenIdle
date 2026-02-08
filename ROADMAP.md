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

## v1.1.0 (Planned - Future)

### Enhancements
- [ ] GUI configuration editor
- [ ] Support for multiple concurrent tasks
- [ ] Statistics and metrics dashboard
- [ ] Support for `--strategy=stop-restart` (alternative to SIGSTOP/SIGCONT)
- [ ] Linux support (in addition to macOS)

### Quality
- [ ] Code coverage > 80%
- [ ] Performance benchmarking
- [ ] Security audit

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

| Category | Status | Notes |
|----------|--------|-------|
| Unit tests | ✅ 13/13 passing | 60% code coverage |
| Integration tests | ✅ Verified | Manual CPU load testing |
| macOS compatibility | ✅ Verified | Tested on Darwin 23.5.0 |
| Edge cases | ✅ Handled | Process group signaling, timeout SIGKILL |

---

## Release History

### v1.0.0 - 2026-02-08
Initial release
- Feature: Complete core functionality
- Quality: 13 unit tests passing
- Docs: Full README, CHANGELOG, inline comments
- Status: Ready for production

---

## Next Steps (Post-Release)

- [ ] User feedback collection
- [ ] Performance monitoring in the wild
- [ ] Plan v1.1 based on feedback
- [ ] Consider Linux support for v2.0
