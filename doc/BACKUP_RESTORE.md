# Backup and Restore

Flashing a new firmware image replaces the entire filesystem on the board. All configuration, memories, sessions, and cron jobs will be lost unless backed up beforehand.

This guide covers how to preserve your data before reflashing and restore it afterward.

> A future `luckyclaw update` command is planned (see [Roadmap](ROADMAP.md)) that will update only the binary without touching your data.

## What Gets Backed Up

| Item | Path on Device | Contains |
|------|---------------|----------|
| Config | `/root/.luckyclaw/config.json` | API key, model, channel tokens, tool settings |
| Workspace | `/root/.luckyclaw/workspace/` | Memories, sessions, cron jobs, skills, identity files |

## Backup (Before Flashing)

Run this from your computer (not the board):

```bash
sshpass -p 'luckfox' scp -r root@<DEVICE_IP>:/root/.luckyclaw/ ./luckyclaw-backup/
```

Replace `<DEVICE_IP>` with your board's IP address (e.g., `192.168.1.156`).

Verify the backup contains your files:

```bash
ls ./luckyclaw-backup/
# Should show: config.json  workspace/
```

## Restore (After Flashing)

After flashing the new firmware and running `luckyclaw onboard`, restore your data:

```bash
# Restore workspace (memories, sessions, cron jobs)
sshpass -p 'luckfox' scp -r ./luckyclaw-backup/workspace/ root@<DEVICE_IP>:/root/.luckyclaw/

# Restore config (API key, model, channel settings)
sshpass -p 'luckfox' scp ./luckyclaw-backup/config.json root@<DEVICE_IP>:/root/.luckyclaw/

# Restart the gateway to pick up restored config
sshpass -p 'luckfox' ssh root@<DEVICE_IP> "/etc/init.d/S99luckyclaw restart"
```

## Notes

- The backup does not include the binary itself (`/usr/bin/luckyclaw`) -- that is part of the firmware image.
- If you are upgrading between major versions, check the release notes for any config format changes before restoring an old `config.json`.
- Session files are JSON. If a new version changes the session format, old sessions may be ignored but will not cause crashes.
