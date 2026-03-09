// LuckyClaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 LuckyClaw contributors

package heartbeat

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jamesrossdev/luckyclaw/pkg/bus"
	"github.com/jamesrossdev/luckyclaw/pkg/constants"
	"github.com/jamesrossdev/luckyclaw/pkg/logger"
	"github.com/jamesrossdev/luckyclaw/pkg/state"
	"github.com/jamesrossdev/luckyclaw/pkg/tools"
)

const (
	minIntervalMinutes     = 5
	defaultIntervalMinutes = 30
)

// HeartbeatHandler is the function type for handling heartbeat.
// It returns a ToolResult that can indicate async operations.
// channel and chatID are derived from the last active user channel.
type HeartbeatHandler func(prompt, channel, chatID string) *tools.ToolResult

// HeartbeatService manages periodic heartbeat checks
type HeartbeatService struct {
	workspace string
	bus       *bus.MessageBus
	state     *state.Manager
	handler   HeartbeatHandler
	interval  time.Duration
	enabled   bool
	mu        sync.RWMutex
	stopChan  chan struct{}
}

// NewHeartbeatService creates a new heartbeat service
func NewHeartbeatService(workspace string, intervalMinutes int, enabled bool) *HeartbeatService {
	// Apply minimum interval
	if intervalMinutes < minIntervalMinutes && intervalMinutes != 0 {
		intervalMinutes = minIntervalMinutes
	}

	if intervalMinutes == 0 {
		intervalMinutes = defaultIntervalMinutes
	}

	return &HeartbeatService{
		workspace: workspace,
		interval:  time.Duration(intervalMinutes) * time.Minute,
		enabled:   enabled,
		state:     state.NewManager(workspace),
	}
}

// SetBus sets the message bus for delivering heartbeat results.
func (hs *HeartbeatService) SetBus(msgBus *bus.MessageBus) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.bus = msgBus
}

// SetHandler sets the heartbeat handler.
func (hs *HeartbeatService) SetHandler(handler HeartbeatHandler) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.handler = handler
}

// Start begins the heartbeat service
func (hs *HeartbeatService) Start() error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	if hs.stopChan != nil {
		logger.InfoC("heartbeat", "Heartbeat service already running")
		return nil
	}

	if !hs.enabled {
		logger.InfoC("heartbeat", "Heartbeat service disabled")
		return nil
	}

	hs.stopChan = make(chan struct{})
	go hs.runLoop(hs.stopChan)

	logger.InfoCF("heartbeat", "Heartbeat service started", map[string]any{
		"interval_minutes": hs.interval.Minutes(),
	})

	return nil
}

// Stop gracefully stops the heartbeat service
func (hs *HeartbeatService) Stop() {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	if hs.stopChan == nil {
		return
	}

	logger.InfoC("heartbeat", "Stopping heartbeat service")
	close(hs.stopChan)
	hs.stopChan = nil
}

// IsRunning returns whether the service is running
func (hs *HeartbeatService) IsRunning() bool {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.stopChan != nil
}

// runLoop runs the heartbeat ticker
func (hs *HeartbeatService) runLoop(stopChan chan struct{}) {
	ticker := time.NewTicker(hs.interval)
	defer ticker.Stop()

	// Run first heartbeat after initial delay
	time.AfterFunc(time.Second, func() {
		hs.executeHeartbeat()
	})

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			hs.executeHeartbeat()
		}
	}
}

