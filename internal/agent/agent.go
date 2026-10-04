package agent

import (
	"context"
	"errors"
)

type Tool func(ctx context.Context, args map[string]interface{}) (string, error)

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

func (r *Registry) Register(name string, t Tool) {
	r.tools[name] = t
}

func (r *Registry) Execute(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	t, ok := r.tools[name]
	if !ok {
		return "", errors.New("tool not found")
	}
	return t(ctx, args)
}

// Example tool
func GetAttendanceTool(ctx context.Context, args map[string]interface{}) (string, error) {
	return "Mock attendance: 85% for ML, 78% for OS.", nil
}