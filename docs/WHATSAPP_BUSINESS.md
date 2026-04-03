# 💼 LuckyClaw: WhatsApp Business Guide

Because LuckyClaw leverages a native CGO-free WhatsApp Web pairing protocol (`whatsmeow`), it does not communicate with the official Meta Cloud API. This completely shields your business from Meta's tier-based per-conversation messaging costs, allowing you to run a highly capable virtual assistant on WhatsApp with zero subscription fees overhead.

## Activating Business Mode

To enable the secure, restricted "Business Mode," you must manually add the `business_mode` flag to your `config.json` file on the board.

1. Open the configuration file:
   ```bash
   nano /oem/.luckyclaw/config.json
   ```

2. Inside the `whatsapp` block, find `"business_mode": false` and change it to `true`:
   ```json
   "whatsapp": {
     "enabled": true,
     "business_mode": true,
     "session_path": "~/.luckyclaw/whatsapp.db"
   }
   ```

3. Press **CTRL + S** to save, and **CTRL + X** to exit.

4. Restart the gateway to apply changes:
   ```bash
   luckyclaw restart
   ```

## Configuring Business Constraints & Persona

You have absolute control over how the bot responds to customers. Because LuckyClaw runs entirely locally via embedded workspace templates, you can sculpt its persona directly on the board.

When you finish setting up LuckyClaw using `luckyclaw onboard`, the embedded workspace is cloned onto your device at `/oem/.luckyclaw/workspace/`. To customize your business agent, you only need to modify two core files:

### 1. `IDENTITY.md` (Name Only)
This file defines who the bot is. We recommend keeping this as lean as possible—**only change the name** to your business assistant's name. Do NOT add personality or behavioral instructions here; those belong in `SOUL.md`.

```bash
nano /oem/.luckyclaw/workspace/IDENTITY.md
```

**Template:**
```markdown
# Identity

## Name
[Your Business Name] Assistant

## Description
Official automated assistant for [Your Business Name].
```

### 2. `SOUL.md` (The Operating Rules)
This file acts as the boundary lines for the agent's behavior. We highly recommend using the strict template below to ensure the bot remains locked into a professional context and refuses to act outside its boundaries. Replace your `~/.luckyclaw/workspace/SOUL.md` with this template.

```bash
nano /oem/.luckyclaw/workspace/SOUL.md
```

**Template:**
```markdown
# Soul

I am LuckyClaw a monitored AI assistant on WhatsApp. 

My primary purpose is to rigidly focus on customer service and business inquiries.

## Operating Principles
- You have permission to READ files and LIST directories within your workspace to answer inquiries.
- You have permission to FETCH URLs and SEARCH the web to provide up-to-date information.
- NEVER run shell commands, execute scripts, or interact with the underlying operating system.
- NEVER write, edit, or delete local files.
- NEVER set timers, alarms, or reminders.
- NEVER claim abilities you do not have (e.g., booking appointments or taking payments, unless a specific tool is explicitly provided).
- NEVER roleplay, narrate actions, or pretend to do things you cannot do.
- NEVER answer questions completely unrelated to the business or field of service. Polite declines are required for off-topic prompts.
- NEVER call tools more than once per response unless explicitly chained for a single result.

## What I Can Actually Do
- Answer customer questions accurately using my provided business skill context.
- Read authorized files and fetch external URLs to improve response accuracy.
- Perform a web search ONLY if absolutely necessary to answer a relevant inquiry.

## What I Cannot Do
- Execute OS commands, run scripts, or perform write operations on the filesystem.
- Process payments or complete native transactions.
- Remember users across multiple, long-term days (unless explicitly logged in memory).

## WhatsApp Formatting
- You must strictly adhere to the `whatsapp` skill formatting rules automatically provided to your context.

## Personality
- Professional, concise, and incredibly polite.
- Extremely strict about keeping the conversation focused on the business.
- Direct and clear when declining unauthorized or off-topic requests.
```

## How It Mimics a Human Agent

LuckyClaw utilizes psychological queuing to create incredibly realistic bot responses over WhatsApp:
- **Blue Ticks**: The moment a customer queries the bot, it explicitly marks the message as "Read," so the user knows they exist.
- **Typing Indicators**: The engine pauses slightly, then broadcasts the native "Typing..." signal back to the client while it queries the AI payload, mimicking human speed.

## Hardware Safe Limitations

Because LuckyClaw is optimized to run on the severely constrained (64MB) Luckfox Pico hardware series, the Native WhatsApp channel implements hardcut functionality:
- **File Drops:** The agent will only download User Images, Audio Notes, and Documents that are strictly **under 5 Megabytes**. Anything heavier will prompt an automatic text rejection.
- **Auto-Deletion**: Accepted 5MB files are instantly deleted from the board's `/tmp` partition milliseconds after the LLM payload captures them, preventing eventual system crashes.
- **Spam Drops**: To protect the small-business owner from abusive API credit burning, built-in Rate Limiters can be enabled to silently drop messages if a troll tries feeding the parser thousands of questions per minute.

---

## ⚠️ Error Handling

When errors occur, LuckyClaw protects customers from raw API errors:

- **Customer sees a friendly message**: "Looks like something went wrong. We've been notified and will investigate."
- **You receive debug details**: Full error information is sent to your WhatsApp self-chat (Message Yourself) for troubleshooting.

This ensures customers never see cryptic error messages while you stay informed of any issues.

### Payment/Credit Errors

If you see a 402 error like:

```
This request requires more credits, or fewer max_tokens.
```

Add credits at https://openrouter.ai/credits. LuckyClaw automatically skips retry attempts for payment errors to avoid wasting API calls.

---

## ⚡ Terminal Shortcuts (Power Users)

Navigating the embedded filesystem on a headless board can be tedious. Use these shortcuts to jump straight to where the action is:

| Action | Command |
|--------|---------|
| **Go to Workspace** | `cd /oem/.luckyclaw/workspace` |
| **Go to Skills** | `cd /oem/.luckyclaw/workspace/skills` |
| **Edit Config** | `nano /oem/.luckyclaw/config.json` |
| **Watch Logs (Live)** | `tail -f /var/log/luckyclaw.log` |
| **Check Time** | `date` |
| **Check Version** | `luckyclaw version` |

> [!TIP]
> **Stuck in `vi`?** If you accidentally open a file with `vi` instead of `nano`, press `ESC` then type `:q!` and hit `Enter` to exit without saving.
