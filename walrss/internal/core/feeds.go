package core

import (
	"errors"
	"github.com/codemicro/walrss/walrss/internal/core/opml"
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

func ExportFeedsForUser(st *state.State, userID string) ([]byte, error) {
	var feeds []*db.Feed
	user, err := GetUserByID(st, userID)
	if err != nil {
		return nil, err
	}
	if err := st.Data.Find(&feeds, bh.Where("UserID").Eq(userID)); err != nil {
		return nil, err
	}
	return opml.FromFeeds(feeds, user.Email).ToBytes()
}

func ImportFeedsForUser(st *state.State, userID string, opmlXML []byte) error {
	o, err := opml.FromBytes(opmlXML)
	if err != nil {
		return AsUserError(400, err)
	}

	// This will be used to filter out feeds included in OPML that would cause
	// duplicates
	existingURLs := make(map[string]struct{})
	{
		feeds, err := GetFeedsForUser(st, userID)
		if err != nil {
			return err
		}
		for _, feed := range feeds {
			if _, found := existingURLs[feed.URL]; !found {
				existingURLs[feed.URL] = struct{}{}
			}
		}
	}

	for _, feed := range o.ToFeeds() {
		if _, found := existingURLs[feed.URL]; found {
			continue
		}
		if _, err := NewFeed(st, userID, feed.Name, feed.URL); err != nil {
			return err
		}
	}

	return nil
}
