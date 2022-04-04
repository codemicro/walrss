package db

import (
	"errors"
	"strings"
	"time"
)

type SendDay uint32

const (
	SendDayNever SendDay = iota
	SendDaily
	SendOnMonday
	SendOnTuesday
	SendOnWednesday
	SendOnThursday
	SendOnFriday
	SendOnSaturday
	SendOnSunday
	LastSendDay
)

func (s SendDay) String() string {
	var x string

	switch s {
	case SendDayNever:
		x = "never"
	case SendDaily:
		x = "daily"
	case SendOnMonday:
		x = "Monday"
	case SendOnTuesday:
		x = "Tuesday"
	case SendOnWednesday:
		x = "Wednesday"
	case SendOnThursday:
		x = "Thursday"
	case SendOnFriday:
		x = "Friday"
	case SendOnSaturday:
		x = "Saturday"
	case SendOnSunday:
		x = "Sunday"
	}

	return x
}

func (s SendDay) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *SendDay) UnmarshalText(x []byte) error {

	switch strings.ToLower(string(x)) {
	case "never":
		*s = SendDayNever
	case "daily":
		*s = SendDaily
	case "monday":
		*s = SendOnMonday
	case "tuesday":
		*s = SendOnTuesday
	case "wednesday":
		*s = SendOnWednesday
	case "thursday":
		*s = SendOnThursday
	case "friday":
		*s = SendOnFriday
	case "saturday":
		*s = SendOnSaturday
	case "sunday":
		*s = SendOnSunday
	default:
		return errors.New("unrecognised day")
	}

	return nil
}

func SendDayFromWeekday(w time.Weekday) SendDay {
	s := new(SendDay)
	if err := s.UnmarshalText([]byte(w.String())); err != nil {
		panic(err)
	}
	return *s
}
