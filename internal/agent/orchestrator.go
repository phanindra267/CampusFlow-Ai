package agent

import (
	"context"
	"fmt"
	"strings"
)

type Orchestrator struct {
	registry *Registry
}

func NewOrchestrator(registry *Registry) *Orchestrator {
	return &Orchestrator{registry: registry}
}

// ExecuteTask runs a basic ReAct loop simulation
func (o *Orchestrator) ExecuteTask(ctx context.Context, prompt string) (string, error) {
	// 1. Parse intent (Simulated LLM call)
	intent := parseIntent(prompt)

	// 2. Select Tool
	var result string
	var err error
	
	switch intent {
	case "check_attendance":
		result, err = o.registry.Execute(ctx, "GetAttendanceTool", map[string]interface{}{})
	case "find_events":
		result, err = o.registry.Execute(ctx, "GetEventsTool", map[string]interface{}{})
	default:
		result = "I couldn't determine the appropriate tool for that request."
	}

	if err != nil {
		return "", fmt.Errorf("tool execution failed: %w", err)
	}

	// 3. Format Response (Simulated LLM synthesis)
	return fmt.Sprintf("Based on the data retrieved: %s", result), nil
}

func parseIntent(prompt string) string {
	p := strings.ToLower(prompt)
	if strings.Contains(p, "attendance") {
		return "check_attendance"
	}
	if strings.Contains(p, "event") || strings.Contains(p, "hackathon") {
		return "find_events"
	}
	return "unknown"
}