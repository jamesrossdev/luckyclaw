# LuckyClaw Flashing Guide (eMMC)

This guide covers flashing the LuckyClaw firmware to a Luckfox Pico Plus/Pro board's eMMC storage using the Rockchip SOCToolKit on Windows.

> **Note:** Currently, only eMMC flashing is supported by this guide. SPI NAND flashing instructions will be added in a future update.

> **Warning:** Flashing replaces the entire filesystem on the board. All existing configuration, memories, sessions, and cron jobs will be lost. If you are upgrading from a previous version, back up your data first -- see [Backup and Restore](BACKUP_RESTORE.md).

## Prerequisites

### Hardware

- Luckfox Pico Plus or Pro board
- USB Type-C to Type-A cable (must be data capable, not charge-only)
- A computer running Windows

### Software and Files

You need three things:

1. **LuckyClaw firmware image** (`luckyclaw-vX.X.X.img`) -- download from the [LuckyClaw GitHub Releases](https://github.com/jamesrossdev/luckyclaw/releases) page.

2. **Rockchip Driver Assistant** -- installs the USB driver so Windows can communicate with the board in MaskROM mode.

3. **Rockchip SOCToolKit** -- the flashing utility that writes the firmware image to the board.

Both Driver Assistant and SOCToolKit can be downloaded from the [Luckfox Downloads page](https://wiki.luckfox.com/Luckfox-Pico-RV1103/Downloads/#common-tools) under "Common Tools". Alternatively, the LuckyClaw release ZIP bundles both tools for convenience.

---

## Step 1: Install the USB Driver

Before flashing, you must install the Rockchip USB driver on your Windows machine.

1. Download and extract the **Driver Assistant** ZIP.
2. Open the extracted folder and run `DriverInstall.exe`.
3. Click **Install Driver**.
4. Wait for the green checkmark confirming the driver was installed successfully.

You only need to do this once per computer. If you have already installed the driver previously, you can skip this step.

---

## Step 2: Open SOCToolKit

1. Download and extract the **SOCToolKit** ZIP.
2. Open the extracted folder and run `SocToolKit.exe`.

SOCToolKit is a portable application -- no installation is required.

![SOCToolKit Main Interface](../assets/flashing/step-01-soctoolkit-main.png)

---

## Step 3: Enter MaskROM Mode

The board must be in MaskROM mode before it can be flashed.

1. **Disconnect** the USB cable from the board if it is currently connected.
2. Locate the **BOOT button** on the Luckfox Pico board (near the USB-C port).
3. **Press and hold** the BOOT button.
4. While holding the BOOT button, plug the USB cable into the board and your computer.
5. Wait 2-3 seconds, then **release** the BOOT button.

If successful, SOCToolKit will display a "Maskrom Device" in the device list at the bottom of the window.

![MaskROM Device Detected](../assets/flashing/step-02-maskrom-device.png)

---

## Step 4: Select the Firmware Image

1. In SOCToolKit, navigate to the **Download Image** tab.
2. Click the browse button next to the Firmware path field.
3. Select the `luckyclaw-vX.X.X.img` file you downloaded.

![Selecting firmware image](../assets/flashing/step-03-select-firmware.png)

---

## Step 5: Flash the Board

1. Confirm your device is still listed as a MaskROM device.
2. Click the **Run** button.
3. The flashing process will begin. A progress bar and log output will be shown. **Do not disconnect the cable during this process.**

![Flashing in progress](../assets/flashing/step-04-flashing-progress.png)

---

## Step 6: Completion

Once the progress reaches 100%, you should see a success message.

1. The board will reboot automatically.
2. If it does not, unplug and replug the USB cable (without holding the BOOT button).
3. The board is now running LuckyClaw.

![Flashing successful](../assets/flashing/step-05-success.png)

### First-Time Setup

After the board boots, connect to it via SSH and run the onboarding wizard:

```bash
ssh root@<DEVICE_IP>
luckyclaw onboard
```

The wizard will walk you through configuring your LLM provider, API key, Telegram bot token, and other settings. For full setup instructions, see the [README](../README.md).

### Restoring Previous Data

If you backed up your data before flashing, follow the [Backup and Restore](BACKUP_RESTORE.md) guide to restore your configuration and workspace.

---

## Troubleshooting

- **Device not detected in SOCToolKit:** Ensure you are using a data-capable USB cable. Try a different USB port. Make sure you held the BOOT button before and during cable insertion. Verify that Driver Assistant was installed successfully.
- **Flashing fails partway through:** This is often caused by a loose USB connection or a faulty cable. Try a different cable.
- **"Test Device Fail" error:** The board may have exited MaskROM mode. Repeat the BOOT button sequence from Step 3.

## Further Reading

- [Official Luckfox Pico burning instructions](https://wiki.luckfox.com/Luckfox-Pico/Luckfox-Pico-quick-start/image-burn)
- [Backup and Restore](BACKUP_RESTORE.md) -- preserve your data before reflashing
- [LuckyClaw README](../README.md) -- full project documentation
