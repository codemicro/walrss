package core

import (
	"errors"
	"github.com/codemicro/walrss/walrss/internal/db"
	"github.com/codemicro/walrss/walrss/internal/state"
	"github.com/lithammer/shortuuid/v4"
	bh "github.com/timshannon/bolthold"
)

func NewFeed(st *state.State, userID, name, url string) (*db.Feed, error) {
	if err := validateFeedName(name); err != nil {
		return nil, err
	}

	if err := validateURL(url); err != nil {
		return nil, err
	}

	feed := &db.Feed{
		ID:     shortuuid.New(),
		URL:    url,
		Name:   name,
		UserID: userID,
	}

	if err := st.Data.Insert(feed.ID, feed); err != nil {
		return nil, err
	}

	return feed, nil
}

func GetFeedsForUser(st *state.State, userID string) ([]*db.Feed, error) {
	var feeds []*db.Feed
	if err := st.Data.Find(&feeds, bh.Where("UserID").Eq(userID)); err != nil {
		return nil, err
	}
	return feeds, nil
}

func GetFeed(st *state.State, id string) (*db.Feed, error) {
	feed := new(db.Feed)
	if err := st.Data.FindOne(feed, bh.Where("ID").Eq(id)); err != nil {
		if errors.Is(err, bh.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return feed, nil
}

func DeleteFeed(st *state.State, id string) error {
	if err := st.Data.Delete(id, new(db.Feed)); err != nil {
		return err
	}
	return nil
}

func UpdateFeed(st *state.State, feed *db.Feed) error {
	if err := validateFeedName(feed.Name); err != nil {
		return err
	}

	if err := validateURL(feed.URL); err != nil {
		return err
	}

	if err := st.Data.Update(feed.ID, feed); err != nil {
		if errors.Is(err, bh.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
