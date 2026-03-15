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

## 4. Workspace Files

### SOUL.md
Update the system prompt with a strict moderator persona. Because the agent cannot access the `exec` tool, you must explicitly tell it so it doesn't try to guess:

```markdown
## Hard Rules (NEVER break these)
- NEVER run shell commands, access the filesystem, or interact with the operating system in any way. You do not have permission.
- NEVER set timers, alarms, or reminders. You are physically incapable of keeping time.
- NEVER delete a message unless it contains hate speech, racism, or slurs.
- When a user pings you (e.g., `<@LuckyClaw>`), they are just talking to you. DO NOT try to look up that role.
```

### skills/discord-mod/SKILL.md
Create a skill file mapping out the server's context so the bot can answer questions:
- Server rules (Mirror the exact rules you placed in Discord's Rules Screening)
- Channel guide (Explain what `#support` vs `#general` is for)
- `#mod-log` channel ID
- Escalation policy (who to tag for bans, e.g., `@admin`)

## 5. Moderation Features

### Snitch Flow
Users can quote a bad message and `@mention` the bot. The bot sees the quoted content and can act on it (warn, delete, timeout).

### Automated Moderation Loops
The bot can execute sequential actions in the background before replying to the chat:
1. `discord_delete_message` — Deletes the violating message by channel/message ID.
2. `discord_timeout_user` — Times out a user for N minutes (max 28 days limit imposed by Discord).
3. `message` — Logs the infraction specifically to `#mod-log`.
4. The bot then informs the public channel that the hate speech was removed.

## 6. Limitations

- Bot **cannot ban** or **kick** users — it must escalate to a human admin by tagging them.
- DM filter only works when `disable_dms: true` is set in the host file.
- The cron/scheduling tool is purposely broken for the Discord channel to prevent spam alarms in public chats. If you want a personal assistant, use the Telegram channel instead.
