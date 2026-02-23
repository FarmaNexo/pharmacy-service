// internal/shared/common/api_response.go
package common

import (
	"time"

	"github.com/farmanexo/pharmacy-service/internal/shared/constants"
	"github.com/google/uuid"
)

type ApiResponse[T any] struct {
	httpStatus *int  `json:"-"`
	Meta       *Meta `json:"meta"`
	Data       *T    `json:"datos"`
}

type Meta struct {
	Messages      []ResponseMessage `json:"mensajes"`
	IdTransaction string            `json:"idTransaccion"`
	Result        bool              `json:"resultado"`
	Timestamp     string            `json:"timestamp"`
}

type ResponseMessage struct {
	Code    string `json:"codigo"`
	Message string `json:"mensaje"`
	Type    string `json:"tipo"`
}

func NewApiResponse[T any]() *ApiResponse[T] {
	return &ApiResponse[T]{
		Meta: &Meta{
			Messages:      make([]ResponseMessage, 0),
			IdTransaction: uuid.New().String(),
			Result:        true,
			Timestamp:     time.Now().Format("20060102 150405"),
		},
		Data: nil,
	}
}

func (r *ApiResponse[T]) SetHttpStatus(statusCode int) { r.httpStatus = &statusCode }
func (r *ApiResponse[T]) GetHttpStatus() *int           { return r.httpStatus }
func (r *ApiResponse[T]) GetHttpStatusOrDefault(defaultStatus int) int {
	if r.httpStatus != nil {
		return *r.httpStatus
	}
	return defaultStatus
}

func (r *ApiResponse[T]) AddMessage(code constants.MessageCode, message string) {
	r.Meta.AddMessage(string(code), message, string(constants.MessageTypeInformation))
}

func (r *ApiResponse[T]) AddMessageWithType(code constants.MessageCode, message string, messageType constants.MessageType) {
	r.Meta.AddMessage(string(code), message, string(messageType))
}

func (r *ApiResponse[T]) AddError(code constants.MessageCode, message string) {
	r.Meta.AddError(string(code), message)
}

func (r *ApiResponse[T]) AddErrorSimple(message string) {
	r.Meta.AddErrorSimple(message)
}

func (r *ApiResponse[T]) AddMessages(messages []ResponseMessage) {
	r.Meta.Messages = append(r.Meta.Messages, messages...)
	for _, msg := range messages {
		if msg.Type == string(constants.MessageTypeError) {
			r.Meta.Result = false
			break
		}
	}
}

func (r *ApiResponse[T]) AddSuccessMessage() {
	r.Meta.AddMessage(string(constants.CodeSuccess), constants.GetDescription(constants.CodeSuccess), string(constants.MessageTypeInformation))
}

func (r *ApiResponse[T]) SetData(data T) { r.Data = &data }

func (r *ApiResponse[T]) GetData() T {
	if r.Data != nil {
		return *r.Data
	}
	var zero T
	return zero
}

func (r *ApiResponse[T]) IsValid() bool   { return r.Meta != nil && r.Meta.Result }
func (r *ApiResponse[T]) HasErrors() bool  { return !r.IsValid() }

func (r *ApiResponse[T]) IsValidOrSetNewMeta(newResponse interface{}) bool {
	valid := r.IsValid()
	if !valid {
		if v, ok := newResponse.(*ApiResponse[T]); ok {
			v.Meta = r.Meta
		}
	}
	return valid
}

func CreateResponseWithError[T any](code constants.MessageCode, message string) *ApiResponse[T] {
	response := NewApiResponse[T]()
	response.AddError(code, message)
	return response
}

func (m *Meta) AddMessage(code, message, messageType string) {
	m.Messages = append(m.Messages, ResponseMessage{Code: code, Message: message, Type: messageType})
	if messageType == string(constants.MessageTypeError) {
		m.Result = false
	}
}

func (m *Meta) AddError(code, message string) {
	m.Result = false
	m.Messages = append(m.Messages, ResponseMessage{Code: code, Message: message, Type: string(constants.MessageTypeError)})
}

func (m *Meta) AddErrorSimple(message string) {
	m.AddError(string(constants.CodeInternalError), message)
}

func (m *Meta) GetErrors() []ResponseMessage {
	errors := make([]ResponseMessage, 0)
	for _, msg := range m.Messages {
		if msg.Type == string(constants.MessageTypeError) {
			errors = append(errors, msg)
		}
	}
	return errors
}

func (m *Meta) GetFirstError() *ResponseMessage {
	for _, msg := range m.Messages {
		if msg.Type == string(constants.MessageTypeError) {
			return &msg
		}
	}
	return nil
}

func (m *Meta) ClearMessages() {
	m.Messages = make([]ResponseMessage, 0)
	m.Result = true
}
