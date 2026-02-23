// internal/shared/constants/message_types.go
package constants

type MessageType string

const (
	MessageTypeInformation MessageType = "INFORMATION"
	MessageTypeWarning     MessageType = "WARNING"
	MessageTypeError       MessageType = "ERROR"
	MessageTypeSuccess     MessageType = "SUCCESS"
)

func (mt MessageType) IsValid() bool {
	switch mt {
	case MessageTypeInformation, MessageTypeWarning, MessageTypeError, MessageTypeSuccess:
		return true
	}
	return false
}

func (mt MessageType) String() string {
	return string(mt)
}
