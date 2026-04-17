# 🛠️ Developer Guide

### Prerequisites

- Go 1.25+
- [Luckfox Pico SDK](https://github.com/LuckfoxTECH/luckfox-pico) (for firmware builds)
- ARM cross-compilation toolchain (included in the SDK)

### Build from source

```bash
git clone https://github.com/jamesrossdev/luckyclaw.git
cd luckyclaw

# Build for your local machine
make build

# Cross-compile for Luckfox Pico (ARMv7)
make build-arm
```

### Development Workflow

Keep the codebase clean using the integrated Makefile targets:

- `make fmt` — Format Go code
- `make vet` — Run static analysis
- `make test` — Run unit tests
- `make check` — Run all of the above (recommended before committing)
- `make clean` — Remove build artifacts

### Build firmware image

The firmware overlay only contains OS-level files that get baked into `rootfs.img`. The **workspace templates** (`SOUL.md`, skills, etc.) are **embedded directly into the binary** via `go:embed workspace` — so every binary already carries the full workspace inside it. Users get workspace files by running `luckyclaw onboard`, which extracts them to `/root/.luckyclaw/workspace/`.

```
firmware/overlay/
└── etc/
    ├── init.d/S99luckyclaw          # Auto-start on boot
    ├── profile.d/luckyclaw-banner.sh # SSH login banner
    └── ssl/certs/ca-certificates.crt # TLS certificates
```

To build a distributable firmware image:

1. **Build the ARM binary** (workspace is embedded automatically):
   ```bash
   make build-arm
   # Output: build/luckyclaw-linux-arm
   ```

2. **Clone the SDK** (one-time setup):
   ```bash
   git clone https://github.com/LuckfoxTECH/luckfox-pico.git luckfox-pico-sdk
   ```

3. **Sync the `etc/` overlay to the SDK** (do this if init script or banner changed):

## SDK Overlay & Buildroot Optimization

Because the ARM binary embeds its `workspace/` at compile time, changes to the daemon are zero-friction. However, to construct a completely stripped and optimized Luckfox firmware image from the 5GB SDK, we use tracked overlays and patches.

### 1. `etc/` Overlay Linkage
The SDK overlay `etc/` partition must stay in sync with our repo to inject the init script and SSH banner inside `rootfs`:
```bash
cp -r firmware/overlay/etc/* luckfox-pico-sdk/project/cfg/BoardConfig_IPC/overlay/luckyclaw-overlay/etc/
```

### 2. Rootfs Bloatware Purging
The stock SDK packs ~30MB of bloated diagnostic tools (Python, iperf, v4l2) and GUI libraries (FreeType, LibDRM) that we do not need, which dangerously compresses our available working flash space. Apply our storage optimization patch BEFORE building the firmware:
```bash
# This patch aggressively strips Buildroot defconfigs of unnecessary media/GUI packages
cd luckfox-pico-sdk
git apply ../firmware/sdk-patches/optimize-rootfs.patch
```

> **Note**: If `git apply` fails because you've modified Buildroot before, you can use `patch -p1 < ../firmware/sdk-patches/optimize-rootfs.patch` instead, or just run `./build.sh clean` first.

### 3. Image Build
```bash
# Strongly recommended to clean rootfs first to avoid carrying over bloated binaries
cd luckfox-pico-sdk 
./build.sh clean rootfs

# Build rootfs and pack it into the final update.img firmware
./build.sh
```

### 4. Find the output image
The full, flashable firmware image will be generated deeply nested inside the `IMAGE` directory, for example:
```
luckfox-pico-sdk/IMAGE/IPC_SPI_NAND_BUILDROOT_RV110x.../IMAGES/update.img
```
Rename this file to match our project release taxonomy: `luckyclaw-luckfox_pico_plus_rv110x-vX.Y.Z.img`.

When a user flashes this image and runs `luckyclaw onboard`, the embedded workspace is extracted to `/root/.luckyclaw/workspace/`.

### Project structure

```
luckyclaw/
├── cmd/luckyclaw/main.go    # Entry point, CLI, onboarding wizard (embeds workspace/)
├── docs/                    # Project documentation exactly like this guide
├── pkg/
│   ├── agent/               # Core agent loop and context builder
│   ├── bus/                 # Internal message bus
│   ├── channels/            # Telegram, Discord, and other messaging integrations
│   ├── config/              # Configuration and system settings
│   ├── providers/           # LLM provider implementations (OpenRouter, etc.)
│   ├── skills/              # Skill loader and installer
│   ├── tools/               # Agent tools (shell, file, i2c, spi, send_file)
│   └── ...
├── firmware/overlay/etc/    # Init script + SSH banner baked into firmware image
├── firmware/sdk-patches/    # Automated patches to trim Luckfox SDK bloat
├── workspace/               # Templates embedded into binary via go:embed
└── assets/                  # Documentation images and media
```

### Performance tuning

LuckyClaw automatically sets `GOGC=20` and `GOMEMLIMIT=24MiB` at startup for memory-constrained boards. These can be overridden via environment variables if your board has more RAM.
