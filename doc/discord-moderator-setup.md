# Discord Moderator Setup Guide

How to configure LuckyClaw as a Discord community moderator.

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

## 2. Server Structure

Recommended minimal structure:

| Category | Channels | Notes |
|----------|----------|-------|
| INFO | `#rules`, `#announcements` | Read-only for members |
| COMMUNITY | `#general`, `#showcase` | Open discussion |
| LUCKYCLAW | `#support`, `#dev` | Help & code |
| MOD TEAM | `#mod-log`, `#mod-chat` | Hidden from members |

### Mod-Only Channels
Set the MOD TEAM category permissions:
- `@everyone` → Deny View Channel
- `Moderator` role → Allow View + Send
- `LuckyClaw` role → Allow View + Send

## 3. Device Config

Add `disable_dms` to your Discord config to block DMs (server-only mode):

```json
"discord": {
  "enabled": true,
  "token": "your-token",
  "disable_dms": true,
  "allow_from": []
}
```

> `disable_dms` defaults to `false`. Personal assistant users who want DMs should leave it unset.

## 4. Workspace Files

### SOUL.md
Update with a moderator persona. Key points:
- Friendly but firm on rule violations
- Knows to use `discord_delete_message` and `discord_timeout_user` tools
- Logs all actions to `#mod-log`
- Escalates bans to a human admin

### skills/discord-mod/SKILL.md
Create a skill file with:
- Server rules
- Channel guide
- `#mod-log` channel ID
- Escalation policy (who to tag for bans)
- FAQ about your project

### memory/MEMORY.md
Seed with server context: admin name, server purpose, channel structure.

## 5. Features

### Reply-with-Quote
All bot responses use Discord's native quoted-reply UI, making conversations easy to follow.

### Snitch Flow
Users can quote a bad message and @mention the bot. The bot sees the quoted content and can act on it (warn, delete, timeout).

### Relative Reminders
Users can ask the bot to remind them "in 90 minutes". The bot uses the cron system for this. Clock-time requests ("at 3pm") are declined due to timezone ambiguity.

### Moderation Tools (conditional)
Only registered when Discord is enabled — invisible to Telegram/Slack users:

| Tool | Action |
|------|--------|
| `discord_delete_message` | Delete a message by channel/message ID |
| `discord_timeout_user` | Timeout a user for N minutes (max 28 days) |

## 6. Limitations

- Bot **cannot ban** users — escalate to a human admin
- Bot **cannot kick** users
- DM filter only works when `disable_dms: true` is in config
- Timeout max is 28 days (Discord limit)
- Bot relies on LLM judgement for what constitutes a rule violation
