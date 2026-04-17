# Backup and Restore

Flashing a new firmware image replaces the entire filesystem on the board. All configuration, memories, sessions, and cron jobs will be lost unless backed up beforehand.

This guide covers how to preserve your data before reflashing and restore it afterward.

> A future `luckyclaw update` command is planned (see [ROADMAP.md](ROADMAP.md)) that will update only the binary without touching your data.

## What Gets Backed Up

| Item | Path on Device | Contains |
|------|---------------|----------|
| Config | `/oem/.luckyclaw/config.json` | API key, model, channel tokens, tool settings |
| Workspace | `/root/.luckyclaw/workspace/` | Memories, cron jobs, skills, identity files |
| WhatsApp | `/root/.luckyclaw/whatsapp.db/` | SQLite database for WhatsApp session stability |

## Backup (Before Flashing)

Run this from your computer (not the board):

```bash
# Backup config and heartbeat log
sshpass -p 'luckfox' scp -r root@<DEVICE_IP>:/oem/.luckyclaw/ ./luckyclaw-backup-oem/

# Backup workspace (memories, sessions, cron, skills, WhatsApp)
sshpass -p 'luckfox' scp -r root@<DEVICE_IP>:/root/.luckyclaw/ ./luckyclaw-backup-root/
```

Replace `<DEVICE_IP>` with your board's IP address (e.g., `192.168.1.86`).

Verify the backup contains your files:

```bash
ls ./luckyclaw-backup-oem/
# Should show: config.json  heartbeat.log
ls ./luckyclaw-backup-root/
# Should show: workspace/  whatsapp.db/
```

## Restore (After Flashing)

After flashing the new firmware and running `luckyclaw onboard`, restore your data:

```bash
# Restore config and heartbeat log to /oem
sshpass -p 'luckfox' scp -r ./luckyclaw-backup-oem/* root@<DEVICE_IP>:/oem/.luckyclaw/

# Restore workspace and WhatsApp data to /root
sshpass -p 'luckfox' scp -r ./luckyclaw-backup-root/* root@<DEVICE_IP>:/root/.luckyclaw/

# Restart the gateway to pick up restored config
sshpass -p 'luckfox' ssh root@<DEVICE_IP> "/etc/init.d/S99luckyclaw restart"
```

## Notes

- **Paths**: Config and heartbeat log are stored on the `/oem` partition. Workspace (sessions, memories, cron jobs, skills) and WhatsApp session data are stored at `/root/.luckyclaw/` on the rootfs partition.
- **Binary**: The backup does not include the binary itself (`/usr/bin/luckyclaw`) -- that is part of the firmware image.
- **WhatsApp**: Session files are **SQLite databases**, not JSON. Backing up the entire `whatsapp.db/` directory is required.
- **Version Compatibility**: If upgrading between major versions, check for config format changes. Old sessions/memories are generally backward-compatible.
