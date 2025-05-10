package models

import "fmt"

type TaskErrorCode string

const (
	ErrDivisionByZero   TaskErrorCode = "division_by_zero"
	ErrUnknownOperation TaskErrorCode = "unknown_operation"
	ErrInternalError    TaskErrorCode = "internal_error"
)

type TaskError struct {
	Code    TaskErrorCode
	Message string
}

func (e *TaskError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewTaskError(code TaskErrorCode, msg string) *TaskError {
	return &TaskError{
		Code:    code,
		Message: msg,
	}
}
