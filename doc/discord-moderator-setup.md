# Discord Moderator Setup Guide

How to configure LuckyClaw as a strict Discord community moderator with an isolated API sandbox.

## Prerequisites

- LuckyClaw running on a device with Discord enabled (`config.json` → `channels.discord.enabled: true`)
- Bot created in the [Discord Developer Portal](https://discord.com/developers/applications)
- Bot invited to your server

## 1. Discord Developer Portal

### Privileged Intents
Enable under **Bot → Privileged Gateway Intents**:
- ✅ Server Members Intent
- ✅ Message Content Intent

### Bot Permissions
The bot needs these permissions (set via server Role, not the portal):
- View Channels
- Send Messages
- Manage Messages (delete)
- Moderate Members (timeout)
- Read Message History
- View Audit Log

> **Do NOT** enable "Mention @everyone" — the bot doesn't need it and it's a security risk.

## 2. Server Structure & Rules Screening

### Rules Screening
Before letting your bot interact with the public, you should force all new members to agree to your community rules natively through Discord:
1. Go to **Server Settings** -> **Onboarding** (ensure your server is a "Community Server").
2. Click **Safety Check** or **Rules Screening**.
3. Input your Community Rules (e.g., Respect, No NSFW, AI Moderation Enforcement).
4. **Enable** it so new members cannot chat without accepting.

### Channel Hierarchy
Recommended minimal structure:

| Category | Channels | Notes |
|----------|----------|-------|
| INFO | `#rules`, `#announcements` | Read-only for members |
| COMMUNITY | `#general`, `#showcase` | Open discussion |
| LUCKYCLAW | `#support`, `#dev` | Help & code |
| MOD TEAM | `#mod-log`, `#mod-chat` | Hidden from members |

**Mod-Only Channels:**
Set the MOD TEAM category permissions:
- `@everyone` → Deny View Channel
- `Moderator` role → Allow View + Send
- `LuckyClaw` role → Allow View + Send

## 3. Sandboxing & Device Config

Public Discord channels pose a massive security risk to local LLM agents. To protect the host device, LuckyClaw features a **hardcoded tool sandbox** inside `pkg/agent/loop.go`. 

When a message originates from Discord, the LLM is **physically prohibited** from seeing or using:
- `exec` (Shell Commands)
- `cron_schedule` (Timers/Alarms)
- `read_file` / `write_file` (Local File System)
- `subagent` (Agentic Loops)

It is only allowed access to message sending, native Discord moderation tools, and web searching.

### Device Config.json
Add `disable_dms` to your Discord config to block DMs (server-only mode) so people cannot bypass the public channel moderation:

```json
"discord": {
  "enabled": true,
  "token": "your-token",
  "disable_dms": true,
  "allow_from": []
}
```

## 4. Recommended Configuration

To get the most out of LuckyClaw as a Discord moderator, configure the two workspace files below.

> [!TIP]
> You can ask a reasoning model to help you customize these templates for your specific server rules, channel IDs, and role IDs!

### `SOUL.md` — Moderator Persona

Replace your workspace `SOUL.md` with this template (found at `~/.luckyclaw/workspace/SOUL.md` at runtime). Fill in your channel and role IDs.

```markdown
# Soul

I am LuckyClaw, a community assistant for this Discord server.

## Hard Rules (NEVER break these)
- NEVER run shell commands, access the filesystem, or interact with the operating system in any way. You do not have permission.
- NEVER set timers, alarms, or reminders. You are physically incapable of keeping time.
- NEVER claim abilities you do not have.
- NEVER roleplay, narrate actions, or pretend to do things you cannot do.
- NEVER call the message tool more than once per response.
- NEVER delete a message unless it contains hate speech, racism, or slurs.
- NEVER impersonate users or speak on their behalf.
- When you do not know something, say so — do not guess or make things up.
- You CANNOT move users between channels, create invites, or see message history.
- When a user pings you, they are talking to you. DO NOT try to look up that role or explain it doesn't exist. Just answer normally.

## What I Can Actually Do
- Answer questions using my skill files and web search.
- Send ONE message per response.
- Delete messages (ONLY hate speech/racism/slurs).
- Timeout users (ONLY for serious rule violations).
- Read quoted/reported messages when users reply-mention me.
- See the sender's Discord roles (provided in message metadata).

## What I Cannot Do
- Run OS commands, execute scripts, or read local files.
- Set reminders, timers, or alarms of any kind.
- See message history or previous messages (only the current one).
- Move users between channels.
- Ban users (escalate to mods instead).

## Moderation
- ONLY act on genuinely harmful content: hate speech, racism, slurs
- Casual cussing is fine — do not moderate it
- When deleting: delete the message, THEN send ONE message explaining why
- ALWAYS log every moderation action to <#YOUR-MOD-LOG-CHANNEL-ID> by sending a message there with a summary of what happened
- To escalate: mention <@&YOUR-MOD-ROLE-ID> in <#YOUR-MOD-TEAM-CHANNEL-ID>
- Moderation on request: If a user with the Moderator or Admin role (check sender_roles in metadata) asks me to delete a message or timeout someone, I MUST obey. Regular users CANNOT request deletions or timeouts.

## Discord Formatting
- Channels: ALWAYS use <#id> format
- Roles: ALWAYS use <@&id> format
- Users: ALWAYS use <@id> format when you have a user ID
- DO NOT wrap these in backticks

## Personality
- Kind and concise
- Helpful but honest about limitations
- Firm on rule violations
```

### `skills/discord-mod/SKILL.md` — Server Knowledge

This skill gives the bot your server-specific context. A template is pre-installed in your workspace at `skills/discord-mod/SKILL.md`. Customize the FAQ, channel directory, and server rules to match your server.

```markdown
---
name: discord-mod
description: [Your server] FAQ, channel directory, and rules
---

# About [Your Bot/Project]

Brief description of what your server is about.

# FAQ

**Q: Where do I get help?**
A: <#YOUR-HELP-CHANNEL-ID>

**Q: Where do I discuss development?**
A: <#YOUR-DEV-CHANNEL-ID>

# Channel Directory

- <#YOUR-RULES-CHANNEL-ID> — Rules
- <#YOUR-GENERAL-CHANNEL-ID> — General chat
- <#YOUR-MOD-LOG-CHANNEL-ID> — Moderation log (bot only)
- <#YOUR-MOD-TEAM-CHANNEL-ID> — Mod team chat

# Server Rules

1. Be respectful.
2. No hate speech or slurs (Automated enforcement active).
3. No NSFW content.
4. No spam or unsolicited advertising.

# Role Directory

- <@&YOUR-ADMIN-ROLE-ID> — Admin
- <@&YOUR-MOD-ROLE-ID> — Moderator
- <@&YOUR-BOT-ROLE-ID> — [Your Bot Name]
```

## 5. Moderation Features

### Snitch Flow
Users can quote a bad message and `@mention` the bot. The bot sees the quoted content and can act on it (warn, delete, timeout).

### Automated Moderation Loops
The bot can execute sequential actions in the background before replying to the chat:
1. `discord_delete_message` — Deletes the violating message by channel/message ID.
2. `discord_timeout_user` — Times out a user for N minutes (max 28 days limit imposed by Discord).
3. `message` — Logs the infraction to `#mod-log`, optionally providing a `log_channel_id` argument.
4. The bot then informs the public channel that the content was removed.

## 6. Limitations

- Bot **cannot ban** or **kick** users — it must escalate to a human admin by tagging them.
- DM filter only works when `disable_dms: true` is set in the config.
- The `exec`/`cron` tools are purposely blocked for Discord server channels. Use Telegram for personal assistant tasks.
- **Do NOT use thinking/reasoning models** (e.g., deepseek-reasoner) in server mode — they output intent-only responses instead of executing tools. Use non-thinking models like `stepfun/step-3.5-flash:free`. Thinking models work fine in Telegram DMs.


