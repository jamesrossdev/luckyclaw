# MiniMax Provider

MiniMax is supported in LuckyClaw via MiniMax's text endpoint (`/text/chatcompletion_v2`).

## Supported Models

LuckyClaw supports both model naming styles:

| Short Name | Full Name | Notes |
|------------|-----------|-------|
| `minimax-m2.7` | `minimax-coding-plan/MiniMax-M2.7` | Primary example in this doc |
| `minimax-m2` | `minimax-coding-plan/MiniMax-M2` | Alternative model |

Both naming styles work. The short name (`minimax-m2.7`) is recommended for simplicity.

## Text and Chat

Text and chat work reliably with minimax-m2.7. Standard conversation, reasoning tasks, and general assistant use cases are supported.

## Image Understanding

Image understanding with minimax-m2.7 is **currently unreliable** in our observed setup. When users send images (via WhatsApp, Telegram, or Discord), the model often returns empty content or reasoning that indicates no image was seen.

**Practical guidance:** If you need image analysis today, consider:
- Using a vision-capable model via OpenRouter or another provider
- Using an external tool/MCP pipeline designed for image understanding
- Until this is verified working, do not rely on minimax-m2.7 for image-heavy tasks

> This is an observed limitation in our setup. Your results may vary based on how images are sent and the specific use case. We are not claiming guaranteed native image support for m2.7 right now.

## Account Types and Billing

MiniMax offers two account types, determined by your API key, not by a config setting:

| Account Type | How to Identify | Billing |
|--------------|-----------------|---------|
| **coding-plan** | Key obtained from platform.minimax.io with coding-plan quota | Uses prepaid plan quota |
| **paygo** | Key with pay-as-you-go billing | Uses paygo balance |

LuckyClaw does not need to know which account type you have — it simply sends your API key as a Bearer token. MiniMax determines billing based on your key/account.

> **Do not set a `billing_mode` config field** — it is not used. Billing is handled entirely by your MiniMax account/key.

## Configuration

```json
{
  "agents": {
    "defaults": {
      "provider": "minimax",
      "model": "minimax-m2.7"
    }
  },
  "providers": {
    "minimax": {
      "api_key": "YOUR_MINIMAX_API_KEY",
      "api_base": ""
    }
  }
}
```

### Minimal Config (required fields only)

```json
{
  "agents": {
    "defaults": {
      "provider": "minimax",
      "model": "minimax-m2.7"
    }
  },
  "providers": {
    "minimax": {
      "api_key": "YOUR_MINIMAX_API_KEY"
    }
  }
}
```

- **`api_key`** (required): Your MiniMax API key from [platform.minimax.io](https://platform.minimax.io/subscribe/token-plan?code=L18gHx02iM&source=link).
- **`api_base`** (optional): Defaults to `https://api.minimax.io/v1`. Only change this if MiniMax provides a custom endpoint.

## Channel Compatibility

| Channel | Status | Details |
|---------|--------|---------|
| **WhatsApp** | ✅ Text supported | Image understanding unreliable |
| **Telegram** | ✅ Text supported | Image understanding unreliable |
| **Discord** | ✅ Text supported | Image understanding unreliable |

## Common Issues

### Provider still set to openrouter/openai

If the gateway ignores your MiniMax config, verify `agents.defaults.provider` is set to `"minimax"` (not `"openrouter"` or `"openai"`):

```json
"agents": {
  "defaults": {
    "provider": "minimax",
    "model": "minimax-m2.7"
  }
}
```

### Missing MiniMax API key

The gateway will return an error on startup if `providers.minimax.api_key` is empty or missing. Ensure you have a valid key from [platform.minimax.io](https://platform.minimax.io/subscribe/token-plan?code=L18gHx02iM&source=link).

### Wrong model string

`minimax-m2.7` is the recommended default. If you use a model name your MiniMax account does not support, the API will return an error.

### Image tasks not working

See the Image Understanding section above. For production image tasks, use a vision-capable model or external pipeline.

## Getting an API Key

1. Sign up at [platform.minimax.io](https://platform.minimax.io/subscribe/token-plan?code=L18gHx02iM&source=link)
2. Navigate to API Keys section
3. Create a new key with appropriate permissions
4. Copy the key into your `config.json` under `providers.minimax.api_key`