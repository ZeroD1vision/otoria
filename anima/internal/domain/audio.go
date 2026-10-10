package domain

import "time"

// --- Работа с аппаратным аудио ---

// Именованные типы для идентификаторов
type BatchID   [16]byte
type SessionID [16]byte

// Функции приведения типов
func NewBatchID(b [16]byte) BatchID {
	return BatchID(b)
}
func NewSessionID(b [16]byte) SessionID {
	return SessionID(b)
}

type AudioBatch struct {
	ID        BatchID     // Айди батча
	SessionID SessionID   // Айди сессии конкретного пользователя, к которой относится батч
	CreatedAt time.Time   // Время создания
	Format    AudioFormat // Формат аудио, чтобы расшифровать Payload
	Payload   []byte      // Сырые данные аудио
}

// TODO(blocked): заполнить поля после ответа Саши
type AudioFormat struct {
	Rate           int // Частота
	Channels       int // Каналы
	BytesPerSample int // Байт на семпл
}
