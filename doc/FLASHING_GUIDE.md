# LuckyClaw Flashing Guide (eMMC)

This guide covers flashing the LuckyClaw firmware to a Luckfox Pico Plus/Pro board's eMMC storage. 

> [!NOTE]
> Currently, only eMMC flashing is supported by this guide. SPI NAND flashing instructions will be added in a future update.

## Prerequisites

### Hardware
* Luckfox Pico Plus or Pro board
* USB Type-C to Type-A cable (data capable, not just charging)
* A computer running Windows (required for SOCToolKit, see IMPROVEMENTS.md for future cross-platform plans)

### Software & Files
* The latest `update.img` file from the LuckyClaw releases page
* **Rockchip SOCToolKit** (The official flashing tool)

## Step 1: Download and Extract SOCToolKit

1. Download the [official Rockchip SOCToolKit](https://files.luckfox.com/wiki/Luckfox-Pico/Software/SocToolKit.zip).
2. Extract the ZIP archive to a convenient location on your Windows machine.
3. Inside the extracted folder, run `SocToolKit.exe`.

![SOCToolKit Main Interface](../assets/flashing/step-01-soctoolkit-main.png)

## Step 2: Enter MaskROM Mode

To flash the board, it must be in MaskROM mode before connecting it to your PC.

1. **Disconnect** the USB cable from the board.
2. Locate the **BOOT button** on the Luckfox Pico board (usually near the USB-C port).
3. **Press and hold** the BOOT button.
4. While holding the BOOT button, plug the USB cable into the board and your computer.
5. Wait 2-3 seconds, then **release** the BOOT button.

If successful, SOCToolKit will display a "Maskrom Device" in the device list at the bottom.

![MaskROM Device Detected](../assets/flashing/step-02-maskrom-device.png)

## Step 3: Select the Firmware Image

1. In the SOCToolKit interface, navigate to the **Download Image** (or Upgrade) tab.
2. Click the browse button (`...`) next to the Firmware path field.
3. Select the `update.img` file you downloaded for LuckyClaw.

![Selecting firmware image](../assets/flashing/step-03-select-firmware.png)

## Step 4: Flash the Board

1. Make sure your device is still listed as a MaskROM device.
2. Click the **Upgrade** (or Run) button.
3. The flashing process will begin. You will see a progress bar and log output. **Do not disconnect the cable.**

![Flashing in progress](../assets/flashing/step-04-flashing-progress.png)

## Step 5: Completion and Reboot

Once the process reaches 100%, you should see a "Download Image Success" message.

1. The board will typically reboot automatically.
2. If it does not, press the Reset button or unplug and replug the USB cable (without holding the BOOT button).
3. The board is now running LuckyClaw!

![Flashing successful](../assets/flashing/step-05-success.png)

## Troubleshooting

* **Device not detected in SOCToolKit:** Ensure you are using a data-capable USB cable. Try a different USB port on your computer. Make sure you are pushing the BOOT button *before* and *during* the cable insertion.
* **Flashing fails at X%:** This is often caused by a loose USB connection or a faulty cable. Try replacing the cable.
* **"Test Device Fail" error:** The board might have exited MaskROM mode. Try the BOOT button sequence again and flash immediately.

## Further Reading

For more details on alternative flashing methods or operating systems, refer to the [official Luckfox Pico burning instructions](https://wiki.luckfox.com/Luckfox-Pico/Luckfox-Pico-quick-start/image-burn).
