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
- 🧠 **Memory-optimized** — Tuned specifically for 64MB RAM boards, not general-purpose servers
- 📟 **SSH banner** — See gateway status and commands on every login
- 🌍 **Timezone-aware** — Correct local time on the board, no `/usr/share/zoneinfo` needed
- 📎 **File attachments** — Send files via Telegram, WhatsApp, or Discord
- 🤙 **Conservative by design** — Fewer features, fewer surprises, fewer crashes

> LuckyClaw is built on top of [PicoClaw](https://github.com/sipeed/picoclaw) by [Sipeed](https://sipeed.com). PicoClaw is the upstream project — LuckyClaw is the simpler, more opinionated fork optimized for Luckfox hardware and everyday users. We cherry-pick stability fixes and genuinely useful features from upstream; we don't try to keep pace with every new addition.

---

## Quick Start (End Users)

### Option A: Flash Pre-Built Firmware (Recommended for Beginners)

Download the firmware image for your board from [GitHub Releases](https://github.com/jamesrossdev/luckyclaw/releases) and follow the [LuckyClaw Flashing Guide](docs/FLASHING_GUIDE.md).

After flashing, connect via SSH and run:
```bash
luckyclaw onboard
```

### Option B: Binary Install on Existing Luckfox (No Reflash)

If your board is already running Luckfox Buildroot, you can install LuckyClaw directly:

```bash
# Download the ARMv7 binary on your computer, then upload it to the board:
# (use your board's IP instead of <DEVICE_IP>)
scp luckyclaw-linux-arm root@<DEVICE_IP>:/usr/bin/luckyclaw

# Set permissions
ssh root@<DEVICE_IP> "chmod +x /usr/bin/luckyclaw"

# Run onboard setup
ssh root@<DEVICE_IP> "luckyclaw onboard"

# Start in background
ssh root@<DEVICE_IP> "luckyclaw gateway -b"
```


| Board | Chip | Image |
|-------|------|-------|
| **Luckfox Pico Plus** | RV1103 | `luckyclaw-luckfox_pico_plus_rv1103-vX.X.X.img` |
| **Luckfox Pico Pro** | RV1106 | `luckyclaw-luckfox_pico_pro_max_rv1106-vX.X.X.img`* |
| **Luckfox Pico Max** | RV1106 | `luckyclaw-luckfox_pico_pro_max_rv1106-vX.X.X.img`* |

\* *The Pico Pro (128MB RAM) and Pico Max (256MB RAM) share the same RV1106 SoC and firmware image.*

> LuckyClaw currently only supports these three board variants. Other Luckfox variants (Pico Mini, Pico Zero, etc.) are untested and may not work.

---
### 1. Flash the firmware

Download the firmware image matching your board from [GitHub Releases](https://github.com/jamesrossdev/luckyclaw/releases).

Follow our detailed documentation to flash the firmware:

👉 **[LuckyClaw Flashing Guide (eMMC)](docs/FLASHING_GUIDE.md)**

### 2. Connect via SSH

```bash
ssh root@<IP>
# Default password: luckfox
```
 
> The device IP depends on your network. Connect the board via USB-C or Ethernet and check your router's DHCP leases, or use `arp -a | grep luckfox` to find it.

### Setting a Static IP

By default, the board obtains an IP address via DHCP. To set a static IP:

**Option A: Use `luckyclaw set-ip` (Recommended)**

```bash
luckyclaw set-ip 192.168.1.100
```

The command automatically detects your gateway and subnet, then reboots to apply the new IP.

**Option B: Manual Configuration**

```bash
nano /etc/network/interfaces
# Replace with your desired IP:
# auto eth0
# iface eth0 inet static
#     address 192.168.1.100
#     netmask 255.255.255.0
#     gateway 192.168.1.1
reboot
```

> To restore DHCP: `luckyclaw set-ip --dhcp`

For persistent configuration across reboots, configure DHCP reservation on your router using the board's MAC address.

You'll see the LuckyClaw banner:

```
 _               _           ____ _
| |   _   _  ___| | ___   _ / ___| | __ ___      __
| |  | | | |/ __| |/ / | | | |   | |/ _` \ \ /\ / /
| |__| |_| | (__|   <| |_| | |___| | (_| |\ V  V /
|_____\__,_|\___|_|\_\\__, |\____|_|\__,_| \_/\_/
                      |___/
  🦞 luckyclaw v0.2.4

  Board:     Pico Plus
  Memory:    33MB available / 55MB total
  Gateway:   running (PID 1234, 15MB RSS)
  MemLimit:  24MiB

  Commands:
    luckyclaw status      — System status
    luckyclaw onboard     — Setup wizard
    luckyclaw gateway      — Start AI gateway
    luckyclaw gateway -b  — Start in background
    luckyclaw stop        — Stop gateway
    luckyclaw restart     — Restart gateway
    luckyclaw set-ip     — Set static IP
    luckyclaw help        — View more commands
```

### 3. Run the setup wizard

```bash
luckyclaw onboard
```

If LuckyClaw detects an existing setup, the wizard now asks you to choose:

1. **Fresh onboard (recommended)** — wipes workspace and starts clean
2. **Keep existing files** — keeps current workspace and updates config

To force a clean reset directly:

```bash
luckyclaw onboard --wipe-workspace
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
| **WhatsApp** | ✅ Ready             | Use `luckyclaw onboard` to scan QR |

Plus other channels inherited from upstream (LINE, QQ, DingTalk, Feishu, MaixCam).

<details>
<summary><b>Telegram Setup</b></summary>

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
<summary><b>WhatsApp Setup</b></summary>

1. Run `luckyclaw onboard` and select WhatsApp
2. Scan the QR code with your WhatsApp app
3. For advanced features (Business Mode), see [WhatsApp Business Guide](docs/WHATSAPP_BUSINESS.md)

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
5. For advanced setup (moderator persona, sandbox), see [Discord Moderator Guide](docs/DISCORD_MODERATOR.md)

</details>

---

## 📋 CLI Reference

| Command                     | Description                     |
| --------------------------- | ------------------------------- |
| `luckyclaw onboard`         | Interactive setup wizard        |
| `luckyclaw onboard --wipe-workspace` | Force fresh onboard (wipes workspace) |
| `luckyclaw config-reset`   | Delete config.json (keeps workspace, needs re-onboard) |
| `luckyclaw status`          | System status (board, memory, gateway) |
| `luckyclaw gateway`         | Start the AI gateway            |
| `luckyclaw gateway -b`      | Start the AI gateway in background |
| `luckyclaw stop`            | Stop the gateway                |
| `luckyclaw restart`         | Restart the gateway             |
| `luckyclaw agent -m "..."`  | Send a message directly         |
| `luckyclaw agent`           | Interactive chat mode           |
| `luckyclaw cron list`       | List scheduled reminders        |
| `luckyclaw skills list`     | List installed skills           |
| `luckyclaw install`         | Install as system service       |
| `luckyclaw set-ip <IP>`     | Set static IP (auto-detects gateway/subnet) |
| `luckyclaw set-ip --dhcp`   | Restore DHCP (automatic IP)     |
| `luckyclaw version`         | Show version info               |

---

## ⚙️ Configuration

Config: `/oem/.luckyclaw/config.json`

### Providers

| Provider       | Purpose                    | Get API Key                                            |
| -------------- | -------------------------- | ------------------------------------------------------ |
| `openrouter`   | Many models (recommended)  | [openrouter.ai/keys](https://openrouter.ai/keys)      |
| `openai`       | GPT models                 | [platform.openai.com](https://platform.openai.com)     |
| `anthropic`    | Claude models              | [console.anthropic.com](https://console.anthropic.com) |
| `gemini`       | Google Gemini              | [aistudio.google.com](https://aistudio.google.com)     |
| `groq`         | Fast inference + voice     | [console.groq.com](https://console.groq.com)           |
| `ollama`       | Local models (no API key)  | [ollama.com](https://ollama.com)                       |

### Default Configuration

| Setting            | Default Value                  |
|--------------------|-------------------------------|
| Provider           | `openrouter`                 |
| Model              | `stepfun/step-3.5-flash:free` |
| Max Tokens         | `auto-clamped to 20% of context_window, max 16384` |
| Allow Unsafe Max Tokens | `false` (clamp enabled) |
| Context Window     | Model-specific (queried via API) |
| Temperature        | `0.6`                         |
| Max Tool Iterations| `25`                          |

> **Max Tokens Safety:** On startup (and during onboarding), `max_tokens` is automatically clamped to `min(20% of context_window, 16384, provider_max_output)` with a floor of 1024. This prevents context-window overflow errors on models like DeepSeek v3.2 while preserving usable output sizes. Existing configs are auto-healed on gateway start.
>
> To disable clamping and use a custom `max_tokens` value exactly as set, add `"allow_unsafe_max_tokens": true` to your `config.json` under `agents.defaults`. This opt-out is intended for advanced users who want maximum output size at the risk of overflow errors.

### Workspace Layout

On-device path: `/root/.luckyclaw/workspace/`

```
/root/.luckyclaw/workspace/
├── sessions/          # Conversation history
├── memory/            # Long-term memory (MEMORY.md)
├── cron/              # Scheduled jobs
├── skills/            # Custom skills
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

Jobs are stored in `/root/.luckyclaw/workspace/cron/` and persist across restarts.

---

## 🔒 Security

LuckyClaw runs in a sandboxed workspace by default. The agent can only access files within `/root/.luckyclaw/workspace/`.

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

For instructions on compiling from source, cross-compiling, and optimizing the Luckfox SDK root filesystem (including our automated `optimize-rootfs.patch`), please see the **[Developer Guide](docs/DEVELOPER_GUIDE.md)**!

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

## ⚡ Quick Navigation (Terminal Shortcuts)

| Destination | Path / Command |
|-------------|----------------|
| **Config File** | `nano /oem/.luckyclaw/config.json` |
| **Workspace** | `/root/.luckyclaw/workspace/` |
| **Skills Dir** | `/root/.luckyclaw/workspace/skills/` |
| **Logs** | `tail -f /var/log/luckyclaw.log` |
| **Gateway Status** | `luckyclaw status` |

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
