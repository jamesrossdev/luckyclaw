# Suggested Improvements (Backlog)

Items listed here are planned enhancements that are not yet scheduled for implementation.

## Cron Tool Enhancements

### Add `at_time` Parameter
**Priority**: Medium
**Description**: Add a new `at_time` parameter to the cron tool that accepts an ISO-8601 timestamp (e.g., `"2026-02-22T07:00:00+03:00"`). The tool would internally convert this to `atMS` using `time.Parse(time.RFC3339, at_time)`. This eliminates the need for the LLM to manually calculate `at_seconds` from `time.Now()` when the user specifies an absolute clock time for a one-time reminder.

**Benefit**: Reduces LLM arithmetic errors when converting "at 7:10 AM" → `at_seconds`. Currently the LLM must compute `target_time - current_time` in seconds, which is error-prone. With `at_time`, it just passes the ISO string directly.

**Blocked by**: Nothing. Can be implemented independently after Phase 12-H.



## Performance Optimizations

### Cache System Prompt Between Messages
**Priority**: Medium
**Description**: `BuildSystemPrompt()` in `pkg/agent/context.go` re-reads `SOUL.md`, `USER.md`, `AGENTS.md`, skills summaries, and memory context from disk on every message. These files rarely change. Caching the result with a file-modification-time check would eliminate repeated disk I/O and string allocations.

**Benefit**: Eliminates ~5 file reads and ~10KB of string allocations per message. On the Luckfox's SPI NAND flash (slower than eMMC), this could save 5-10ms per message.


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


### Improved Token Estimator
**Priority**: Low
**Description**: Port `utf8.RuneCountInString` with 2.5 chars/token ratio from picoclaw upstream (vs our current `len` with 3 chars/token). More accurate for mixed-language content and CJK text.

**Benefit**: Better context budget estimation, especially for non-English conversations.

## Context Management

### Pre-emptive Context Compression
**Priority**: Medium
**Description**: Before sending messages to the LLM, estimate the token count and proactively compress history if it exceeds the model's context window (minus a safety margin). Currently, compression only happens AFTER a 400 error from the API.

**Implementation**:
1. Add `EstimateTokens(messages []providers.Message) int` function using character count × ratio
2. In `runAgentLoop`, before calling `provider.Chat()`, check: `if estimatedTokens > (contextWindow * 0.85)` then trigger compression
3. Log the pre-emptive compression for debugging

**Benefit**: Avoids wasted API calls and provides smoother user experience. Currently users see "Context window exceeded" message before compression kicks in.

**Blocked by**: Nothing. Can be implemented independently.

## Skill System

### Channel-Based Skill Filtering
**Priority**: Medium
**Description**: Filter skills by message origin channel to prevent cross-channel skill leakage. Currently, all skills are visible to the LLM regardless of which channel the message came from, causing the LLM to read Discord-specific moderation content when responding to WhatsApp users.

**Implementation**:

1. **Add `channels:` field to SKILL.md frontmatter (YAML)**:
   ```yaml
   ---
   name: discord-mod
   description: Server FAQ, channel directory, and rules
   channels: [discord]
   ---
   ```
   - `channels: [discord]` → only visible on Discord
   - `channels: [whatsapp]` → only visible on WhatsApp
   - No `channels:` field or `channels: [all]` → visible on all channels

2. **Modify `SkillMetadata` struct** in `pkg/skills/loader.go`:
   ```go
   type SkillMetadata struct {
       Name        string   `json:"name"`
       Description string   `json:"description"`
       Channels    []string `json:"channels"` // Optional: channels this skill applies to
   }
   ```

3. **Modify `BuildSkillsSummary()`** in `pkg/skills/loader.go`:
   - Accept `channel string` parameter
   - Filter skills: `skill.Channels == nil || contains(skill.Channels, channel) || contains(skill.Channels, "all")`

4. **Update `BuildSystemPrompt()`** in `pkg/agent/context.go`:
   - Pass `channel` parameter to `BuildSkillsSummary(channel)`

5. **Skill channel assignments** (initial):
   - `discord-mod/SKILL.md` → `channels: [discord]`
   - `whatsapp/SKILL.md` → `channels: [whatsapp]`
   - `weather/SKILL.md` → omitempty (all channels)
   - `summarize/SKILL.md` → omitempty (all channels)
   - `hardware/SKILL.md` → omitempty (all channels)

**Files affected**:
- `workspace/skills/*/SKILL.md` — add `channels:` frontmatter
- `pkg/skills/loader.go` — filter by channel
- `pkg/agent/context.go` — pass channel to skills loader

**Benefit**: Prevents LLM from reading Discord moderation rules when responding to WhatsApp users, and vice versa. Reduces irrelevant context in system prompt, saves tokens.

**Blocked by**: Nothing. Can be implemented independently.

## Process Management

### PID Validation & Stale File Cleanup
**Priority**: Medium
**Description**: Port `pkg/pid` from PicoClaw for robust process management.

PicoClaw adds:
- Token-based PID ownership validation
- Stale PID file detection and cleanup on startup
- Version field for compatibility checking
- Health endpoint includes PID for runtime verification
- Prevents stop/restart when process isn't the gateway

