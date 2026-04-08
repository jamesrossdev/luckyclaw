package tools

import (
	"context"
	"fmt"
	"time"
)

type EventCallback func(channel, chatID string, event EventDetails) error

type EventDetails struct {
	Name            string
	Description     string
	StartTime       int64 // Unix timestamp in UTC
	EndTime         int64 // Unix timestamp in UTC
	LocationName    string
	LocationAddress string
	Latitude        float64
	Longitude       float64
	JoinLink        string
	IsCall          bool
	IsCanceled      bool
}

type ScheduleTool struct {
	eventCallback  EventCallback
	defaultChannel string
	defaultChatID  string
}

func NewScheduleTool() *ScheduleTool {
	return &ScheduleTool{}
}

func (t *ScheduleTool) Name() string {
	return "schedule_event"
}

func (t *ScheduleTool) Description() string {
	return "Schedule a calendar event/appointment that will be sent as a WhatsApp event message. Use this to create calendar events, appointments, or schedule calls with specific times. Events include reminder notifications on WhatsApp."
}

func (t *ScheduleTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "The event/appointment title (e.g., 'Doctor Appointment', 'Team Meeting')",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Optional: Detailed description of the event",
			},
			"start_time": map[string]interface{}{
				"type":        "string",
				"description": "Start time in ISO 8601 format (e.g., '2024-01-15T14:00:00Z' for UTC or '2024-01-15T14:00:00+02:00' for specific timezone)",
			},
			"end_time": map[string]interface{}{
				"type":        "string",
				"description": "Optional: End time in ISO 8601 format. If not provided, defaults to 1 hour after start time",
			},
			"location_name": map[string]interface{}{
				"type":        "string",
				"description": "Optional: Name of the location (e.g., 'City Medical Center', 'Conference Room A')",
			},
			"location_address": map[string]interface{}{
				"type":        "string",
				"description": "Optional: Full address of the location",
			},
			"is_call": map[string]interface{}{
				"type":        "boolean",
				"description": "Optional: Set to true if this is a video/voice call (default: false)",
			},
			"join_link": map[string]interface{}{
				"type":        "string",
				"description": "Optional: Meeting link for video/voice calls (Zoom, Google Meet, etc.)",
			},
		},
		"required": []string{"name", "start_time"},
	}
}

func (t *ScheduleTool) SetContext(channel, chatID string) {
	t.defaultChannel = channel
	t.defaultChatID = chatID
}

func (t *ScheduleTool) SetEventCallback(callback EventCallback) {
	t.eventCallback = callback
}

func (t *ScheduleTool) Execute(ctx context.Context, args map[string]interface{}) *ToolResult {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return &ToolResult{ForLLM: "name is required", IsError: true}
	}

	startTimeStr, ok := args["start_time"].(string)
	if !ok || startTimeStr == "" {
		return &ToolResult{ForLLM: "start_time is required", IsError: true}
	}

	// Parse start time
	startTime, err := parseTime(startTimeStr)
	if err != nil {
		return &ToolResult{ForLLM: fmt.Sprintf("invalid start_time format: %v", err), IsError: true}
	}

	// Parse end time (optional, defaults to 1 hour after start)
	var endTime int64
	if endTimeStr, ok := args["end_time"].(string); ok && endTimeStr != "" {
		parsed, err := parseTime(endTimeStr)
		if err != nil {
			return &ToolResult{ForLLM: fmt.Sprintf("invalid end_time format: %v", err), IsError: true}
		}
		endTime = parsed
	} else {
		endTime = startTime + 3600 // Default 1 hour duration
	}

	event := EventDetails{
		Name:            name,
		Description:     getStringArg(args, "description"),
		StartTime:       startTime,
		EndTime:         endTime,
		LocationName:    getStringArg(args, "location_name"),
		LocationAddress: getStringArg(args, "location_address"),
		JoinLink:        getStringArg(args, "join_link"),
		IsCall:          getBoolArg(args, "is_call"),
	}

	if t.eventCallback == nil {
		return &ToolResult{ForLLM: "event scheduling not configured", IsError: true}
	}

	channel := t.defaultChannel
	chatID := t.defaultChatID

	if err := t.eventCallback(channel, chatID, event); err != nil {
		return &ToolResult{
			ForLLM:  fmt.Sprintf("failed to schedule event: %v", err),
			IsError: true,
			Err:     err,
		}
	}

	startFormatted := time.Unix(startTime, 0).Format("Jan 2, 2006 at 3:04 PM")
	return &ToolResult{
		ForLLM: fmt.Sprintf("Event '%s' scheduled for %s", name, startFormatted),
	}
}

func parseTime(timeStr string) (int64, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, timeStr); err == nil {
			return t.Unix(), nil
		}
	}
	return 0, fmt.Errorf("unable to parse time: %s", timeStr)
}

func getStringArg(args map[string]interface{}, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

func getBoolArg(args map[string]interface{}, key string) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return false
}