// executeHeartbeat performs a single heartbeat check
func (hs *HeartbeatService) executeHeartbeat() {
	hs.mu.RLock()
	enabled := hs.enabled
	handler := hs.handler
	if !hs.enabled || hs.stopChan == nil {
		hs.mu.RUnlock()
		logger.InfoC("heartbeat", "[AUDIT] executeHeartbeat: skipped (disabled or stopChan nil)")
		return
	}
	hs.mu.RUnlock()

	if !enabled {
		logger.InfoC("heartbeat", "[AUDIT] executeHeartbeat: skipped (not enabled)")
		return
	}

	logger.InfoC("heartbeat", "[AUDIT] executeHeartbeat: START")

	prompt := hs.buildPrompt()
	if prompt == "" {
		logger.InfoC("heartbeat", "[AUDIT] executeHeartbeat: empty prompt, aborting")
		hs.logInfo("No heartbeat prompt (HEARTBEAT.md empty or missing)")
		return
	}

	if handler == nil {
		logger.InfoC("heartbeat", "[AUDIT] executeHeartbeat: handler nil, aborting")
		hs.logError("Heartbeat handler not configured")
		return
	}

	// Get last channel info for context
	lastChannel := hs.state.GetLastChannel()
	channel, chatID := hs.parseLastChannel(lastChannel)

	// Debug log for channel resolution
	logger.InfoCF("heartbeat", "[AUDIT] Pre-handler", map[string]interface{}{
		"channel": channel, "chatID": chatID, "lastChannel": lastChannel,
	})
	hs.logInfo("Resolved channel: %s, chatID: %s (from lastChannel: %s)", channel, chatID, lastChannel)

	result := handler(prompt, channel, chatID)

	if result == nil {
		logger.InfoC("heartbeat", "[AUDIT] Post-handler: result is nil")
		hs.logInfo("Heartbeat handler returned nil result")
		return
	}

	// AUDIT: log every field of the result
	logger.InfoCF("heartbeat", "[AUDIT] Post-handler result", map[string]interface{}{
		"Silent":  result.Silent,
		"IsError": result.IsError,
		"Async":   result.Async,
		"ForLLM":  truncateForLog(result.ForLLM, 150),
		"ForUser": truncateForLog(result.ForUser, 150),
	})

	// Handle different result types
	if result.IsError {
		hs.logError("Heartbeat error: %s", result.ForLLM)
		return
	}

	if result.Async {
		hs.logInfo("Async task started: %s", result.ForLLM)
		return
	}

	// Check if silent
	if result.Silent {
		logger.InfoC("heartbeat", "[AUDIT] DROPPED: result.Silent=true — NOT sending to user")
		hs.logInfo("Heartbeat OK - silent")
		return
	}

	// Filter out the silent "HEARTBEAT_OK" acknowledgment
	content := result.ForUser
	if content == "" {
		content = result.ForLLM
	}

	if strings.TrimSpace(content) == "HEARTBEAT_OK" {
		logger.InfoC("heartbeat", "[AUDIT] DROPPED: exact HEARTBEAT_OK match — NOT sending to user")
		hs.logInfo("Heartbeat OK - normal metrics, silent drop")
		return
	}

	// LEAK PATH: if we reach here, the message WILL be sent to the user
	logger.WarnCF("heartbeat", "[AUDIT] LEAK: Heartbeat message reaching sendResponse!", map[string]interface{}{
		"content": truncateForLog(content, 200),
		"Silent":  result.Silent,
		"ForUser": truncateForLog(result.ForUser, 100),
		"ForLLM":  truncateForLog(result.ForLLM, 100),
	})

	// Send result to user
	if content != "" {
		hs.sendResponse(content)
	}

	hs.logInfo("Heartbeat completed: %s", content)
}