**Implementation**:

1. **Create `pkg/pid/pidfile.go`**:
   ```go
   type PidFileData struct {
       PID     int    `json:"pid"`
       Token   string `json:"token"`   // Cryptographic token for ownership
       Version string `json:"version"`
       Port    int    `json:"port"`
       Host    string `json:"host"`
   }
   
   func WritePidFile(homePath, host string, port int) (*PidFileData, error)
   func ReadPidFile(homePath string) (*PidFileData, error)
   func RemovePidFileIfPID(homePath string, pid int) error
   func IsRunning(data *PidFileData) bool
   ```

2. **Update `cmd/luckyclaw/main.go`**:
   - Replace simple PID file read with validated `ReadPidFile()`
   - Add stale file cleanup on `stopCmd()` and `restartCmd()`
   - Validate PID ownership before sending signals

3. **Update `firmware/overlay/etc/init.d/S99luckyclaw`**:
   - Check if PID file exists and is valid before starting
   - Clean stale PID files as part of start sequence

**RAM Impact**: Minimal (small JSON file)

**Blocked by**: Nothing. Can be implemented independently.

## Extensibility

### Model Context Protocol (MCP) Support
**Priority**: Medium
**Description**: Port MCP support from PicoClaw. MCP allows users to extend LuckyClaw with external tool servers via Anthropic's open standard protocol.

**What it enables**:
- Filesystem access tools
- Database connectors
- Custom API integrations
- Community MCP servers
- Browser automation (via `agent-browser` MCP server)

**Implementation**:

1. **Add `pkg/mcp/` package** (from PicoClaw):
   ```
   pkg/mcp/
   ├── manager.go              // MCP server lifecycle management
   └── isolated_command_transport.go  // Process isolation for MCP servers
   ```

2. **Add MCP config to `config.json`**:
   ```json
   {
     "mcp": {
       "servers": [
         {
           "name": "filesystem",
           "command": "/usr/local/bin/filesystem-server",
           "args": ["--path", "/oem/.luckyclaw/workspace"]
         }
       ]
     }
   }
   ```

3. **Wire MCP tools into agent tool system**:
   - Load MCP tools during gateway startup
   - Convert MCP tool definitions to LuckyClaw tool format
   - Handle MCP JSON-RPC protocol for tool execution

4. **Ship with zero servers by default**:
   - Minimal base overhead (~100KB)
   - Users can add servers as needed via config

**RAM Impact**: Minimal base overhead (~100KB). Scales with connected servers. Not recommended for Pico Plus (64MB RAM). Suitable for Pro/Max.

**Blocked by**: Nothing. Can be implemented independently.

## New Skills

### agent-browser Skill
**Priority**: Low (Pro/Max only)
**Description**: Add browser automation skill from PicoClaw for web scraping and form filling.

**Source**: `picoclaw-latest/workspace/skills/agent-browser/SKILL.md`

**What it enables**:
- Navigate websites
- Fill forms and click buttons
- Take screenshots
- Extract data from web pages
- Test web applications

**Requirements**:
- `agent-browser` npm package: `npm install -g agent-browser`
- Chromium browser (~100MB RAM)

**RAM Impact**: ~100MB+ (Chromium + Node). **Pico Plus (64MB) is NOT sufficient.**

**Files affected**:
- `workspace/skills/agent-browser/SKILL.md` — add from PicoClaw

**Note**: Document RAM requirements in skill. Only recommend for Pro/Max boards.

**Blocked by**: Nothing. Not recommended for Pico Plus.

### GitHub PR Review Skill
**Priority**: Low
**Description**: Add comprehensive PR review agent skill inspired by ZeroClaw's `github-pr-review`.

**Source Inspiration**: `zeroclaw-latest/.claude/skills/github-pr-review/SKILL.md`

**What it enables**:
- Autonomous PR triage and review
- Check CI status and template compliance
- Risk-routed code review depth
- Create worktrees for testing
- Post structured review comments with severity tags
- Never merges (leaves to human maintainer)

**ZeroClaw's workflow phases**:
| Phase | What happens | Key gates |
|-------|-------------|-----------|
| 1. Triage | Read PR, comprehension summary, draft/assignee/path/CI checks | Draft → stop |
| 2. Gate Checks | Malicious scan, template, size, privacy, duplicates | Any gate fail → block/close |
| 3. Review | Risk-routed depth, code review with severity-tagged comments | Format: `[blocking]/[suggestion]/[question]` |
| 4. Final Review | Re-read for changes, handle new commits | Three outcomes: ready/needs-author/needs-maintainer |
| 5. Report | Session report, delete worktree | Every field filled |

**Implementation approach**:
1. Add `workspace/skills/github-pr-review/SKILL.md`
2. Leverage existing `github` skill's `gh` CLI foundation
3. Focus on PR review workflow, not full queue processing

**RAM Impact**: Minimal (uses `gh` CLI, no additional processes)

**Blocked by**: Nothing. Can be implemented independently.

**Reference**: [ZeroClaw PR Review Skill](https://github.com/zeroclaw-labs/zeroclaw/blob/master/.claude/skills/github-pr-review/SKILL.md)
