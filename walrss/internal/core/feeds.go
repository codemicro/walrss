package core

import (
	"context"
	"database/sql"
	"errors"
	"github.com/codemicro/walrss/walrss/internal/core/opml"
	"github.com/codemicro/walrss/walrss/internal/db"
	"github.com/codemicro/walrss/walrss/internal/state"
	"github.com/lithammer/shortuuid/v4"
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

	if _, err := st.Data.NewInsert().Model(feed).Exec(context.Background()); err != nil {
		return nil, err
	}

	return feed, nil
}

func NewFeeds(st *state.State, userID string, fs []*db.Feed) error {
	if len(fs) == 0 {
		return nil
	}

	for i, f := range fs {
		f.ID = shortuuid.New()
		f.UserID = userID
		if err := validateFeedName(f.Name); err != nil {
			return NewUserErrorWithStatus(400, "validate feed %d: %w", i, err)
		}
		if err := validateURL(f.URL); err != nil {
			return NewUserErrorWithStatus(400, "validate feed %d: %w", i, err)
		}
	}

	_, err := st.Data.NewInsert().Model(&fs).Exec(context.Background())
	return err
}

func GetFeedsForUser(st *state.State, userID string) (res []*db.Feed, err error) {
	err = st.Data.NewSelect().
		Model(&res).
		Relation("User").
		Where("Feed.user_id = ?", userID).
		Scan(context.Background())
	return
}

func GetFeed(st *state.State, id string) (res *db.Feed, err error) {
	res = new(db.Feed)
	err = st.Data.NewSelect().Model(res).Where("id = ?", id).Scan(context.Background())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return
}

func DeleteFeed(st *state.State, id string) error {
	_, err := st.Data.NewDelete().Model((*db.Feed)(nil)).Where("id = ?", id).Exec(context.Background())
	return err
}

func UpdateFeed(st *state.State, feed *db.Feed) error {
	if err := validateFeedName(feed.Name); err != nil {
		return err
	}

	if err := validateURL(feed.URL); err != nil {
		return err
	}

	_, err := st.Data.NewUpdate().Model(feed).WherePK().Exec(context.Background())
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func ExportFeedsForUser(st *state.State, userID string) ([]byte, error) {
	var feeds []*db.Feed
	user, err := GetUserByID(st, userID)
	if err != nil {
		return nil, err
	}

	if err := st.Data.NewSelect().
		Model(&feeds).
		Where("user_id = ?", userID).
		Scan(context.Background()); err != nil {
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

	var (
		fs = o.ToFeeds()
		n  int
	)
	for _, feed := range fs {
		if _, found := existingURLs[feed.URL]; !found {
			fs[n] = feed
			n += 1
		}
	}
	fs = fs[:n]

	return NewFeeds(st, userID, fs)
}
