package apperrors

import (
	"errors"
	"fmt"
)

type AppError struct {
	Code        string
	Message     string
	PublicError string
	Err         error
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code, message string, publicError ...string) *AppError {
	var pubErr string
	if len(publicError) > 0 {
		pubErr = publicError[0]
	}
	return &AppError{Code: code, Message: message, PublicError: pubErr}
}

func Wrap(err error, code, message string, publicError ...string) *AppError {
	var pubErr string
	if len(publicError) > 0 {
		pubErr = publicError[0]
	}
	return &AppError{Code: code, Message: message, PublicError: pubErr, Err: err}
}

func HandleError(err error, context string) (string, string) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		publicMessage := appErr.PublicError
		if publicMessage == "" {
			publicMessage = "An error occurred. Please try again later."
		}
		return publicMessage,
			fmt.Sprintf("%s: %s (Code: %s)", context, appErr.Message, appErr.Code)
	}
	return "An unexpected error occurred. Please try again later.",
		fmt.Sprintf("Unexpected error in %s: %v", context, err)
}
