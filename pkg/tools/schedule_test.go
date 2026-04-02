package tools

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestScheduleTool_Execute_Success(t *testing.T) {
	tool := NewScheduleTool()
	tool.SetContext("whatsapp", "123456789")

	var capturedChannel, capturedChatID string
	var capturedEvent EventDetails

	tool.SetEventCallback(func(channel, chatID string, event EventDetails) error {
		capturedChannel = channel
		capturedChatID = chatID
		capturedEvent = event
		return nil
	})

	ctx := context.Background()
	args := map[string]interface{}{
		"name":          "Doctor Appointment",
		"start_time":    "2024-01-15T14:00:00Z",
		"end_time":      "2024-01-15T14:30:00Z",
		"location_name": "City Medical Center",
	}

	result := tool.Execute(ctx, args)

	if capturedChannel != "whatsapp" {
		t.Errorf("Expected channel 'whatsapp', got '%s'", capturedChannel)
	}
	if capturedChatID != "123456789" {
		t.Errorf("Expected chatID '123456789', got '%s'", capturedChatID)
	}
	if capturedEvent.Name != "Doctor Appointment" {
		t.Errorf("Expected event name 'Doctor Appointment', got '%s'", capturedEvent.Name)
	}
	if capturedEvent.LocationName != "City Medical Center" {
		t.Errorf("Expected location 'City Medical Center', got '%s'", capturedEvent.LocationName)
	}

	if result.IsError {
		t.Errorf("Expected no error, got: %s", result.ForLLM)
	}
}

func TestScheduleTool_Execute_VideoCall(t *testing.T) {
	tool := NewScheduleTool()
	tool.SetContext("whatsapp", "123456789")

	var capturedEvent EventDetails

	tool.SetEventCallback(func(channel, chatID string, event EventDetails) error {
		capturedEvent = event
		return nil
	})

	ctx := context.Background()
	args := map[string]interface{}{
		"name":       "Team Standup",
		"start_time": "2024-01-15T09:00:00Z",
		"is_call":    true,
		"join_link":  "https://zoom.us/j/123456789",
	}

	result := tool.Execute(ctx, args)

	if !capturedEvent.IsCall {
		t.Error("Expected IsCall to be true")
	}
	if capturedEvent.JoinLink != "https://zoom.us/j/123456789" {
		t.Errorf("Expected join link 'https://zoom.us/j/123456789', got '%s'", capturedEvent.JoinLink)
	}

	if result.IsError {
		t.Errorf("Expected no error, got: %s", result.ForLLM)
	}
}

func TestScheduleTool_Execute_MissingName(t *testing.T) {
	tool := NewScheduleTool()
	tool.SetContext("whatsapp", "123456789")

	ctx := context.Background()
	args := map[string]interface{}{
		"start_time": "2024-01-15T14:00:00Z",
	}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error for missing name")
	}
	if result.ForLLM != "name is required" {
		t.Errorf("Expected 'name is required', got '%s'", result.ForLLM)
	}
}

func TestScheduleTool_Execute_MissingStartTime(t *testing.T) {
	tool := NewScheduleTool()
	tool.SetContext("whatsapp", "123456789")

	ctx := context.Background()
	args := map[string]interface{}{
		"name": "Test Event",
	}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error for missing start_time")
	}
	if result.ForLLM != "start_time is required" {
		t.Errorf("Expected 'start_time is required', got '%s'", result.ForLLM)
	}
}

func TestScheduleTool_Execute_InvalidTimeFormat(t *testing.T) {
	tool := NewScheduleTool()
	tool.SetContext("whatsapp", "123456789")

	ctx := context.Background()
	args := map[string]interface{}{
		"name":       "Test Event",
		"start_time": "not-a-valid-time",
	}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error for invalid time format")
	}
}

func TestScheduleTool_Execute_NoCallback(t *testing.T) {
	tool := NewScheduleTool()
	tool.SetContext("whatsapp", "123456789")
	// No SetEventCallback called

	ctx := context.Background()
	args := map[string]interface{}{
		"name":       "Test Event",
		"start_time": "2024-01-15T14:00:00Z",
	}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error when callback not configured")
	}
	if result.ForLLM != "event scheduling not configured" {
		t.Errorf("Expected 'event scheduling not configured', got '%s'", result.ForLLM)
	}
}

