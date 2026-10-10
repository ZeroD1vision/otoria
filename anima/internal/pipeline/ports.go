package pipeline

import (
	"context"

	"github.com/ZeroD1vision/otoria/anima/internal/domain"
)

type STTClient interface {
}

type LLMClient interface {
}

type TTSClient interface {
}

type ToolExecutor interface {
	Execute(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error)
}

type ToolProvider interface {
	AvailableTools(ctx context.Context) ([]domain.ToolDefinition, error)
}