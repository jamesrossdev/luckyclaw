<div align="center">
  <img src="assets/logo.jpg" alt="LuckyClaw" width="512">

  <h1>🦞 LuckyClaw: AI Assistant for Luckfox Pico</h1>

  <h3>One-stop AI firmware for Luckfox Pico boards</h3>

  <p>
    <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go">
    <img src="https://img.shields.io/badge/Board-Luckfox_Pico-orange" alt="Board">
    <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
  </p>
</div>

---

LuckyClaw is a purpose-built AI assistant for [Luckfox Pico](https://wiki.luckfox.com/Luckfox-Pico/Luckfox-Pico-quick-start/) boards. It's a fork of [PicoClaw](https://github.com/sipeed/picoclaw), optimized specifically for Luckfox hardware with baked-in memory management, interactive setup, and pre-built firmware images.

**What makes it different from PicoClaw:**

- 🔧 **Pre-built firmware** — Flash and go, no SDK required for end-users
- 🧙 **Interactive onboarding** — `luckyclaw onboard` walks you through API key, model, timezone, and Telegram setup
- 🧠 **Memory-optimized** — GOGC and GOMEMLIMIT baked into the binary for 64MB boards
- 📟 **SSH banner** — See gateway status and available commands on login
- 🦞 **Board-aware** — Detects Luckfox Pico model, shows board-specific info in `status`

> [!NOTE]
> LuckyClaw is built on top of [PicoClaw](https://github.com/sipeed/picoclaw) by [Sipeed](https://sipeed.com). All credit for the core AI agent engine goes to the PicoClaw team and the original [nanobot](https://github.com/HKUDS/nanobot) project.

## ⚡ Quick Start (End Users)

### 1. Flash the firmware

Download the latest firmware image from [GitHub Releases](https://github.com/jamesrossdev/luckyclaw/releases).

Flash to your Luckfox Pico board using the [Luckfox flashing tool](https://wiki.luckfox.com/Luckfox-Pico/Luckfox-Pico-SD-Card-burn-image/):

```bash
# On Linux, use SocToolKit or dd
# On Windows, use the Luckfox burn tool
```

### 2. Connect via SSH

```bash
ssh root@192.168.1.175
# Password: luckfox
```

You'll see the LuckyClaw banner:

```
  _               _          ____ _
 | |   _   _  ___| | __   / ___| | __ ___      __
 | |  | | | |/ __| |/ / | |   | |/ _` \ \ /\ / /
 | |__| |_| | (__|   <| |_| |___| | (_| |\ V  V /
 |_____\__,_|\___|_|\_\__, \___|_|\__,_| \_/\_/
                       |___/
  🦞 luckyclaw v0.2.1

  Gateway: stopped
  Memory:  16MB / 32MB available

  Commands:
    luckyclaw status    — System status
    luckyclaw onboard   — Setup wizard
    luckyclaw gateway   — Start AI gateway
```

### 3. Run the setup wizard

```bash
luckyclaw onboard
```

The wizard walks you through:

1. **API Provider** — Choose OpenRouter (recommended), OpenAI, Anthropic, Groq, Gemini, or Ollama
2. **API Key** — Paste your key, it's validated in real-time
3. **Timezone** — Auto-detected via IP, or enter manually
4. **Messaging** — Optionally set up Telegram (Discord, WhatsApp, Slack coming soon)
5. **Start gateway** — Optionally start the AI gateway in the background

### 4. Chat!

```bash
# Direct message
luckyclaw agent -m "What time is it?"

# Interactive mode
luckyclaw agent

# Or use Telegram (if configured)
# Just message your bot!
```

---

## 💬 Chat Channels

| Channel      | Status     | Setup                      |
| ------------ | ---------- | -------------------------- |
| **Telegram** | ✅ Ready    | Token from @BotFather      |
| **Discord**  | ✅ Ready    | Bot token + intents        |
| **WhatsApp** | 🔜 Planned | —                          |
| **Slack**    | 🔜 Planned | —                          |

<details>
<summary><b>Telegram Setup</b> (Recommended)</summary>

1. Message `@BotFather` on Telegram → `/newbot` → copy the token
2. Get your user ID from `@userinfobot`
3. The onboarding wizard handles the rest, or edit `config.json`:

```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "token": "YOUR_BOT_TOKEN",
      "allow_from": ["YOUR_USER_ID"]
    }
  }
}
```

</details>

<details>
<summary><b>Discord Setup</b></summary>

1. Go to https://discord.com/developers/applications → Create app → Bot → Copy token
2. Enable **MESSAGE CONTENT INTENT** in Bot settings
3. Edit `~/.luckyclaw/config.json`:

```json
{
  "channels": {
    "discord": {
      "enabled": true,
      "token": "YOUR_BOT_TOKEN",
      "allow_from": ["YOUR_USER_ID"]
    }
  }
}
```

4. Invite: OAuth2 → `bot` scope → `Send Messages` + `Read Message History`

</details>

---

## 📋 CLI Reference

| Command                     | Description                     |
| --------------------------- | ------------------------------- |
| `luckyclaw onboard`         | Interactive setup wizard        |
| `luckyclaw status`          | System status (board, memory, gateway) |
| `luckyclaw gateway`         | Start the AI gateway            |
| `luckyclaw agent -m "..."`  | Send a message directly         |
| `luckyclaw agent`           | Interactive chat mode           |
| `luckyclaw cron list`       | List scheduled reminders        |
| `luckyclaw skills list`     | List installed skills           |
| `luckyclaw version`         | Show version info               |

---

## ⚙️ Configuration

Config: `~/.luckyclaw/config.json`

### Providers

| Provider       | Purpose                    | Get API Key                                            |
| -------------- | -------------------------- | ------------------------------------------------------ |
| `openrouter`   | Many models (recommended)  | [openrouter.ai/keys](https://openrouter.ai/keys)      |
| `openai`       | GPT models                 | [platform.openai.com](https://platform.openai.com)     |
| `anthropic`    | Claude models              | [console.anthropic.com](https://console.anthropic.com) |
| `gemini`       | Google Gemini              | [aistudio.google.com](https://aistudio.google.com)     |
| `groq`         | Fast inference + voice     | [console.groq.com](https://console.groq.com)           |
| `ollama`       | Local models (no API key)  | [ollama.com](https://ollama.com)                       |

### Workspace Layout

```
~/.luckyclaw/workspace/
├── sessions/          # Conversation history
├── memory/            # Long-term memory (MEMORY.md)
├── cron/              # Scheduled jobs
├── skills/            # Custom skills
├── AGENT.md           # Agent behavior guide
├── HEARTBEAT.md       # Periodic tasks (every 30 min)
├── IDENTITY.md        # Agent identity
└── USER.md            # User preferences
```

### Heartbeat (Periodic Tasks)

LuckyClaw checks `HEARTBEAT.md` every 30 minutes and runs the tasks listed there (e.g. check time, system health, network status).

```json
{
  "heartbeat": {
    "enabled": true,
    "interval": 30
  }
}
```

### Scheduled Reminders

Ask the agent to set reminders naturally:

- *"Remind me in 10 minutes to check dinner"*
- *"Remind me every 2 hours to drink water"*
- *"Set an alarm for 9am daily"*

Jobs are stored in `~/.luckyclaw/workspace/cron/` and persist across restarts.

---

## 🔒 Security

LuckyClaw runs in a sandboxed workspace by default. The agent can only access files within `~/.luckyclaw/workspace/`.

To allow system-wide access (use with caution):

```json
{
  "agents": {
    "defaults": {
      "restrict_to_workspace": false
    }
  }
}
```

---

## 🛠️ Developer Guide

### Prerequisites

- Go 1.21+
- [Luckfox Pico SDK](https://github.com/LuckfoxTECH/luckfox-pico) (for firmware builds)
- ARM cross-compilation toolchain (included in the SDK)

### Build from source

```bash
git clone https://github.com/jamesrossdev/luckyclaw.git
cd luckyclaw

# Build for your local machine
make build

# Cross-compile for Luckfox Pico (ARM)
GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build \
  -ldflags "-s -w -X main.version=$(git describe --tags --always)" \
  -o build/luckyclaw-linux-arm ./cmd/luckyclaw
```

### Build firmware image

The `firmware/` directory contains the SDK overlay files that get baked into the firmware image:

```
firmware/overlay/
├── etc/
│   ├── init.d/S99luckyclaw       # Auto-start on boot
│   ├── profile.d/luckyclaw-banner.sh  # SSH login banner
│   └── ssl/certs/ca-certificates.crt  # TLS certificates
├── root/.luckyclaw/
│   ├── config.json               # Default config
│   └── workspace/                # Default workspace files
└── usr/local/bin/luckyclaw       # The binary
```

To build a firmware image:

```bash
# 1. Clone the Luckfox Pico SDK (if not already done)
git clone https://github.com/LuckfoxTECH/luckfox-pico.git luckfox-pico-sdk

# 2. Copy the overlay into the SDK
cp -r firmware/overlay/* luckfox-pico-sdk/project/cfg/BoardConfig_IPC/overlay/luckyclaw-overlay/

# 3. Copy the ARM binary
cp build/luckyclaw-linux-arm luckfox-pico-sdk/project/cfg/BoardConfig_IPC/overlay/luckyclaw-overlay/usr/local/bin/luckyclaw

# 4. Build the firmware (uses the SDK's build system)
cd luckfox-pico-sdk
./build.sh lunch   # Select your board config
./build.sh
```

The firmware image will be in `luckfox-pico-sdk/output/image/`.

### Project structure

```
luckyclaw/
├── cmd/luckyclaw/main.go    # Entry point, CLI commands, onboarding wizard
├── pkg/
│   ├── agent/               # AI agent loop
│   ├── channels/            # Telegram, Discord, etc.
│   ├── config/              # Configuration management
│   ├── cron/                # Scheduled tasks
│   ├── heartbeat/           # Periodic tasks
│   ├── providers/           # LLM providers (OpenRouter, OpenAI, etc.)
│   ├── tools/               # Agent tools (shell, file, i2c, spi)
│   └── ...
├── firmware/                # SDK overlay files for firmware builds
├── workspace/               # Default workspace templates
└── assets/                  # Images and media
```

### Performance tuning

LuckyClaw automatically sets `GOGC=20` and `GOMEMLIMIT=8MiB` at startup for memory-constrained boards. These can be overridden via environment variables if your board has more RAM.

---

## 🐛 Troubleshooting

### Gateway keeps getting killed (OOM)

LuckyClaw v0.2+ automatically caps memory usage. If you're on an older version, set:

```bash
export GOGC=20
export GOMEMLIMIT=8MiB
luckyclaw gateway
```

### Telegram bot says "terminated by other getUpdates"

Only one gateway instance can run at a time. Kill any existing process:

```bash
killall luckyclaw
luckyclaw gateway
```

### Reminders not firing

If reminders were created but never fire, the cron service may not have started. Check the log:

```bash
tail -20 /var/log/luckyclaw.log
```

Look for `✓ Cron service started`. If missing, the jobs.json file may be corrupted — v0.2.1+ handles this automatically.

### Time is wrong

LuckyClaw uses NTP for time sync. If the time is wrong:

```bash
ntpd -p pool.ntp.org
date
```

Set timezone during `luckyclaw onboard` or manually:

```bash
export TZ=Africa/Nairobi  # or your timezone
```

---

## 🙏 Credits

- [PicoClaw](https://github.com/sipeed/picoclaw) by [Sipeed](https://sipeed.com) — The upstream project that LuckyClaw is forked from
- [nanobot](https://github.com/HKUDS/nanobot) — The original Python AI agent that inspired PicoClaw
- [Luckfox](https://wiki.luckfox.com/) — For making excellent, affordable Linux boards

## 📄 License

MIT — see [LICENSE](LICENSE)
