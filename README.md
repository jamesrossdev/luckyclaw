<div align="center">
  <img src="assets/logo.png" alt="LuckyClaw" width="512">

  <h1>🦞 LuckyClaw: AI Assistant for Luckfox Pico</h1>

  <h3>The streamlined AI companion for Luckfox hardware.</h3>

  <p>
    <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go">
    <img src="https://img.shields.io/badge/Board-Luckfox_Pico-orange" alt="Board">
    <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
    <a href="https://discord.gg/TRdD9dBe"><img src="https://img.shields.io/badge/Discord-Community-5865F2?logo=discord&logoColor=white" alt="Discord"></a>
  </p>
</div>

---

LuckyClaw is a streamlined, self-contained AI assistant purpose-built for the [Luckfox Pico](https://wiki.luckfox.com/Luckfox-Pico/Luckfox-Pico-quick-start/) ecosystem. While based on the excellent work of [PicoClaw](https://github.com/sipeed/picoclaw), LuckyClaw prioritizes absolute stability and ease of use for everyday users over the complex feature velocity of its upstream counterpart.

**Who it's for:** LuckyClaw is designed for those who want a reliable, 24/7 digital companion on Telegram or Discord without the overhead of manual compilation, complex configurations, or dedicated server maintenance. If you have a Luckfox board, you have a professional-grade AI assistant.

**What makes it different:**

- 🔧 **Pre-built firmware** — Flash and go, no SDK or compilation required
- 🧙 **Interactive onboarding** — `luckyclaw onboard` walks you through everything in 2 minutes
- 🧠 **Memory-optimized** — Tuned specifically for 64MB boards, not general-purpose servers
- 📟 **SSH banner** — See gateway status and commands on every login
- 🌍 **Timezone-aware** — Correct local time on the board, no `/usr/share/zoneinfo` needed
- 📎 **File attachments** — Send files directly via Telegram
- 🤙 **Conservative by design** — Fewer features, fewer surprises, fewer crashes

> [!NOTE]
> LuckyClaw is built on top of [PicoClaw](https://github.com/sipeed/picoclaw) by [Sipeed](https://sipeed.com). PicoClaw is the upstream project — LuckyClaw is the simpler, more opinionated fork optimized for Luckfox hardware and everyday users. We cherry-pick stability fixes and genuinely useful features from upstream; we don't try to keep pace with every new addition.

## ⚡ Quick Start (End Users)

### Option A: Flash Pre-Built Firmware (Recommended for Beginners)

Download the firmware image for your board from [GitHub Releases](https://github.com/jamesrossdev/luckyclaw/releases) and follow the [LuckyClaw Flashing Guide](doc/FLASHING_GUIDE.md).

After flashing, connect via SSH and run:
```bash
luckyclaw onboard
```

### Option B: Binary Install on Existing Luckfox (No Reflash)

If your board is already running Luckfox Buildroot, you can install LuckyClaw directly:

```bash
# Download the ARMv7 binary
wget https://github.com/jamesrossdev/luckyclaw/releases/latest/download/luckyclaw-linux-arm -O /usr/bin/luckyclaw
chmod +x /usr/bin/luckyclaw

# Run onboard setup
luckyclaw onboard

# Start in background
luckyclaw gateway -b
```


| Board | Chip | Image |
|-------|------|-------|
| **Luckfox Pico Plus** | RV1103 | `luckyclaw-luckfox_pico_plus_rv1103-vX.X.X.img` |
| **Luckfox Pico Pro** | RV1106 | `luckyclaw-luckfox_pico_pro_max_rv1106-vX.X.X.img`* |
| **Luckfox Pico Max** | RV1106 | `luckyclaw-luckfox_pico_pro_max_rv1106-vX.X.X.img`* |

\* *The Pico Pro (128MB RAM) and Pico Max (256MB RAM) share the same RV1106 SoC and firmware image.*

> [!IMPORTANT]
> LuckyClaw currently only supports these three board variants. Other Luckfox variants (Pico Mini, Pico Zero, etc.) are untested and may not work.

### 1. Flash the firmware

Download the firmware image matching your board from [GitHub Releases](https://github.com/jamesrossdev/luckyclaw/releases).

Follow our detailed documentation to flash the firmware:

👉 **[LuckyClaw Flashing Guide (eMMC)](doc/FLASHING_GUIDE.md)**

### 2. Connect via SSH

```bash
ssh root@<IP>
# Default password: luckfox
```

> [!TIP]
> The device IP depends on your network. Connect the board via USB-C or Ethernet and check your router's DHCP leases, or use `arp -a | grep luckfox` to find it.

You'll see the LuckyClaw banner:

```
 _               _           ____ _
| |   _   _  ___| | ___   _ / ___| | __ ___      __
| |  | | | |/ __| |/ / | | | |   | |/ _` \ \ /\ / /
| |__| |_| | (__|   <| |_| | |___| | (_| |\ V  V /
|_____\__,_|\___|_|\_\\__, |\____|_|\__,_| \_/\_/
                      |___/
  🦞 luckyclaw v0.2.0

  Gateway: running (PID 1234, 15MB)
  Memory:  33MB / 55MB available

  Commands:
    luckyclaw status      — System status
    luckyclaw onboard     — Setup wizard
    luckyclaw gateway     — Start AI gateway
    luckyclaw gateway -b  — Start in background
    luckyclaw stop        — Stop gateway
    luckyclaw restart     — Restart gateway
```

### 3. Run the setup wizard

```bash
luckyclaw onboard
```

The wizard walks you through:

1. **API Provider** — OpenRouter - but you can manually set up OpenAI, Anthropic, Ollama and others in config.json
2. **API Key** — Paste your key, it's validated in real-time
3. **Timezone** — Explicitly enter your IANA Zone classification via the [Wikipedia TZ Database List](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones#List) 
4. **Messaging** — Optionally set up Telegram and/or Discord
5. **Start gateway** — Optionally start the AI gateway in the background

### 4. Chat!

```bash
# Direct message
luckyclaw agent -m "What time is it?"

# Interactive mode
luckyclaw agent

# Or use Telegram/Discord (if configured)
# Just message your bot!
```

---

## 💬 Chat Channels

| Channel      | Status              | Setup                      |
| ------------ | ------------------- | -------------------------- |
| **Telegram** | ✅ Ready             | Token from @BotFather      |
| **Discord**  | ✅ Ready             | Bot token + intents        |
| **WhatsApp** | 🚧 Work in Progress | —                          |
| **Slack**    | 🧬 Inherited (untested) | —                       |

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

- Go 1.25+
- [Luckfox Pico SDK](https://github.com/LuckfoxTECH/luckfox-pico) (for firmware builds)
- ARM cross-compilation toolchain (included in the SDK)

### Build from source

```bash
git clone https://github.com/jamesrossdev/luckyclaw.git
cd luckyclaw

# Build for your local machine
make build

# Cross-compile for Luckfox Pico (ARMv7)
make build-arm
```

### Development Workflow

Keep the codebase clean using the integrated Makefile targets:

- `make fmt` — Format Go code
- `make vet` — Run static analysis
- `make test` — Run unit tests
- `make check` — Run all of the above (recommended before committing)
- `make clean` — Remove build artifacts

### Build firmware image

The firmware overlay only contains OS-level files that get baked into `rootfs.img`. The **workspace templates** (`SOUL.md`, skills, etc.) are **embedded directly into the binary** via `go:embed workspace` — so every binary already carries the full workspace inside it. Users get workspace files by running `luckyclaw onboard`, which extracts them to `/oem/.luckyclaw/workspace/`.

```
firmware/overlay/
└── etc/
    ├── init.d/S99luckyclaw          # Auto-start on boot
    ├── profile.d/luckyclaw-banner.sh # SSH login banner
    └── ssl/certs/ca-certificates.crt # TLS certificates
```

To build a distributable firmware image:

1. **Build the ARM binary** (workspace is embedded automatically):
   ```bash
   make build-arm
   # Output: build/luckyclaw-linux-arm
   ```

2. **Clone the SDK** (one-time setup):
   ```bash
   git clone https://github.com/LuckfoxTECH/luckfox-pico.git luckfox-pico-sdk
   ```

3. **Sync the `etc/` overlay to the SDK** (do this if init script or banner changed):
   ```bash
   cp -r firmware/overlay/etc/ \
     luckfox-pico-sdk/project/cfg/BoardConfig_IPC/overlay/luckyclaw-overlay/etc/
   ```

4. **Copy the ARM binary into the SDK overlay**:
   ```bash
   cp build/luckyclaw-linux-arm \
     luckfox-pico-sdk/project/cfg/BoardConfig_IPC/overlay/luckyclaw-overlay/usr/bin/luckyclaw
   chmod +x \
     luckfox-pico-sdk/project/cfg/BoardConfig_IPC/overlay/luckyclaw-overlay/usr/bin/luckyclaw
   ```

5. **Build the firmware image**:
   ```bash
   cd luckfox-pico-sdk && ./build.sh
   ```

6. **Find the output image**:
   ```
   luckfox-pico-sdk/IMAGE/<timestamp>/IMAGES/update.img
   ```
   Rename it: `luckyclaw-luckfox_pico_plus_rv110x-vX.Y.Z.img` depending on your board and version.

When a user flashes this image and runs `luckyclaw onboard`, the embedded workspace is extracted to `/oem/.luckyclaw/workspace/`.

### Project structure

```
luckyclaw/
├── cmd/luckyclaw/main.go    # Entry point, CLI, onboarding wizard (embeds workspace/)
├── pkg/
│   ├── agent/               # Core agent loop and context builder
│   ├── bus/                 # Internal message bus
│   ├── channels/            # Telegram, Discord, and other messaging integrations
│   ├── config/              # Configuration and system settings
│   ├── providers/           # LLM provider implementations (OpenRouter, etc.)
│   ├── skills/              # Skill loader and installer
│   ├── tools/               # Agent tools (shell, file, i2c, spi, send_file)
│   └── ...
├── firmware/overlay/etc/    # Init script + SSH banner baked into firmware image
├── workspace/               # Templates embedded into binary via go:embed
└── assets/                  # Documentation images and media
```



### Performance tuning

LuckyClaw automatically sets `GOGC=20` and `GOMEMLIMIT=24MiB` at startup for memory-constrained boards. These can be overridden via environment variables if your board has more RAM.

---

## 🐛 Troubleshooting

### Gateway keeps getting killed (OOM)

LuckyClaw v0.2+ automatically caps memory usage. If you're on an older version or running on a custom board, set:

```bash
export GOGC=20
export GOMEMLIMIT=24MiB
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

Look for `✓ Cron service started`. If missing, the jobs.json file may be corrupted — v0.2.0+ handles this automatically.

### Time is wrong

LuckyClaw v0.2.0+ embeds its own timezone database and sets the timezone during onboarding. If the time is still wrong:

1. **System clock**: Sync via NTP:
   ```bash
   ntpd -p pool.ntp.org && date
   ```

2. **Timezone**: Re-run onboarding or set manually:
   ```bash
   echo "export TZ='America/New_York'" > /etc/profile.d/timezone.sh
   ```
   Then restart the gateway: `luckyclaw restart`

---

## 💬 Community

Join our Discord for help, feedback, and discussion:

👉 **[LuckyClaw Discord](https://discord.gg/TRdD9dBe)**

## 🙏 Credits

- [PicoClaw](https://github.com/sipeed/picoclaw) by [Sipeed](https://sipeed.com) — The upstream project that LuckyClaw is forked from
- [nanobot](https://github.com/HKUDS/nanobot) — The original Python AI agent that inspired PicoClaw
- [Luckfox](https://wiki.luckfox.com/) — For making excellent, affordable Linux boards

## 📄 License

MIT — see [LICENSE](LICENSE)
