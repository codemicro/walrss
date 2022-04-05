package db

import (
	"encoding/json"
	bh "github.com/timshannon/bolthold"
	"strings"
)

func New(filename string) (*bh.Store, error) {
	store, err := bh.Open(filename, 0644, &bh.Options{
		Encoder: json.Marshal,
		Decoder: json.Unmarshal,
	})
	if err != nil {
		return nil, err
	}
	return store, nil
}

type User struct {
	ID       string `boldholdKey:""`
	Email    string `boltholdUnique:"UniqueEmail"`
	Password []byte
	Salt     []byte

	Schedule struct {
		Active bool    `boltholdIndex:"Active"`
		Day    SendDay `boltholdIndex:"Day"`
		Hour   int     `boltholdIndex:"Hour"`
	}
}

type Feed struct {
	ID     string `boltholdKey:""`
	URL    string
	Name   string
	UserID string `boldholdIndex:"UserID"`
}

type FeedSlice []*Feed

func (f FeedSlice) Len() int {
	return len(f)
}

func (f FeedSlice) Less(i, j int) bool {
	return strings.ToLower(f[i].Name) < strings.ToLower(f[j].Name)
}

func (f FeedSlice) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
