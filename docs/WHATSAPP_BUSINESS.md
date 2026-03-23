# 💼 LuckyClaw: WhatsApp Business Guide

Because LuckyClaw leverages a native CGO-free WhatsApp Web pairing protocol (`whatsmeow`), it does not communicate with the official Meta Cloud API. This completely shields your business from Meta's tier-based per-conversation messaging costs, allowing you to run a highly capable virtual assistant on WhatsApp with zero subscription fees overhead.

## Configuring Business Constraints & Persona

You have absolute control over how the bot responds to customers. Because LuckyClaw runs entirely locally via embedded workspace templates, you can sculpt its persona directly on the board.

When you finish setting up LuckyClaw using `luckyclaw onboard`, the embedded workspace is cloned onto your device at `~/.luckyclaw/workspace/`. To customize your business agent, you only need to modify two core files:

### 1. `IDENTITY.md` (The Bot's Foundation)
This file tells the LLM perfectly *who* it is. Keep it sharp and strict.

```markdown
You are a highly professional, friendly, and deeply apologetic customer support agent for "Bob's Plumbing LLC".
Your name is "LuckyClaw". You are communicating exclusively via WhatsApp.
You must always speak in a supportive, concise, and human-like tone.
```

### 2. `SOUL.md` / `AGENT.md` (The Operating Rules)
This file acts as the boundary lines for the agent's behavior. You can explicitly forbid hallucinations or unapproved business practices here.

```markdown
CRITICAL RULES:
1. You MUST NEVER offer discounts or predict pricing estimates to customers.
2. If a user asks a complex technical question about a pipe burst, politely inform them that a human manager will review their WhatsApp message and call them shortly.
3. Keep your responses to ONE standard paragraph for readability on mobile screens. Do not output massive markdown tables.
4. If the user uses profanity or becomes hostile, inform them the chat is being disconnected and stop responding.
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
