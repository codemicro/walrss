package core

import (
	"context"
	"database/sql"
	"errors"
	"github.com/codemicro/walrss/walrss/internal/core/opml"
	"github.com/codemicro/walrss/walrss/internal/db"
	"github.com/codemicro/walrss/walrss/internal/state"
	"github.com/lithammer/shortuuid/v4"
	"strings"
)

func NewFeed(st *state.State, userID, name, url, categoryID string) (*db.Feed, error) {
	if err := validateFeedName(name); err != nil {
		return nil, err
	}

	if err := validateURL(url); err != nil {
		return nil, err
	}

	if err := validateCategoryID(st, categoryID); err != nil {
		return nil, err
	}

	feed := &db.Feed{
		ID:         shortuuid.New(),
		URL:        url,
		Name:       name,
		UserID:     userID,
		CategoryID: categoryID,
	}

	if _, err := st.Data.NewInsert().Model(feed).Exec(context.Background()); err != nil {
		return nil, err
	}

	return feed, nil
}

type GetFeedsArgs struct {
	UserID     string
	CategoryID *string
}

func GetFeeds(st *state.State, args *GetFeedsArgs) (res []*db.Feed, err error) {
	q := st.Data.NewSelect().
		Model(&res)

	if args.UserID != "" {
		q = q.Relation("User")
		q = q.Where("Feed.user_id = ?", args.UserID)
	}

	if args.CategoryID != nil {
		q = q.Relation("Category")
		if *args.CategoryID == "" {
			q = q.Where("Feed.category_id IS NULL")
		} else {
			q = q.Where("Feed.category_id = ?", *args.CategoryID)
		}
	}

	err = q.Scan(context.Background())
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
	feed.Name = strings.TrimSpace(feed.Name)
	feed.URL = strings.TrimSpace(feed.URL)

	if err := validateFeedName(feed.Name); err != nil {
		return err
	}

	if err := validateURL(feed.URL); err != nil {
		return err
	}

	if err := validateCategoryID(st, feed.CategoryID); err != nil {
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
		feeds, err := GetFeeds(st, &GetFeedsArgs{UserID: userID})
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
		if _, err := NewFeed(st, userID, feed.Name, feed.URL, ""); err != nil {
			return err
		}
	}

	return nil
}
