package tools

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jamesrossdev/luckyclaw/pkg/logger"
	"github.com/jamesrossdev/luckyclaw/pkg/providers"
)

type ToolEntry struct {
	Tool   Tool
	IsCore bool
	TTL    int
}

type ToolRegistry struct {
	tools   map[string]*ToolEntry
	mu      sync.RWMutex
	version atomic.Uint64
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]*ToolEntry),
	}
}

func (r *ToolRegistry) Register(tool Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := tool.Name()
	r.tools[name] = &ToolEntry{
		Tool:   tool,
		IsCore: true,
		TTL:    0,
	}
	r.version.Add(1)
}

func (r *ToolRegistry) RegisterHidden(tool Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := tool.Name()
	r.tools[name] = &ToolEntry{
		Tool:   tool,
		IsCore: false,
		TTL:    0,
	}
	r.version.Add(1)
}

func (r *ToolRegistry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entry, ok := r.tools[name]
	if !ok {
		return nil, false
	}
	if !entry.IsCore && entry.TTL <= 0 {
		return nil, false
	}
	return entry.Tool, true
}

func (r *ToolRegistry) Execute(ctx context.Context, name string, args map[string]interface{}) *ToolResult {
	return r.ExecuteWithContext(ctx, name, args, "", "", nil)
}

func (r *ToolRegistry) ExecuteWithContext(
	ctx context.Context,
	name string,
	args map[string]interface{},
	channel, chatID string,
	asyncCallback AsyncCallback,
) *ToolResult {
	logger.InfoCF("tool", "Tool execution started",
		map[string]interface{}{
			"tool": name,
			"args": args,
		})

	tool, ok := r.Get(name)
	if !ok {
		logger.ErrorCF("tool", "Tool not found",
			map[string]interface{}{
				"tool": name,
			})
		return ErrorResult(fmt.Sprintf("tool %q not found", name)).WithError(fmt.Errorf("tool not found"))
	}

	// Inject context
	ctx = WithToolContext(ctx, channel, chatID)

	// Backward compatibility for ContextualTool
	if ct, ok := tool.(ContextualTool); ok && channel != "" && chatID != "" {
		ct.SetContext(channel, chatID)
	}

	var result *ToolResult
	start := time.Now()

	// Panic recovery
	func() {
		defer func() {
			if re := recover(); re != nil {
				errMsg := fmt.Sprintf("Tool '%s' crashed with panic: %v", name, re)
				logger.ErrorCF("tool", "Tool execution panic recovered",
					map[string]interface{}{
						"tool":  name,
						"panic": fmt.Sprintf("%v", re),
					})
				result = &ToolResult{
					ForLLM:  errMsg,
					ForUser: errMsg,
					IsError: true,
					Err:     fmt.Errorf("panic: %v", re),
				}
			}
		}()

		if asyncExec, ok := tool.(AsyncExecutor); ok && asyncCallback != nil {
			result = asyncExec.ExecuteAsync(ctx, args, asyncCallback)
		} else if asyncTool, ok := tool.(AsyncTool); ok && asyncCallback != nil {
			// Backward compatibility for AsyncTool
			asyncTool.SetCallback(asyncCallback)
			result = tool.Execute(ctx, args)
		} else {
			result = tool.Execute(ctx, args)
		}
	}()

	if result == nil {
		errMsg := fmt.Sprintf("Tool '%s' returned nil result unexpectedly", name)
		result = &ToolResult{
			ForLLM:  errMsg,
			ForUser: errMsg,
			IsError: true,
			Err:     fmt.Errorf("nil result"),
		}
	}

	duration := time.Since(start)
	if result.IsError {
		logger.ErrorCF("tool", "Tool execution failed", map[string]interface{}{"tool": name, "error": result.ForLLM})
	} else {
		logger.InfoCF("tool", "Tool execution completed", map[string]interface{}{"tool": name, "duration_ms": duration.Milliseconds()})
	}

	return result
}

func (r *ToolRegistry) sortedToolNames() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *ToolRegistry) GetDefinitions() []map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sorted := r.sortedToolNames()
	definitions := make([]map[string]interface{}, 0, len(sorted))
	for _, name := range sorted {
		entry := r.tools[name]
		if !entry.IsCore && entry.TTL <= 0 {
			continue
		}
		definitions = append(definitions, ToolToSchema(entry.Tool))
	}
	return definitions
}

func (r *ToolRegistry) ToProviderDefs() []providers.ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sorted := r.sortedToolNames()
	definitions := make([]providers.ToolDefinition, 0, len(sorted))
	for _, name := range sorted {
		entry := r.tools[name]
		if !entry.IsCore && entry.TTL <= 0 {
			continue
		}
		schema := ToolToSchema(entry.Tool)
		fn, ok := schema["function"].(map[string]interface{})
		if !ok {
			continue
		}
		definitions = append(definitions, providers.ToolDefinition{
			Type: "function",
			Function: providers.ToolFunctionDefinition{
				Name:        fn["name"].(string),
				Description: fn["description"].(string),
				Parameters:  fn["parameters"].(map[string]interface{}),
			},
		})
	}
	return definitions
}

func (r *ToolRegistry) Clone() *ToolRegistry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	clone := &ToolRegistry{
		tools: make(map[string]*ToolEntry, len(r.tools)),
	}
	for name, entry := range r.tools {
		clone.tools[name] = &ToolEntry{
			Tool:   entry.Tool,
			IsCore: entry.IsCore,
			TTL:    entry.TTL,
		}
	}
	return clone
}

func (r *ToolRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.sortedToolNames()
}

func (r *ToolRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

func (r *ToolRegistry) GetSummaries() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sorted := r.sortedToolNames()
	summaries := make([]string, 0, len(sorted))
	for _, name := range sorted {
		entry := r.tools[name]
		if !entry.IsCore && entry.TTL <= 0 {
			continue
		}
		summaries = append(summaries, fmt.Sprintf("- `%s` - %s", entry.Tool.Name(), entry.Tool.Description()))
	}
	return summaries
}