// buildPrompt builds the heartbeat prompt from HEARTBEAT.md
func (hs *HeartbeatService) buildPrompt() string {
	heartbeatPath := filepath.Join(hs.workspace, "HEARTBEAT.md")

	data, err := os.ReadFile(heartbeatPath)
	if err != nil {
		if os.IsNotExist(err) {
			hs.createDefaultHeartbeatTemplate()
			return ""
		}
		hs.logError("Error reading HEARTBEAT.md: %v", err)
		return ""
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return ""
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	diskUsage := getDiskUsage()
	diskStatus := fmt.Sprintf("%.1f%%", diskUsage)
	if diskUsage < 95.0 {
		diskStatus = "Normal (Under 95%)"
	} else {
		diskStatus = fmt.Sprintf("CRITICAL - %.1f%% used", diskUsage)
	}

	return fmt.Sprintf(`# Heartbeat Check

Current time: %s
System Disk Status: %s

You are a proactive AI assistant. This is a scheduled heartbeat check.
Review the following tasks and execute any necessary actions using available skills.

CRITICAL INSTRUCTION: When ALL of the following are true, respond with ONLY the exact text HEARTBEAT_OK — nothing else, no extra information, no status summary:
  1. System Status is Normal (disk, memory, network all healthy)
  2. No tasks in HEARTBEAT.md require execution today
  3. You have NOT executed any tools during this check

If ANY issue, alert, anomaly, or task result needs reporting, do NOT include HEARTBEAT_OK anywhere in your response. Write a concise report instead.

%s
`, now, diskStatus, content)
}

// createDefaultHeartbeatTemplate creates the default HEARTBEAT.md file
func (hs *HeartbeatService) createDefaultHeartbeatTemplate() {
	heartbeatPath := filepath.Join(hs.workspace, "HEARTBEAT.md")

	defaultContent := `# Heartbeat Tasks

Execute ALL tasks below every heartbeat cycle. Use shell commands for local data — do NOT waste API tokens on info available locally.

## 1. Time & Date (local — use shell)
- Run: ` + "`date '+%A, %B %d %Y — %I:%M %p %Z'`" + `
- Note any upcoming reminders from memory files

## 2. Device Health (local — use shell)
- Run: ` + "`free -m | grep Mem`" + ` — report available memory
- Run: ` + "`uptime`" + ` — report uptime and load
- If available memory < 5MB, warn the user immediately

## 3. Network (local — use shell)
- Run: ` + "`ping -c 1 -W 2 8.8.8.8 > /dev/null 2>&1 && echo \"Online\" || echo \"OFFLINE\"`" + `
- If offline, alert the user

## Instructions
- Use shell tool for ALL tasks above — they are local system checks
- Keep responses brief — one line per task max
- Only respond with HEARTBEAT_OK after ALL tasks are complete and nothing needs attention
- If any task shows a problem, flag it clearly

---

Add your heartbeat tasks below this line:
`

	if err := os.WriteFile(heartbeatPath, []byte(defaultContent), 0644); err != nil {
		hs.logError("Failed to create default HEARTBEAT.md: %v", err)
	} else {
		hs.logInfo("Created default HEARTBEAT.md template")
	}
}

// sendResponse sends the heartbeat response to the last channel
func (hs *HeartbeatService) sendResponse(response string) {
	hs.mu.RLock()
	msgBus := hs.bus
	hs.mu.RUnlock()

	if msgBus == nil {
		hs.logInfo("No message bus configured, heartbeat result not sent")
		return
	}

	// Get last channel from state
	lastChannel := hs.state.GetLastChannel()
	if lastChannel == "" {
		hs.logInfo("No last channel recorded, heartbeat result not sent")
		return
	}

	platform, userID := hs.parseLastChannel(lastChannel)

	// Skip internal channels that can't receive messages
	if platform == "" || userID == "" {
		return
	}

	msgBus.PublishOutbound(bus.OutboundMessage{
		Channel: platform,
		ChatID:  userID,
		Content: response,
	})

	hs.logInfo("Heartbeat result sent to %s", platform)
}

// parseLastChannel parses the last channel string into platform and userID.
// Returns empty strings for invalid or internal channels.
func (hs *HeartbeatService) parseLastChannel(lastChannel string) (platform, userID string) {
	if lastChannel == "" {
		return "", ""
	}

	// Parse channel format: "platform:user_id" (e.g., "telegram:123456")
	parts := strings.SplitN(lastChannel, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		hs.logError("Invalid last channel format: %s", lastChannel)
		return "", ""
	}

	platform, userID = parts[0], parts[1]

	// Skip internal channels
	if constants.IsInternalChannel(platform) {
		hs.logInfo("Skipping internal channel: %s", platform)
		return "", ""
	}

	return platform, userID
}

// logInfo logs an informational message to the heartbeat log
func (hs *HeartbeatService) logInfo(format string, args ...any) {
	hs.log("INFO", format, args...)
}

// logError logs an error message to the heartbeat log
func (hs *HeartbeatService) logError(format string, args ...any) {
	hs.log("ERROR", format, args...)
}

// log writes a message to the heartbeat log file, with stderr fallback
func (hs *HeartbeatService) log(level, format string, args ...any) {
	logFile := filepath.Join(hs.workspace, "heartbeat.log")
	message := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, message)

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		// Fallback: log to stderr and the structured logger so we don't silently lose entries
		logger.WarnCF("heartbeat", "Failed to write heartbeat.log", map[string]interface{}{
			"error":   err.Error(),
			"message": message,
		})
		return
	}
	defer f.Close()

	fmt.Fprint(f, line)
}

// truncateForLog truncates a string for safe logging
func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
