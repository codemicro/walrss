package core

import "fmt"

var ErrNotFound = NewUserErrorWithStatus(404, "item not found")

type UserError struct {
	Original error
	Status   int
}

func (ue *UserError) Error() string {
	return ue.Original.Error()
}

func (ue *UserError) Unwrap() error {
	return ue.Original
}

func AsUserError(status int, err error) error {
	return &UserError{
		Original: err,
		Status:   status,
	}
}

func NewUserError(format string, args ...any) error {
	return NewUserErrorWithStatus(400, format, args...)
}

func NewUserErrorWithStatus(status int, format string, args ...any) error {
	return &UserError{
		Original: fmt.Errorf(format, args...),
		Status:   status,
	}
}

func IsUserError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*UserError)
	return ok
}

func GetUserErrorStatus(err error) int {
	ue, ok := err.(*UserError)
	if !ok {
		return 0
	}
	return ue.Status
}
