# 🦞 LuckyClaw: Differences from PicoClaw

LuckyClaw is a fork of [PicoClaw](https://github.com/sipeed/picoclaw) optimized for Luckfox Pico boards. This document lists what's been changed or tuned for the constrained hardware (64MB DDR2, ~33MB usable).

## Performance Tuning (Baked In)

| Setting | PicoClaw Default | LuckyClaw Default | Why |
|---------|-----------------|-------------------|-----|
| `GOGC` | 100 (Go default) | **20** | Aggressive GC to keep RSS low |
| `GOMEMLIMIT` | unlimited | **8 MiB** | Hard cap for 64MB boards |
| `max_tokens` | 8192 | **4096** | Reduces memory spikes during LLM response parsing |
| `max_tool_iterations` | 20 | **10** | Fewer iterations = less memory per conversation |

These are set automatically at startup. Environment variables can still override them.

## Onboarding Simplified

- **PicoClaw**: Offers 7 provider choices (OpenRouter, OpenAI, Anthropic, Groq, Gemini, Ollama, Skip)
- **LuckyClaw**: OpenRouter only — one key, access to all models, simplest setup for new users

Other providers (OpenAI, Anthropic, Groq, Gemini, Ollama) still work — edit `~/.luckyclaw/config.json` directly to use them.

## Added Features

- **SSH banner** (`/etc/profile.d/luckyclaw-banner.sh`) — ASCII art + gateway status on login
- **`luckyclaw stop`** — Stop the gateway cleanly
- **`luckyclaw restart`** — Restart gateway in background
- **`luckyclaw gateway -b`** — Start gateway in background, return to shell
- **Enhanced `luckyclaw status`** — Shows board model, memory, PID, RSS, channels, providers
- **Auto-start on boot** — Init script starts gateway if config exists

## What's Preserved

All PicoClaw channel and device connectivity is maintained:

- ✅ Telegram, Discord, QQ, DingTalk, LINE, Slack, WhatsApp, OneBot, Feishu
- ✅ Heartbeat (periodic tasks)
- ✅ Cron/reminders
- ✅ Skills system
- ✅ Voice transcription (Groq Whisper)
- ✅ Web search (Brave, DuckDuckGo)
- ✅ All agent tools (exec, read_file, write_file, web, etc.)

Nothing has been removed from the codebase — only defaults have been tuned for memory-constrained devices.
