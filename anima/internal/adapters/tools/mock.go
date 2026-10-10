package tools

import (
	"context"

	"github.com/ZeroD1vision/otoria/anima/internal/domain"
)

type MockToolExecutor struct {
	onExecute func(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error) // Сам тест определяет, что тестировать
}

func (m *MockToolExecutor) Execute(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error) {
	if m.onExecute != nil {
		return m.onExecute(ctx, call)
	}
	// Если тест ничего не настроил
	return domain.ToolResult{
		CallID:  call.ID,
		Payload: call.Payload,
		Error:   0,
	}, nil
}