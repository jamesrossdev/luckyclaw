# LuckyClaw Roadmap

Items are prioritized by readiness and impact. Items may be moved between versions or dropped based on progress and real-world usage feedback.

## v0.2.0 (Current Release)

- Heartbeat hardening (HeartbeatMode, SilentResult, audit logging)
- Memory optimization (GOMEMLIMIT tuning, GOGC=20)
- Flashing guide with backup/restore documentation
- SSH banner and init script improvements
- Default response improvement (echoes user's question on failure)

## v0.2.x (Patch Releases)

- Port `registry_test.go` from upstream PicoClaw (tool registry test coverage)
- Port `shell_process_unix.go` from upstream (process group cleanup for exec tool)
- Performance benchmark tests (`make bench`)
- System prompt caching between messages
- Cron tool `at_time` parameter (ISO-8601 absolute time for reminders)

## v0.3.x (Next Minor)

- Auto-update command (`luckyclaw update`) -- binary-only OTA updates
- WhatsApp channel integration
- Tool definition caching
- Versioned firmware image naming in build pipeline
- Session save optimization (json.Marshal vs MarshalIndent)

## Future

- Cross-platform flashing tool (replace Windows-only SOCToolKit)
- Multi-model routing (small model for easy tasks, large for hard)
- Skill marketplace / remote skill install

## Upstream Watchlist

Items from PicoClaw upstream that may be worth integrating if they mature:

- `pkg/routing` -- model routing (1,103 lines, added upstream post-fork)
- `pkg/media` -- media handling for attachments (801 lines)
- `shell_process_windows.go` -- Windows cross-platform support (28 lines)
- `pkg/identity` -- identity/personality management (336 lines)
