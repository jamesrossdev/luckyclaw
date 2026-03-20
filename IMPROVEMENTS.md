# Suggested Improvements (Backlog)

Items listed here are planned enhancements that are not yet scheduled for implementation.

## Cron Tool Enhancements

### Add `at_time` Parameter
**Priority**: Medium
**Description**: Add a new `at_time` parameter to the cron tool that accepts an ISO-8601 timestamp (e.g., `"2026-02-22T07:00:00+03:00"`). The tool would internally convert this to `atMS` using `time.Parse(time.RFC3339, at_time)`. This eliminates the need for the LLM to manually calculate `at_seconds` from `time.Now()` when the user specifies an absolute clock time for a one-time reminder.

**Benefit**: Reduces LLM arithmetic errors when converting "at 7:10 AM" → `at_seconds`. Currently the LLM must compute `target_time - current_time` in seconds, which is error-prone. With `at_time`, it just passes the ISO string directly.

**Blocked by**: Nothing. Can be implemented independently after Phase 12-H.

## PicoClaw Upstream Bugs

### Infinite Optimization Loop
**Priority**: High
**Description**: In `pkg/agent/loop.go` -> `summarizeSession()`, if the conversational history consists entirely of tool outputs (which causes `len(validMessages) == 0` during the extraction phase), the function returns early without invoking `al.sessions.TruncateHistory()`. This causes the token window boundary to be instantly breached again on the very next turn, locking the agent into an infinite "Memory threshold reached. Optimizing conversation history..." cycle that never actually optimizes.

**Benefit**: Prevents catastrophic session corruption when LLM APIs fail or large tool exchanges dominate a short time window.

