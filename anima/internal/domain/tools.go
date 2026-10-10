// --- Работа с инструментами ---

// Именованные типы для идентификаторов
type ToolID    uint16

// enum для айди инструментов
const (
	ToolUnknown ToolID = iota
	ToolStop
)

func (t ToolID) Verbose() string {
	switch t {
	case ToolUnknown:
		return "UNKNOWN"
	case ToolStop:
		return "STOP"
	default:
		return "UNKNOWN"
	}
}

type ToolCall struct {
	ID       uint64 // Айди вызова
	ToolID   ToolID // Айди вызванного инструмента
	Payload  []byte // Данные, которые были переданы дополнительно
}

type ToolDefinition struct {
	ID 			uint64 // Айди инструмента
	ToolID  	ToolID // Айди инструмента
	ReturnsData	bool   // Возвращает ли инструмент данные для LLM (Не MVP)
}

type ToolResult struct {
	CallID  uint64 // Айди вызова
	Payload []byte // Данные, которые были возвращены инструментом
	Error   ErrorCode // 1 байт: Код ошибки, если инструмент вернул ошибку. 0 - успех
}

// Работа с ошибками инструментов
type ErrorCode uint8

const (
	CodeOK               ErrorCode = iota // 0: Успешно
	CodeInternalErr                       // 1: Сбой внутри самого инструмента
	CodeInvalidArgs                       // 2: LLM передала невалидные аргументы
	CodeTimeout                           // 3: Внешнее API не ответило вовремя
	CodeNetworkErr                        // 4: Проблемы с сетью
	CodePermissionDenied                  // 5: Нет доступа к ресурсу
)

func (c ErrorCode) String() string {
	switch c {
	case CodeOK: return "OK"
	case CodeInternalErr: return "INTERNAL_ERROR"
	case CodeInvalidArgs: return "INVALID_ARGUMENTS"
	case CodeTimeout: return "TIMEOUT"
	case CodeNetworkErr: return "NETWORK_ERROR"
	case CodePermissionDenied: return "PERMISSION_DENIED"
	default: return "UNKNOWN_ERROR"
	}
}