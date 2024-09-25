package errors

import (
	"context"
	"errors"

	middleware "github.com/mohammedrefaat/hamber/Middleware"
)

type CustomError struct {
	Message string
	Code    int
}

func (e *CustomError) NewError() string {
	return e.Message
}

func NewError(err CustomError) *CustomError {
	return &CustomError{
		Message: err.Message,
		Code:    err.Code,
	}
}

func (a *CustomError) ErrUnknownUserId(ctx context.Context) error {
	lng, ok := ctx.Value(middleware.LanguageKey).(string)
	if !ok {
		lng = "en" // fallback if no language found
	}

	if lng == "ar" { // Assuming "ar" is the code for Arabic
		return errors.New("المستخدم غير معروف")
	}
	return errors.New("unknown user id")
}
