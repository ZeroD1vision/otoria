package mock

import (
	"context"
	"sync"
	"time"

	"github.com/ZeroD1vision/otoria/anima/internal/domain"
)

type MockSink struct {
	SessionID domain.SessionID
	Interval  time.Duration
	Received  []domain.AudioBatch
	mu        sync.Mutex
	Err       error // Чтобы протестировать обработку ошибок
}

func (m *MockSink) Run(ctx context.Context, in <-chan domain.AudioBatch) error {
	if m.Err != nil {
		return m.Err // Проверка обработки ошибки
	}
	ticker := time.NewTicker(m.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err() // Возвращаем ошибку контекста, если он был отменен
		case batch := <-in:
			m.mu.Lock()
			m.Received = append(m.Received, batch)
			m.mu.Unlock()
		}
	}
}
