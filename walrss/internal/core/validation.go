package core

import (
	"regexp"
)

var emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func validateEmailAddress(email string) error {
	if !emailRegexp.MatchString(email) {
		return NewUserError("invalid email address")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) <= 3 {
		return NewUserError("password must be at least three characters long")
	}
	return nil
}
