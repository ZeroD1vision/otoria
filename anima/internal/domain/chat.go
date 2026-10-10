package domain

// --- Работа с LLM ---

type MessageRole uint8

const (
	RoleUnknown MessageRole = iota
	RoleUser
	RoleAssistant
	RoleTools
)

func (r MessageRole) String() string {
	switch r {
	case RoleUser: 		return "User"
	case RoleAssistant: return "Assistant"
	case RoleTools: 	return "Tools"
	default: 			return "Unknown"
	}
}

type MessageHeader struct {
	SessionID SessionID 
	Role 	  MessageRole
}

type ChatMessage struct {
	Header	MessageHeader
	Content []byte
}