package tools

import "context"

// Tool is the interface that all tools must implement.
type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]interface{}
	Execute(ctx context.Context, args map[string]interface{}) *ToolResult
}

// --- Request-scoped tool context (channel / chatID) ---
//
// Carried via context.Value so that concurrent tool calls each receive
// their own immutable copy — no mutable state on singleton tool instances.

type toolCtxKey struct{ name string }

var (
	ctxKeyChannel = &toolCtxKey{"channel"}
	ctxKeyChatID  = &toolCtxKey{"chatID"}
)

// WithToolContext returns a child context carrying channel and chatID.
func WithToolContext(ctx context.Context, channel, chatID string) context.Context {
	ctx = context.WithValue(ctx, ctxKeyChannel, channel)
	ctx = context.WithValue(ctx, ctxKeyChatID, chatID)
	return ctx
}

// ToolChannel extracts the channel from ctx, or "" if unset.
func ToolChannel(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyChannel).(string)
	return v
}

// ToolChatID extracts the chatID from ctx, or "" if unset.
func ToolChatID(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyChatID).(string)
	return v
}

// ContextualTool is an optional interface that tools can implement
// to receive the current message context (channel, chatID).
// @deprecated: Use ToolChannel(ctx) and ToolChatID(ctx) instead.
type ContextualTool interface {
	Tool
	SetContext(channel, chatID string)
}

// AsyncCallback is a function type that async tools use to notify completion.
type AsyncCallback func(ctx context.Context, result *ToolResult)

// AsyncTool is an optional interface that tools can implement to support
// asynchronous execution with completion callbacks.
// @deprecated: Use AsyncExecutor instead for thread-safety.
type AsyncTool interface {
	Tool
	SetCallback(cb AsyncCallback)
}

// AsyncExecutor is the modern interface for async tool execution.
// It receives the callback as a parameter, avoiding shared state races.
type AsyncExecutor interface {
	Tool
	ExecuteAsync(ctx context.Context, args map[string]interface{}, cb AsyncCallback) *ToolResult
}

func ToolToSchema(tool Tool) map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        tool.Name(),
			"description": tool.Description(),
			"parameters":  tool.Parameters(),
		},
	}
}
