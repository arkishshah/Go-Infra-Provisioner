package models

import "fmt"

type ProvisionError struct {
	Code    string
	Message string
	Err     error
}

func (e *ProvisionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewProvisionError(code, message string, err error) *ProvisionError {
	return &ProvisionError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
