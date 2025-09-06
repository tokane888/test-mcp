package response

import "time"

type Error struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func NewError(code, message string) Error {
	return Error{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	}
}
