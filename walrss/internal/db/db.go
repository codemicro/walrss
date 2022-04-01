package db

import (
	bh "github.com/timshannon/bolthold"
)

func New(filename string) (*bh.Store, error) {
	store, err := bh.Open(filename, 0644, nil)
	if err != nil {
		return nil, err
	}
	return store, nil
}

type User struct {
	ID       string `boldholdKey:""`
	Email    string `boltholdUnique:"UniqueEmail" boltholdIndex:"Email"`
	Password []byte
	Salt     []byte

	Schedule struct {
		Day  SendDay `boltholdIndex:"Day"`
		Hour int     `boltholdIndex:"Hour"`
	}
}

type SendDay uint32

const (
	SendDayNever = iota
	SendDaily
	SendOnMonday
	SendOnTuesday
	SendOnWednesday
	SendOnThursday
	SendOnFriday
	SendOnSaturday
	SendOnSunday
)
