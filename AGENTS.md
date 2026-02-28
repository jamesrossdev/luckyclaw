# AGENTS.md — Project Context for AI Agents

> If you're an AI agent working on this project, **read this first**. It contains critical context built from months of hands-on work with the hardware.

## What Is LuckyClaw?

LuckyClaw is a fork of [PicoClaw](https://github.com/sipeed/picoclaw) (which itself is based on [nanobot](https://github.com/HKUDS/nanobot)). It's an ultra-lightweight AI assistant optimized for **Luckfox Pico Plus** boards — ARM-based embedded Linux devices with only 64MB DDR2 RAM (~33MB usable).

The binary runs on the board 24/7, connected to Telegram (and other channels), answering user messages via OpenRouter/LLM APIs.

## Architecture

```
cmd/luckyclaw/main.go     ← Entry point, CLI commands, onboarding wizard
pkg/agent/                ← Agent loop (LLM ↔ tools ↔ channels)
pkg/channels/             ← Telegram, Discord, Slack, etc (telego for Telegram)
pkg/bus/                  ← Message bus (inbound/outbound routing)
pkg/config/               ← Config loading/saving (~/.luckyclaw/config.json)
pkg/cron/                 ← Scheduled tasks (reminders etc)
pkg/heartbeat/            ← Periodic heartbeat (checks HEARTBEAT.md)
pkg/tools/                ← Agent tools (exec, read_file, write_file, web, etc)
pkg/skills/               ← Skill system (installable capabilities)
workspace/                ← Embedded templates (USER.md, IDENTITY.md, AGENT.md)
firmware/overlay/         ← Files baked into Luckfox firmware image
```

## Critical Hardware Constraints

| Constraint | Value | Impact |
|-----------|-------|--------|
| Total RAM | 64MB DDR2 | Only ~33MB usable after kernel |
| Usable RAM | ~33MB | Gateway uses 10-14MB, leaves 2-12MB free |
| GOMEMLIMIT | 8MiB | Baked into binary via `applyPerformanceDefaults()` |
| GOGC | 20 | Aggressive GC to prevent RSS growth |
| Flash | 128MB SPI NAND | Limited storage, no swap |
| CPU | ARM Cortex-A7 (RV1103) | Single core, GOARM=7 |

## Changes from PicoClaw

### What We Changed
- **Onboarding**: Simplified to OpenRouter only (was 7 provider choices)
- **Performance**: Baked `GOGC=20` + `GOMEMLIMIT=8MiB` into binary
- **CLI**: Added `luckyclaw stop`, `restart`, `gateway -b` (background)
- **Init script**: Auto-starts gateway on boot with OOM protection
- **SSH banner**: Shows ASCII art, status, memory, all commands on login
- **Default model**: `google/gemini-2.0-flash-exp:free` (free tier)
- **Defaults**: `max_tokens=4096`, `max_tool_iterations=10` (was 8192/20)

### What We Did NOT Change
All PicoClaw channels (Telegram, Discord, QQ, LINE, Slack, WhatsApp, etc.) and tools remain in the codebase. Users can configure any provider via `config.json` directly.

## Lessons Learned (Read These!)

### 1. OOM Killer Targets Go Binaries
Go allocates ~500MB virtual memory (lazy reservations). The Linux OOM killer uses `total-vm` in its scoring, so it kills luckyclaw first even at 10MB RSS. **Fix**: Set `oom_score_adj=-200` in the init script after starting the daemon.

### 2. Cron Service Crashes on Empty JSON
The `loadStore()` function in `pkg/cron/service.go` panicked on empty or corrupted `jobs.json`. **Fix**: Added graceful handling — treats empty/corrupt files as fresh state.

### 3. Init Script Must Bake Environment Variables
The init script at `/etc/init.d/S99luckyclaw` MUST export `GOGC=20`, `GOMEMLIMIT=8MiB`, and `TZ` before starting the daemon. Without these, the binary runs with Go defaults and immediately OOMs.

### 4. Busybox Limitations
Luckfox uses Busybox. `wget` doesn't support HTTPS. `sudo` doesn't exist—you're already root. `curl` isn't available. The Go binary handles all HTTPS via `net/http`.

### 5. DNS Resolution Workaround
Telegram API DNS (`api.telegram.org`) sometimes fails to resolve. The init script adds a static entry to `/etc/hosts`.

### 6. Don't Add Unnecessary Dependencies
Every byte counts. The binary is already ~15MB stripped. Adding dependencies increases memory usage. Always test with `GOMEMLIMIT=8MiB`.

### 7. AI Agent Access to the Device
If you are an AI agent and need to test changes, examine logs, or execute commands directly on the Luckfox Pico hardware, **do not guess the IP or password**. Simply ask the user to provide the SSH IP address and password for the device, and use the `run_command` tool via `sshpass` (e.g., `sshpass -p <password> ssh root@<ip>`).

### 8. Committing and Pushing Code
If you are an AI agent, you **MUST NEVER** commit or push code without explicit permission from the user. When you are asked to commit, you must ensure that the tracked `firmware/overlay` directory is completely up to date with whatever modifications were made inside the untracked `luckfox-pico-sdk` directory. This is the only way secondary developers receive OS-level modifications.

### 9. Execution Requires Approved Implementation Plan
If you are an AI agent, you **MUST NEVER** execute code changes, environment modifications, or configuration adjustments without explicitly drafting an implementation plan and receiving the user's explicit approval first. Do not make unauthorized technical assumptions.

### 10. Multiple Daemon Instances & PID Tracking
If `luckyclaw gateway -b` is executed while a daemon started by `/etc/init.d/S99luckyclaw` is already running it will overwrite the `/var/run/luckyclaw.pid` file. Because the init script only tracks the latest PID, subsequent `stop` or `restart` commands will leave the original daemon alive as a zombie, causing duplicate Telegram processing and hallucinated timestamps in session memory. **Fix:** Going forward, making sure we strictly append `&& killall -9 luckyclaw` alongside the init script (which I've started doing in my deploy commands) completely eliminates the possibility of this happening again.

## Build & Deploy

### Testing Before Commits
Always ensure the CI tests pass before committing any changes. Run:
```bash
make check
```
This runs `deps`, `fmt`, `vet`, and the full `test` suite in one command.

### Cross-Compile
```bash
GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 \
  go build -ldflags "-s -w -X main.version=0.2.0" \
  -o build/luckyclaw-linux-arm ./cmd/luckyclaw
```

### Deploy to Device

> **⚠️ IMPORTANT:** The binary MUST be deployed to `/usr/bin/luckyclaw` — this is where the init script
> (`/etc/init.d/S99luckyclaw`) and PATH (`which luckyclaw`) expect it. Do NOT deploy to `/usr/local/bin/`.
> The running process locks the file, so you must kill it before copying.

```bash
# 1. Kill running process (required — scp fails if binary is locked)
sshpass -p 'luckfox' ssh root@<IP> "killall -9 luckyclaw"

# 2. Copy new binary to /usr/bin/ (NOT /usr/local/bin/)
sshpass -p 'luckfox' scp build/luckyclaw-linux-arm root@<IP>:/usr/bin/luckyclaw

# 3. Restart via init script and verify
sshpass -p 'luckfox' ssh root@<IP> "chmod +x /usr/bin/luckyclaw && /etc/init.d/S99luckyclaw restart && sleep 2 && luckyclaw version"
```

### Test on Device
```bash
sshpass -p 'luckfox' ssh root@<IP>
luckyclaw status      # Check everything
luckyclaw gateway -b  # Start in background
luckyclaw stop        # Stop cleanly
```

### SDK Overlay (for firmware builds)
Keep these directories in sync:
- `firmware/overlay/` — canonical overlay files
- `luckfox-pico-sdk/project/cfg/BoardConfig_IPC/overlay/luckyclaw-overlay/` — SDK build overlay

## File Map

| File | Purpose |
|------|---------|
| `cmd/luckyclaw/main.go` | CLI entry, onboarding, gateway, stop/restart |
| `pkg/channels/telegram.go` | Telegram bot (telego, long polling) |
| `pkg/cron/service.go` | Cron/reminders (be careful with empty JSON) |
| `pkg/config/config.go` | Config structure and defaults |
| `firmware/overlay/etc/profile.d/luckyclaw-banner.sh` | SSH login banner |
| `firmware/overlay/etc/init.d/S99luckyclaw` | Init script (auto-start) |
| `CULLED.md` | What changed from PicoClaw and why |
| `workspace/` | Embedded workspace templates |
