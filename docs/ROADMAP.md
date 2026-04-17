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

## v0.2.2 ✅

- `luckyclaw install` — sets up init script, SSH banner, and OOM protection on stock Buildroot (no reflash needed)
- Native WhatsApp channel (whatsmeow, QR pairing, quoted replies, media, deduplication)
- Dynamic board detection — MemTotal-based (Plus/Pro/Max), replaces unreliable device tree matching
- Dynamic GOMEMLIMIT per board variant (24/48/96MiB) with GOGC=20
- Init script reads timezone from config.json (UTC fallback)
- Safe gateway startup after onboarding — uses init script `start` to ensure env vars
- Process-safe stop/restart — no `killall` that kills onboarding or user commands
- SSH banner shows board name, memory, RSS, GOMEMLIMIT
- NTP wait on boot — prevents TLS failures on cold boot
- OOM protection — `oom_score_adj=-200` set after daemon start
- Heartbeat self-chat alerts routed to WhatsApp
- Session binary guard & dynamic context window (queried from OpenRouter during onboarding)
- Port `registry_test.go` from upstream (tool registry test coverage)
- Port `shell_process_unix.go` from upstream (process group cleanup for exec tool)
- Port Empty Response Message Fix (`100720b`) from upstream for stability
- `scripts/sync-overlay.sh` for SDK overlay synchronization

## v0.2.3 ✅

- `luckyclaw set-ip <IP>` — set static IP with auto-detected gateway/subnet, auto-reboot
- `luckyclaw set-ip --dhcp` — restore DHCP (auto-reboot)
- Init script `override_static_ip()` — kills vendor `udhcpc` and reapplies static config on boot

## v0.2.4 ✅

- **Storage fix: Workspace relocated to rootfs** — workspace now lives at `/root/.luckyclaw/workspace/` instead of `/oem/.luckyclaw/workspace/`. This uses rootfs free space (~143MB) instead of the cramped `/oem` partition (~20MB). Config and heartbeat log stay on `/oem`. Existing workspaces trigger a migration notice on upgrade.
- `luckyclaw set-ip <IP>` — set static IP with IPv4 validation (octets 0-255, gateway collision check, subnet validation) and confirmation prompt before applying
- `luckyclaw set-ip --dhcp` — restore DHCP with confirmation prompt
- Init script `override_static_ip()` — enhanced with validation-aware static IP application

## v0.2.5 (Planned)

- Evaluate MCP (Model Context Protocol) support — external tool server integration
- Smart one-time cron fallback — auto-detect fully-specified cron expressions and set `DeleteAfterRun=true` so one-time clock reminders don't leave orphaned jobs
- Explore pre-emptive context compression — compress history before API call instead of after 400 error

## v0.2.x (Future Minor)

- Auto-update command (`luckyclaw update`) — binary-only OTA updates
- Tool definition caching
- Session save optimization (json.Marshal vs MarshalIndent)

## Future

- System prompt caching (requires dynamic/static section split to avoid stale timestamps)
- Telegram MarkdownV2 sanitizer (`parse_markdown_to_md_v2.go`) port
- Custom DNS backup resolver (`0fe0582`) port
- `at_time` parameter for cron tool (ISO-8601 absolute time) — revisit if LLM behavior changes
- Cross-platform flashing tool (Windows/Linux/macOS replacement for SOCToolKit)
- Agent browser skill — see IMPROVEMENTS.md (Pro/Max boards only, 100MB+ RAM required)
- GitHub PR review skill — see IMPROVEMENTS.md

## Upstream Watchlist

Items from PicoClaw upstream that may be worth integrating if they mature and benefit everyday users:

- History compression retry logic — better multi-byte/CJK handling
- Token masking in logs — hides bot tokens from log output (security)
- Symlinked path whitelist fix — tool path security hardening
- `pkg/identity` — identity/personality management (336 lines)
