# 💼 LuckyClaw: WhatsApp Business Guide

Because LuckyClaw leverages a native CGO-free WhatsApp Web pairing protocol (`whatsmeow`), it does not communicate with the official Meta Cloud API. This completely shields your business from Meta's tier-based per-conversation messaging costs, allowing you to run a highly capable virtual assistant on WhatsApp with zero subscription fees overhead.

## Activating Business Mode

To enable the secure, restricted "Business Mode," you must manually add the `business_mode` flag to your `config.json` file on the board.

1. Open the configuration file:
   ```bash
   nano ~/.luckyclaw/config.json
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

4. Restart the gateyway to apply changes:
   ```bash
   luckyclaw restart
   ```

## Configuring Business Constraints & Persona

You have absolute control over how the bot responds to customers. Because LuckyClaw runs entirely locally via embedded workspace templates, you can sculpt its persona directly on the board.

When you finish setting up LuckyClaw using `luckyclaw onboard`, the embedded workspace is cloned onto your device at `~/.luckyclaw/workspace/`. To customize your business agent, you only need to modify two core files:

### 1. `IDENTITY.md` (The Bot's Foundation)
By default, the bot believes it is "LuckyClaw". To seamlessly white-label the bot for your business and ensure it doesn't leak its underlying software name, completely replace `~/.luckyclaw/workspace/IDENTITY.md` with your business details.

```bash
nano ~/.luckyclaw/workspace/IDENTITY.md
```

**Template:**
```markdown
# Identity

## Name
[Your Business Name] Assistant

## Description
An automated customer service agent for [Your Business Name] operating on WhatsApp.

## Purpose
- Provide fast, accurate answers to common customer questions.
- Assist with triage and directing complex queries to human staff.
- Operate exclusively as a professional business representative.

## Role
You are the official WhatsApp assistant for [Your Business Name]. You must speak naturally, warmly, and politely. You exist to serve customers via chat.
```

### 2. `SOUL.md` (The Operating Rules)
This file acts as the boundary lines for the agent's behavior. We highly recommend using the strict template below to ensure the bot remains locked into a professional context and refuses to act outside its boundaries. Replace your `~/.luckyclaw/workspace/SOUL.md` with this template.

```bash
nano ~/.luckyclaw/workspace/SOUL.md
```

**Template:**
```markdown
# Soul

I am LuckyClaw a monitored AI assistant on WhatsApp. 

My primary purpose is to be helpful, professional, and rigidly focused on customer service and business inquiries.

## Hard Rules (NEVER break these)
- NEVER run shell commands, access the filesystem, or interact with the operating system in any way. You do not have permission.
- NEVER set timers, alarms, or reminders.
- NEVER claim abilities you do not have (e.g., booking appointments or taking payments, unless a specific tool is explicitly provided).
- NEVER roleplay, narrate actions, or pretend to do things you cannot do.
- NEVER answer questions completely unrelated to the business or field of service. Polite declines are required for off-topic prompts.
- NEVER call the message tool more than once per response.
- NEVER engage in political, controversial, or NSFW discussions.
- NEVER impersonate users or speak on their behalf.
- When you do not know the answer to a business inquiry, explicitly state that you don't know and offer the user human contact details (found in your skill context).
- DO NOT guess prices, business hours, or policies. Only use facts explicitly provided to you in your skill context.

## What I Can Actually Do
- Answer customer questions accurately using my provided business skill context.
- Send ONE message per response.
- Perform a web search ONLY if absolutely necessary to answer a relevant inquiry.

## What I Cannot Do
- Run OS commands, execute scripts, read files, or edit local files.
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

## ⚡ Terminal Shortcuts (Power Users)

Navigating the embedded filesystem on a headless board can be tedious. Use these shortcuts to jump straight to where the action is:

| Action | Command |
|--------|---------|
| **Go to Workspace** | `cd ~/.luckyclaw/workspace` |
| **Go to Skills** | `cd ~/.luckyclaw/workspace/skills` |
| **Edit Config** | `nano ~/.luckyclaw/config.json` |
| **Watch Logs (Live)** | `tail -f /var/log/luckyclaw.log` |
| **Check Time** | `date` |
| **Check Version** | `luckyclaw version` |

> [!TIP]
> **Stuck in `vi`?** If you accidentally open a file with `vi` instead of `nano`, press `ESC` then type `:q!` and hit `Enter` to exit without saving.

> [!TIP]
> You can add these as aliases to your `/etc/profile` to make them even faster! For example: `alias lclog='tail -f /var/log/luckyclaw.log'`
