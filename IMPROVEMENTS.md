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

**Blocked by**: Needs to be submitted as a PR to the openclaw upstream repository.

## Installation / Deployment

### Open-Source Cross-Platform Flashing Tool
**Priority**: Low
**Description**: Currently, flashing the eMMC requires using the proprietary Rockchip `SOCToolKit`, which is Windows-only. We should develop or adopt an open-source, cross-platform CLI tool (e.g., in Python or Go) that can communicate with the Rockchip MaskROM protocol to flash `update.img` directly from Linux and macOS without needing Windows VMs or proprietary software.

**Benefit**: Dramatically simplifies the onboarding process for non-Windows users and allows for scripted/automated deployments.
**Blocked by**: Reverse engineering of Rockchip protocols or integrating existing open-source alternatives like `rkdeveloptool`.
