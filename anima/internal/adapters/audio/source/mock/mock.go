package mock

import (
	"context"
	"time"

	"github.com/ZeroD1vision/otoria/anima/internal/domain"
	"github.com/google/uuid"
)

type MockSource struct {
	SessionID domain.SessionID
	Interval  time.Duration
	Payload   []byte
	Err       error // Чтобы протестировать обработку ошибок
}

func (m *MockSource) Run(ctx context.Context, out chan<- domain.AudioBatch) error {
	if m.Err != nil {
		return m.Err // Проверка обработки ошибки
	}
	ticker := time.NewTicker(m.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err() // Возвращаем ошибку контекста, если он был отменен
		case <-ticker.C:
			batchUUID := uuid.New()
			sessionUUID := uuid.New()
			out <- domain.AudioBatch{
				ID:        domain.NewBatchID(batchUUID), // Генерируем UUID, решили заранее перейти на него вместо uint64
				SessionID: domain.NewSessionID(sessionUUID), // Генерируем UUID для сессии
				CreatedAt: time.Now(),
				Payload:   m.Payload,
			}
		}
	}
}
