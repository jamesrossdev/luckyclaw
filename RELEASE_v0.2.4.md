# v0.2.4 — Storage Fix & Static IP Hardening

This release fixes a critical storage issue where the `/oem` partition (20MB) would fill up, preventing workspace files from being saved. It also hardens the `set-ip` command with validation and confirmation prompts.

## What's New

### Storage Fix: Workspace Moved to Rootfs

**The Problem:** The `/oem` partition on Luckfox boards is only ~20MB usable. After vendor firmware takes most of that space, there was only a few MB left for workspace data (sessions, WhatsApp DB, skills). Users would hit "NOT ENOUGH SPACE" errors when saving files.

**The Fix:** On embedded boards (Luckfox Pico), the workspace now lives on **rootfs** instead of `/oem`. Rootfs has ~143MB free on most boards.

**New workspace path:** `/root/.luckyclaw/workspace/`

**What this means:**
- Workspace data (sessions, WhatsApp DB, memory, skills) now uses the large rootfs partition
- Config file and heartbeat log stay on `/oem` (they're tiny)
- This matches how normal PC apps share the same disk space

**For existing users:** If you had a workspace at `/oem/.luckyclaw/workspace/`, you may need to manually migrate your data to `/root/.luckyclaw/workspace/`. The gateway will log a migration notice if it detects this situation.

### `luckyclaw set-ip` — IPv4 Validation

Before applying any static IP, the command now validates:
- All four octets are in the 0-255 range
- Target IP is not the same as the gateway
- Target IP is on the same subnet as the detected gateway

Invalid inputs are rejected with a clear error message.

### Confirmation Prompts

Both `set-ip <IP>` and `set-ip --dhcp` now ask for confirmation before rebooting. This prevents accidental network changes when mistyping an IP address.

## Bug Fixes

- Fixed `strconv.Atoi` error discard in IP parsing that allowed invalid octets (e.g., `192.168.1.foo`) to pass silently as `192.168.1.0`

## Migration

### Existing Workspace Data

If you upgraded from v0.2.3 or earlier and had workspace data at `/oem/.luckyclaw/workspace/`:

1. The gateway will log a notice if it detects both old and new workspaces exist
2. To migrate manually:
   ```bash
   # Stop the gateway first
   luckyclaw stop

   # Move your data
   mv /oem/.luckyclaw/workspace /root/.luckyclaw/workspace

   # Restart
   luckyclaw gateway -b
   ```

### Config File

No changes required. Config file location unchanged at `/oem/.luckyclaw/config.json`.

## Supported Boards

| Board | RAM | GOMEMLIMIT | Image |
|-------|-----|------------|-------|
| Luckfox Pico Plus | 64MB DDR2 | 24MiB | `luckyclaw-luckfox_pico_plus_rv1103-v0.2.4.img` |
| Luckfox Pico Pro | 128MB DDR3 | 48MiB | `luckyclaw-luckfox_pico_pro_rv1106-v0.2.4.img` |
| Luckfox Pico Max | 256MB DDR3 | 96MiB | `luckyclaw-luckfox_pico_max_rv1106-v0.2.4.img` |

## Downloads

All files needed to flash are attached below. For a quick upgrade on an existing board:

```bash
# Kill running process
killall -9 luckyclaw

# Upload new binary to /usr/bin/luckyclaw
# (SCP or use your preferred method)

# Restart
chmod +x /usr/bin/luckyclaw
/etc/init.d/S99luckyclaw start
```

## What's Changed from v0.2.3

| Feature | v0.2.3 | v0.2.4 |
|---------|---------|---------|
| Workspace location | `/oem/.luckyclaw/workspace/` | `/root/.luckyclaw/workspace/` |
| Config location | `/oem/.luckyclaw/config.json` | Unchanged |
| Heartbeat log | `/oem/.luckyclaw/heartbeat.log` | Unchanged |
| `set-ip <IP>` | Basic static IP | IPv4 validation + confirmation |
| `set-ip --dhcp` | Direct restore | Confirmation prompt before restore |
| Invalid octets (e.g., `.foo`) | Silently accepted as `0` | Rejected with error |
| Gateway collision check | Not checked | Checked and rejected |
