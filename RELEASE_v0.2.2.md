**Title:** `v0.2.2 — Native WhatsApp & Embedded Reliability`

---

This release introduces native WhatsApp messaging with QR code pairing, dynamic hardware detection for all Luckfox Pico variants, and critical reliability improvements for 24/7 embedded operation.

## What's New

*   **Native WhatsApp Channel**: Full WhatsApp integration via whatsmeow with QR code pairing during onboarding. Supports direct messages, group chats, quoted replies, media (images/audio/video), and business mode whitelisting.
    *   Interactive onboarding: scan QR code or enter phone number
    *   Quoted replies appear as proper WhatsApp quote blocks
    *   Offline detection with user-friendly error messages
    *   Self-chat loop prevention and deduplication

*   **Dynamic Board Detection**: Board variant (Plus/Pro/Max) is now detected by total RAM rather than unreliable device tree strings. GOMEMLIMIT and board name are set automatically:
    *   Pico Plus (≤60MB): `GOMEMLIMIT=24MiB`
    *   Pico Pro (61-200MB): `GOMEMLIMIT=48MiB`
    *   Pico Max (>200MB): `GOMEMLIMIT=96MiB`

*   **Safe Gateway Startup After Onboarding**: Onboarding now starts the gateway via the init script (`S99luckyclaw start`) instead of spawning directly, ensuring GOGC, GOMEMLIMIT, and TZ environment variables are always set correctly. No more running without memory limits.

*   **Init Script Timezone**: Timezone is read from `config.json` instead of being hardcoded. Falls back to UTC if config is missing.

*   **Cron/Reminder Tool**: New `cron` tool for scheduling reminders (one-shot and repeating). Supports `at_seconds`, `at_time` (ISO-8601), and `at_time` natural language formats.

*   **Heartbeat Self-Chat Alerts**: Heartbeat alerts are now routed to WhatsApp self-chat, providing real-time notifications when the device has issues.

*   **Session Binary Guard & Dynamic Context Window**: Binary file uploads are rejected before they enter session history. Context window is dynamically queried from OpenRouter during onboarding per model.

## Reliability Improvements

*   **OOM Protection**: `GOGC=20` and dynamic `GOMEMLIMIT` baked into init script. `oom_score_adj=-200` set after startup to prevent Linux OOM killer from targeting LuckyClaw first.
*   **SSH Banner**: Login banner now shows board name, memory stats, gateway PID, RSS, and GOMEMLIMIT — all based on stable MemTotal detection.
*   **NTP Wait**: Gateway won't start until system clock is synced (prevents TLS failures on cold boot).
*   **ADB Daemon Kill**: `adbd` is stopped during init to reclaim ~1.5MB RAM.
*   **Telegram DNS Workaround**: Static `api.telegram.org` entry added to `/etc/hosts`.
*   **Process-Safe Stop/Restart**: Init script's `stop` command no longer kills `luckyclaw onboard` or other user-facing commands.

## Supported Boards
| Board | RAM | GOMEMLIMIT | Image |
|-------|-----|------------|-------|
| Luckfox Pico Plus | 64MB DDR2 | 24MiB | `luckyclaw-luckfox_pico_plus_rv1103-v0.2.2.img` |
| Luckfox Pico Pro | 128MB DDR3 | 48MiB | `luckyclaw-luckfox_pico_pro_rv1106-v0.2.2.img` |
| Luckfox Pico Max | 256MB DDR3 | 96MiB | `luckyclaw-luckfox_pico_max_rv1106-v0.2.2.img` |

## Downloads

All files needed to flash are attached below. For a fresh install, flash the `update.img` or simply replace the binary on your existing board:

```bash
# Quick upgrade (on device)
killall -9 luckyclaw
# [Upload new binary to /usr/bin/luckyclaw]
/etc/init.d/S99luckyclaw start
```
