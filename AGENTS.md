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
workspace/                ← Workspace templates embedded into binary via go:embed
firmware/overlay/etc/     ← Init script + SSH banner baked into firmware rootfs.img
```

## Critical Hardware Constraints

| Constraint | Value | Impact |
|-----------|-------|--------|
| Total RAM | 64MB DDR2 | Only ~33MB usable after kernel |
| Usable RAM | ~33MB | Gateway uses 10-14MB, leaves 2-12MB free |
| GOMEMLIMIT | 24MiB | Set in init script; binary default is 8MiB but overridden |
| GOGC | 20 | Aggressive GC to prevent RSS growth |
| Flash | 128MB SPI NAND | Limited storage, no swap |
| CPU | ARM Cortex-A7 (RV1103) | Single core, GOARM=7 |

## Changes from PicoClaw

### What We Changed
- **Onboarding**: Simplified to OpenRouter only (was 7 provider choices)
- **Performance**: Baked `GOGC=20` into binary; init script sets `GOMEMLIMIT=24MiB`
- **CLI**: Added `luckyclaw stop`, `restart`, `gateway -b` (background)
- **Init script**: Auto-starts gateway on boot with OOM protection
- **SSH banner**: Shows ASCII art, status, memory, all commands on login
- **Default model**: `stepfun/step-3.5-flash:free` (free tier)
- **Defaults**: `max_tokens=16384`, `max_tool_iterations=25` (tuned for web search headroom)

### What We Did NOT Change
All PicoClaw channels (Telegram, Discord, QQ, LINE, Slack, WhatsApp, etc.) and tools remain in the codebase. Users can configure any provider via `config.json` directly.

## Lessons Learned (Read These!)

### 1. OOM Killer Targets Go Binaries
Go allocates ~500MB virtual memory (lazy reservations). The Linux OOM killer uses `total-vm` in its scoring, so it kills luckyclaw first even at 10MB RSS. **Fix**: Set `oom_score_adj=-200` in the init script after starting the daemon.

### 2. Cron Service Crashes on Empty JSON
The `loadStore()` function in `pkg/cron/service.go` panicked on empty or corrupted `jobs.json`. **Fix**: Added graceful handling — treats empty/corrupt files as fresh state.

### 3. Init Script Must Bake Environment Variables
The init script at `/etc/init.d/S99luckyclaw` MUST export `GOGC=20`, `GOMEMLIMIT=24MiB`, and `TZ` before starting the daemon. Without these, the binary runs with Go defaults and immediately OOMs. WARNING: setting GOMEMLIMIT too low (e.g. 8MiB) causes the GC to spin at 100% CPU.

### 4. Busybox Limitations
Luckfox uses Busybox. `wget` doesn't support HTTPS. `sudo` doesn't exist—you're already root. `curl` isn't available. The Go binary handles all HTTPS via `net/http`.

### 5. DNS Resolution Workaround
Telegram API DNS (`api.telegram.org`) sometimes fails to resolve. The init script adds a static entry to `/etc/hosts`.

### 6. Don't Add Unnecessary Dependencies
Every byte counts. The binary is already ~15MB stripped. Adding dependencies increases memory usage. Always test with `GOMEMLIMIT=24MiB`.

### 7. AI Agent Access to the Device
If you are an AI agent and need to test changes, examine logs, or execute commands directly on the Luckfox Pico hardware, **do not guess the IP or password**. Simply ask the user to provide the SSH IP address and password for the device, and use the `run_command` tool via `sshpass` (e.g., `sshpass -p <password> ssh root@<ip>`).

### 8. Committing and Pushing Code
If you are an AI agent, you **MUST NEVER** commit or push code without explicit permission from the user. When you are asked to commit, you must ensure that the tracked `firmware/overlay` directory is completely up to date with whatever modifications were made inside the untracked `luckfox-pico-sdk` directory. This is the only way secondary developers receive OS-level modifications.

### 9. Execution Requires Approved Implementation Plan
If you are an AI agent, you **MUST NEVER** execute code changes, environment modifications, or configuration adjustments without explicitly drafting an implementation plan and receiving the user's explicit approval first. Do not make unauthorized technical assumptions.

### 10. Multiple Daemon Instances & PID Tracking
If `luckyclaw gateway -b` is executed while a daemon started by `/etc/init.d/S99luckyclaw` is already running it will overwrite the `/var/run/luckyclaw.pid` file. Because the init script only tracks the latest PID, subsequent `stop` or `restart` commands will leave the original daemon alive as a zombie, causing duplicate Telegram processing and hallucinated timestamps in session memory. **Fix:** Going forward, making sure we strictly append `&& killall -9 luckyclaw` alongside the init script (which I've started doing in my deploy commands) completely eliminates the possibility of this happening again.

### 11. PicoClaw Upstream Reference
A shallow clone of the upstream PicoClaw repo is kept at `picoclaw-latest/` (gitignored). This is used for comparing upstream changes and evaluating code worth porting. To refresh it: `cd picoclaw-latest && git pull`. Do not commit this directory.

### 12. Log File Destinations & Workspace Paths
- **Gateway log**: `/var/log/luckyclaw.log` (stdout/stderr from the init script). The init script uses an `sh -c "exec ..."` wrapper because BusyBox's `start-stop-daemon -b` redirects fds to `/dev/null` before shell redirects take effect.
- **Heartbeat log**: `<workspace>/heartbeat.log` (written directly by the heartbeat service, not stdout).
- **Runtime workspace**: `/oem/.luckyclaw/workspace/` — this is where the bot reads/writes data at runtime. `luckyclaw onboard` creates it by extracting the `workspace/` directory that is **embedded directly into the binary** via `go:embed` at compile time. `firmware/overlay/root/` is NOT involved in this — nothing reads `/root/.luckyclaw/` at runtime.

### 13. Firmware Overlay Structure
Only two parts of `firmware/overlay/` are meaningful:
- `firmware/overlay/etc/` — init script, SSH banner, timezone. **Must be tracked in git.** Gets baked into `rootfs.img`.
- `firmware/overlay/root/` — **Dead weight. Do not use.** Nothing reads `/root/.luckyclaw/` at runtime; workspace data comes from the binary embed.
- `firmware/overlay/usr/` — **Not tracked in git.** The ARM binary is compiled at SDK build time and placed here; it is not stored in the repo.

### 14. Binary-Only Updates (No Reflash Required)
The binary at `/usr/bin/luckyclaw` lives on the writable `rootfs` partition and can be replaced via SCP at any time without reflashing the firmware. This is how all development deploys work. Because `workspace/` is embedded in the binary, updating the binary also delivers new/updated skills and templates to users when they next run `luckyclaw onboard`. This architecture makes **over-the-air (OTA) auto-update** possible: the binary could check GitHub Releases, download a new ARM build, kill itself, overwrite `/usr/bin/luckyclaw`, and restart via the init script. 

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
  go build -ldflags "-s -w -X main.version=0.2.x" \
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

### Build Distributable Firmware Image

A distributable `.img` bundles the ARM binary (with `workspace/` embedded) + the init script + SSH banner into a single flashable file. Steps:

```bash
# 1. Build ARM binary (go:embed bakes workspace/ into it automatically)
make build-arm
# Output: build/luckyclaw-linux-arm

