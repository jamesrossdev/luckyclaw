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
- **Allowlist normalization fix** — WhatsApp sender IDs now canonicalize correctly for allowlist matching, resolving message drops after restart
- **Heartbeat template migration** — now non-destructive: preserves user-edited `HEARTBEAT.md` when custom tasks exist, creates backup before auto-migration
- **`etc/` overlay sync restored** — `scripts/sync-overlay.sh` now syncs init script and SSH banner into SDK overlay again (was inadvertently dropped in v0.2.4 development)
- **Onboarding idempotency** — re-running `luckyclaw onboard` now merges with existing config (preserving values) and skips workspace template overwrite for non-skill files
- **Safe token clamp** — `max_tokens` is auto-clamped to `min(20% of context_window, 16384, provider_max_output)`, floor 1024, preventing context-window overflow errors on large-context models like DeepSeek v3.2

## v0.2.5 ✅

- **Unified onboarding** — single provider menu (all 13 supported), no quick/advanced split. Auto-fetches model metadata (context window, thinking support, vision support) from OpenRouter public API. Manual context window fallback when metadata not found.
- **DeepSeek fixes** — reasoning_content preserved across agent turns (prevents "must be passed back" API errors). Vision-unsupported error detection with automatic retry without images.
- **Reasoning controls** — `show_reasoning` config toggle (default false). Think sanitizer strips orphan/malformed think tags. Separate reasoning block format when enabled.
- **Provider improvements** — MiniMax base_resp error handling, reasoning_split support. DeepSeek API base fix (removed /v1 suffix). OpenRouter default model changed to `nvidia/nemotron-3-super-120b-a12b:free`.
- **WhatsApp status filter** — `ignore_status_updates` config toggle (default true) blocks status@broadcast messages.
- **Channel skill filter** — `channel_skill_filter` config toggle (default false) hides discord-mod outside Discord, whatsapp outside WhatsApp.
- **Agent CLI quiet mode** — non-debug agent mode suppresses tool execution logs for clean conversational output.
- **Dynamic status providers** — `luckyclaw status` shows all configured providers, not just 5 hardcoded ones.
- **Config-reset separation** — fresh onboard no longer touches config; `luckyclaw config-reset` handles intentional deletion.

## v0.2.6 (Planned)

- Auto-update command (`luckyclaw update`) — binary-only OTA updates
- Smart one-time cron fallback — auto-detect fully-specified cron expressions and set `DeleteAfterRun=true`
- Evaluate pre-emptive context compression — compress history before API call instead of after 400 error

## Future

- MCP (Model Context Protocol) support — external tool server integration (evaluated, deferred due to complexity/RAM concerns on Pico Plus)
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