**Blocked by**: Should be submitted as a PR to the [picoclaw](https://github.com/sipeed/picoclaw) upstream repository.

> **Status (checked 2026-03-09):** Bug confirmed still present in picoclaw-latest at `pkg/agent/loop.go` lines 1457-1459. The `len(validMessages) == 0` early return skips `TruncateHistory()`, causing the infinite loop. We fixed this in our fork but have not yet opened a PR.

## Installation / Deployment

### `luckyclaw install` Command
**Priority**: High
**Description**: Create a new subcommand `luckyclaw install` that automates the setup of LuckyClaw on a stock Linux/Buildroot environment. It should:
1. Extract and write the init script to `/etc/init.d/S99luckyclaw` (using `go:embed` from the binary).
2. Extract and write the SSH banner to `/etc/profile.d/luckyclaw-banner.sh`.
3. Configure the binary for OOM protection (calling `oom_score_adj` logic).
4. Ensure the default `/oem/.luckyclaw/workspace` exists (calling `onboard` logic if missing).

**Benefit**: Enables a "one-liner" installation for users who already have a working board running stock firmware, without requiring them to reflash using our custom image. Supports the "conservative brother" vision by making the tool easier to adopt on any ARM/Linux hardware.

**Blocked by**: Nothing.

### OTA Binary Updates (No Reflash)
**Priority**: Medium
**Description**: The LuckyClaw binary at `/usr/bin/luckyclaw` can be replaced via SCP without reflashing the entire firmware, since user data lives on `/oem/.luckyclaw/` (a separate partition). An `luckyclaw update` command could check the GitHub Releases API for the latest version, download the matching ARM binary, replace itself, and restart — all without touching config, sessions, cron jobs, or memory.

**Benefit**: Users can update without Windows, without SOCToolKit, and without losing any data. Dramatically lowers the friction of staying current.

**Blocked by**: Needs a stable releases workflow publishing individual ARM binaries (not just full `.img` files). Also needs a version comparison check (`luckyclaw version` already embeds the version tag).

### Open-Source Cross-Platform Flashing Tool
**Priority**: Low
**Description**: Currently, flashing the eMMC requires using the proprietary Rockchip `SOCToolKit`, which is Windows-only. We should develop or adopt an open-source, cross-platform CLI tool (e.g., in Python or Go) that can communicate with the Rockchip MaskROM protocol to flash `update.img` directly from Linux and macOS without needing Windows VMs or proprietary software.

**Benefit**: Dramatically simplifies the onboarding process for non-Windows users and allows for scripted/automated deployments.
**Blocked by**: Reverse engineering of Rockchip protocols or integrating existing open-source alternatives like `rkdeveloptool`.

## Performance Optimizations

### Cache System Prompt Between Messages
**Priority**: Medium
**Description**: `BuildSystemPrompt()` in `pkg/agent/context.go` re-reads `SOUL.md`, `USER.md`, `AGENTS.md`, skills summaries, and memory context from disk on every message. These files rarely change. Caching the result with a file-modification-time check would eliminate repeated disk I/O and string allocations.

**Benefit**: Eliminates ~5 file reads and ~10KB of string allocations per message. On the Luckfox's SPI NAND flash (slower than eMMC), this could save 5-10ms per message.

### Cache Tool Provider Definitions
**Priority**: Low
**Description**: `al.tools.ToProviderDefs()` in `runLLMIteration` rebuilds the full tool definition JSON on every LLM iteration (up to 15 per message). The tool registry doesn't change at runtime, so this can be computed once at startup and cached.

**Benefit**: Avoids rebuilding ~2KB of JSON schema per iteration. Minor memory saving but reduces GC pressure.

### Use `json.Marshal` Instead of `json.MarshalIndent` for Session Save
**Priority**: Low
**Description**: `SessionManager.Save()` uses `json.MarshalIndent` for pretty-printing. This is ~2x slower than `json.Marshal` and produces larger files on flash storage.

**Benefit**: Faster session saves, smaller session files on limited SPI NAND storage.

### Pre-allocate HTTP Response Buffer
**Priority**: Low
**Description**: `HTTPProvider.Chat()` uses `io.ReadAll(resp.Body)` which starts with a small buffer and grows dynamically. Pre-allocating based on `Content-Length` header (when available) would reduce reallocations.

**Benefit**: Fewer intermediate allocations during LLM response parsing.

## Benchmark Tests

### Add Performance Benchmarks to `make check`
**Priority**: Medium
**Description**: Introduce Go benchmark tests (`func BenchmarkXxx(b *testing.B)`) that measure the performance of critical hot-path functions. These should run as part of `make check` or as a separate `make bench` target. Proposed benchmarks:

1. **`BenchmarkBuildSystemPrompt`** — Measures time to build the full system prompt from disk files. Baseline: should be <5ms.
2. **`BenchmarkBuildMessages`** — Measures context assembly with varying history sizes (10, 50, 100 messages). Guards against regression as history grows.
3. **`BenchmarkSessionSave`** — Measures JSON serialization + atomic write for sessions of varying sizes. Ensures save stays <50ms.
4. **`BenchmarkToProviderDefs`** — Measures tool definition generation. Should be <1ms.
5. **`BenchmarkForceCompression`** — Measures conversation compression performance. Critical for memory-constrained devices.
6. **`BenchmarkGetHistory`** — Measures session history copy for varying message counts. Guards against O(n²) regressions.

**Benefit**: Catches performance regressions early, provides baseline numbers for the Luckfox board, and validates that optimization PRs actually improve performance.

**Blocked by**: Nothing. Can be implemented independently.

## Session Management

### Configurable Summarization Thresholds
**Priority**: Medium
**Description**: Port `SummarizeMessageThreshold` and `SummarizeTokenPercent` from picoclaw upstream into our config struct. Currently hardcoded at 20 messages / 75% of context window in `loop.go`. Making these configurable allows users to tune conversation memory behavior without rebuilding.

**Benefit**: Power users can trade token cost for longer conversation context, or reduce it on very small models.

### Improved Token Estimator
**Priority**: Low
**Description**: Port `utf8.RuneCountInString` with 2.5 chars/token ratio from picoclaw upstream (vs our current `len` with 3 chars/token). More accurate for mixed-language content and CJK text.

**Benefit**: Better context budget estimation, especially for non-English conversations.