# 2. Copy binary into the SDK overlay (untracked — do this every time before building image)
cp build/luckyclaw-linux-arm \
  luckfox-pico-sdk/project/cfg/BoardConfig_IPC/overlay/luckyclaw-overlay/usr/bin/luckyclaw
chmod +x luckfox-pico-sdk/project/cfg/BoardConfig_IPC/overlay/luckyclaw-overlay/usr/bin/luckyclaw

# 3. Build the firmware image
cd luckfox-pico-sdk && ./build.sh

# 4. Output image is at:
# luckfox-pico-sdk/IMAGE/<timestamp>/IMAGES/update.img
# Rename for distribution: luckyclaw-luckfox_pico_plus_rv1103-vX.Y.Z.img
```

> **Note:** The SDK overlay `etc/` is kept in sync with `firmware/overlay/etc/` in the repo. If you modify the init script or SSH banner, copy the changes to both locations before building.

> **What's in the image:** `update.img` = kernel + rootfs (containing `/usr/bin/luckyclaw` with embedded workspace) + oem partition. When a user runs `luckyclaw onboard` after flashing, the embedded workspace is extracted to `/oem/.luckyclaw/workspace/`.

### SDK Overlay Sync
The SDK overlay `etc/` must stay in sync with the repo:
- `firmware/overlay/etc/` — canonical, tracked in git
- `luckfox-pico-sdk/project/cfg/BoardConfig_IPC/overlay/luckyclaw-overlay/etc/` — SDK copy, NOT tracked in git

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
