# LuckyClaw Roadmap

Items are prioritized by readiness and impact. Items may be moved between versions or dropped based on progress and real-world usage feedback.

## v0.2.0 ✅

- Heartbeat hardening (HeartbeatMode, SilentResult, audit logging)
- Memory optimization (GOMEMLIMIT tuning, GOGC=20)
- Flashing guide with backup/restore documentation
- SSH banner and init script improvements
- Default response improvement (echoes user's question on failure)

## v0.2.1 ✅

- Discord moderation tools: message deletion, user timeouts (7s–4w)
- Discord DM sandbox bypass — full tool access in DMs, sandboxed in server channels
- User metadata injection — agent sees display name, roles, and DM status in system prompt
- Reasoning model support — thinking tokens hidden from chat, retained in context
- Warning added against using thinking models in Discord server mode
- `discord-mod` community skill template added to workspace
- `firmware/overlay/root/` removed — workspace delivered via `go:embed`, not firmware
- README and AGENTS.md updated to reflect conservative project philosophy
- Pico Pro / Pico Max board compatibility clarified
- Improved memory reporting clarity in status and banner (available / total)

## v0.2.2 (Current)

- `luckyclaw install` — sets up init script, SSH banner, and OOM protection on stock Buildroot (no reflash needed)
- WhatsApp channel integration
- Port `registry_test.go` from upstream (tool registry test coverage)
- Port `shell_process_unix.go` from upstream (process group cleanup for exec tool)
- Port Empty Response Message Fix (`100720b`) from upstream for stability

## v0.2.x (Next Minor)

- Auto-update command (`luckyclaw update`) — binary-only OTA updates
- Tool definition caching
- Session save optimization (json.Marshal vs MarshalIndent)

## Future

- Telegram MarkdownV2 Sanitizer (`parse_markdown_to_md_v2.go`) port
- Custom DNS Backup Resolver (`0fe0582`) port
- Cron tool `at_time` parameter (ISO-8601 absolute time for reminders)
- Cross-platform flashing tool (replace Windows-only SOCToolKit)
- Skill marketplace / remote skill install

## Upstream Watchlist

Items from PicoClaw upstream that may be worth integrating if they mature and benefit everyday users:

- History compression retry logic — better multi-byte/CJK handling
- Token masking in logs — hides bot tokens from log output (security)
- Symlinked path whitelist fix — tool path security hardening
- `pkg/identity` — identity/personality management (336 lines)


