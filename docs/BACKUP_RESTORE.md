# Backup and Restore

Flashing a new firmware image replaces the entire filesystem on the board. All configuration, memories, sessions, and cron jobs will be lost unless backed up beforehand.

This guide covers how to preserve your data before reflashing and restore it afterward.

> A future `luckyclaw update` command is planned (see [ROADMAP.md](ROADMAP.md)) that will update only the binary without touching your data.

## What Gets Backed Up

| Item | Path on Device | Contains |
|------|---------------|----------|
| Config | `/oem/.luckyclaw/config.json` | API key, model, channel tokens, tool settings |
| Workspace | `/oem/.luckyclaw/workspace/` | Memories, cron jobs, skills, identity files |
| WhatsApp | `/oem/.luckyclaw/whatsapp.db/` | SQLite database for WhatsApp session stability |

## Backup (Before Flashing)

Run this from your computer (not the board):

```bash
sshpass -p 'luckfox' scp -r root@<DEVICE_IP>:/oem/.luckyclaw/ ./luckyclaw-backup/
```

Replace `<DEVICE_IP>` with your board's IP address (e.g., `192.168.1.86`).

Verify the backup contains your files:

```bash
ls ./luckyclaw-backup/
# Should show: config.json  workspace/  whatsapp.db/
```

## Restore (After Flashing)

After flashing the new firmware and running `luckyclaw onboard`, restore your data:

```bash
# Restore entire data directory
sshpass -p 'luckfox' scp -r ./luckyclaw-backup/* root@<DEVICE_IP>:/oem/.luckyclaw/

# Restart the gateway to pick up restored config
sshpass -p 'luckfox' ssh root@<DEVICE_IP> "/etc/init.d/S99luckyclaw restart"
```

## Notes

- **Paths**: Production runtime data is stored on the writable `/oem` partition, NOT `/root`.
- **Binary**: The backup does not include the binary itself (`/usr/bin/luckyclaw`) -- that is part of the firmware image.
- **WhatsApp**: Session files are **SQLite databases**, not JSON. Backing up the entire `whatsapp.db/` directory is required.
- **Version Compatibility**: If upgrading between major versions, check for config format changes. Old sessions/memories are generally backward-compatible.
