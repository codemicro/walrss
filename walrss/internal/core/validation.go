package core

import (
	"net/url"
	"regexp"
	"strings"
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

func validateFeedName(name string) error {
	if strings.TrimSpace(name) == "" {
		return NewUserError("feed name cannot be blank")
	}
	return nil
}

func validateURL(inputURL string) error {
	u, err := url.ParseRequestURI(inputURL)
	if err != nil {
		return NewUserError("invalid URL")
	}
	if s := strings.ToLower(u.Scheme); !(s == "http" || s == "https") {
		return NewUserError("invalid URL request scheme - must be HTTP or HTTPS")
	}
	return nil
}
