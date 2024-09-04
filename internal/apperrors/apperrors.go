package apperrors

import (
	"errors"
	"fmt"
)

type AppError struct {
	Code        string
	Message     string
	Err         error
	UserMessage string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code, message string, userMessage ...string) *AppError {
	err := &AppError{
		Code:    code,
		Message: message,
	}
	if len(userMessage) > 0 {
		err.UserMessage = userMessage[0]
	}
	return err
}

func Wrap(err error, code, message string, userMessage ...string) *AppError {
	appErr := &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
	if len(userMessage) > 0 {
		appErr.UserMessage = userMessage[0]
	}
	return appErr
}

func HandleError(err error, context string) (string, string) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.UserMessage == "" {
			appErr.UserMessage = "An error occurred. Please try again later."
		}

		return appErr.UserMessage,
			fmt.Sprintf("%s: %s (Code: %s)", context, appErr.Message, appErr.Code)
	}
	return "An unexpected error occurred. Please try again later.",
		fmt.Sprintf("Unexpected error in %s: %v", context, err)
}