func TestScheduleTool_Execute_CallbackError(t *testing.T) {
	tool := NewScheduleTool()
	tool.SetContext("whatsapp", "123456789")

	expectedErr := errors.New("failed to send event")
	tool.SetEventCallback(func(channel, chatID string, event EventDetails) error {
		return expectedErr
	})

	ctx := context.Background()
	args := map[string]interface{}{
		"name":       "Test Event",
		"start_time": "2024-01-15T14:00:00Z",
	}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error when callback returns error")
	}
	if result.Err != expectedErr {
		t.Errorf("Expected original error to be preserved, got %v", result.Err)
	}
}

func TestScheduleTool_Execute_DefaultEndTime(t *testing.T) {
	tool := NewScheduleTool()
	tool.SetContext("whatsapp", "123456789")

	var capturedEvent EventDetails

	tool.SetEventCallback(func(channel, chatID string, event EventDetails) error {
		capturedEvent = event
		return nil
	})

	ctx := context.Background()
	startTime := time.Now().Unix()
	startTimeStr := time.Unix(startTime, 0).Format(time.RFC3339)

	args := map[string]interface{}{
		"name":       "Test Event",
		"start_time": startTimeStr,
		// No end_time provided
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Unexpected error: %s", result.ForLLM)
	}

	// End time should default to start + 1 hour
	expectedEndTime := startTime + 3600
	if capturedEvent.EndTime != expectedEndTime {
		t.Errorf("Expected end time %d, got %d", expectedEndTime, capturedEvent.EndTime)
	}
}

func TestScheduleTool_Name(t *testing.T) {
	tool := NewScheduleTool()
	if tool.Name() != "schedule_event" {
		t.Errorf("Expected name 'schedule_event', got '%s'", tool.Name())
	}
}

func TestScheduleTool_Description(t *testing.T) {
	tool := NewScheduleTool()
	desc := tool.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
}

func TestScheduleTool_Parameters(t *testing.T) {
	tool := NewScheduleTool()
	params := tool.Parameters()

	typ, ok := params["type"].(string)
	if !ok || typ != "object" {
		t.Error("Expected type 'object'")
	}

	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	required, ok := params["required"].([]string)
	if !ok || len(required) != 2 || required[0] != "name" || required[1] != "start_time" {
		t.Error("Expected 'name' and 'start_time' to be required")
	}

	if _, ok := props["name"]; !ok {
		t.Error("Expected 'name' property")
	}
	if _, ok := props["start_time"]; !ok {
		t.Error("Expected 'start_time' property")
	}
	if _, ok := props["is_call"]; !ok {
		t.Error("Expected 'is_call' property")
	}
	if _, ok := props["join_link"]; !ok {
		t.Error("Expected 'join_link' property")
	}
}

func TestScheduleTool_TimeFormats(t *testing.T) {
	tool := NewScheduleTool()
	tool.SetContext("whatsapp", "123456789")

	tool.SetEventCallback(func(channel, chatID string, event EventDetails) error {
		_ = event // captured but not used in this test
		return nil
	})

	testCases := []struct {
		name      string
		timeStr   string
		shouldErr bool
	}{
		{"RFC3339", "2024-01-15T14:00:00Z", false},
		{"ISO variant 1", "2024-01-15T15:04:05Z", false},
		{"ISO variant 2", "2024-01-15T15:04:05", false},
		{"Date only", "2024-01-15", false},
		{"Invalid", "not-a-time", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			args := map[string]interface{}{
				"name":       "Test Event",
				"start_time": tc.timeStr,
			}

			result := tool.Execute(ctx, args)

			if tc.shouldErr && !result.IsError {
				t.Errorf("Expected error for time format: %s", tc.timeStr)
			}
			if !tc.shouldErr && result.IsError {
				t.Errorf("Unexpected error for time format %s: %s", tc.timeStr, result.ForLLM)
			}
		})
	}
}
