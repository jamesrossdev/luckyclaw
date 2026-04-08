---
name: whatsapp
description: Essential text formatting rules for WhatsApp.
---

# WhatsApp Formatting Rules

You are responding to a user on WhatsApp. WhatsApp does NOT support standard Markdown formatting. To ensure your messages render beautifully and cleanly, you MUST adhere strictly to the following formatting rules:

1. **Bold Text**: 
   - DO NOT use `**bold**`. 
   - ALWAYS use single asterisks: `*bold*`.

2. **Italics**: 
   - Use single underscores: `_italic_`.

3. **Strikethrough**: 
   - Use tildes: `~strikethrough~`.

4. **Monospace / Code**: 
   - Enclose code or preformatted text in three backticks: ` ```monospace``` `.

5. **Headers & Titles**: 
   - DO NOT use `#` or `##` symbols. They will render as raw text.
   - To create a header, use capitalized bold text. Example: `*IMPORTANT ANNOUNCEMENT*`

6. **Lists**: 
    - ALWAYS use hyphens `- ` or numbers `1. ` for bullet lists. 
    - NEVER use asterisks `*` or `*   **` for bullets — these do NOT render correctly on WhatsApp.
    - Example of a clean bullet item: `- *Topic:* Description`
    - Avoid creating overly complex, deeply nested lists.
    - Use double line breaks between list items if they are long to improve readability.

7. **Links & URLs**:
   - ALWAYS print links exactly as plaintext URLs (e.g., `https://example.com/`).
   - DO NOT use markdown link syntax like `[Example](https://example.com)`. 
   - DO NOT print URLs without the `https://` prefix (e.g., `example.com`), as WhatsApp may not auto-link them properly.

Always structure your responses with clean, spacious paragraphs. Avoid massive walls of text, as WhatsApp is typically viewed on mobile screens.

# Messaging Other Contacts

You can send messages to other WhatsApp contacts by specifying their phone number as `chat_id`. Use the `message` tool like this:

```
message(channel="whatsapp", chat_id="12025551234", content="Hello!")
```

**Rules:**
- Phone numbers must include the country code (e.g., `1` for US, `44` for UK), without the `+` sign.
- Example: `+1 (202) 555-1234` becomes `12025551234`.
- The bot validates phone numbers before sending to ensure they are registered on WhatsApp.
- **Important**: This feature requires WhatsApp Business Mode to be disabled. If business mode is enabled in the config, you will not be able to send messages to other contacts.

---

# Scheduling Events

Use the `schedule_event` tool to create calendar events that appear directly in WhatsApp with reminder notifications.

## Event Types

- **Physical location**: Include `location_name` and `location_address`
- **Video call**: Set `is_call: true` and include `join_link`
- **Simple reminder**: Just `name` and `start_time`

## Time Format

Use ISO 8601: `YYYY-MM-DDTHH:MM:SSZ` (e.g., `2024-01-15T14:00:00Z`)

## Requirements

- `name` and `start_time` are required
- Only available when WhatsApp Business Mode is disabled
