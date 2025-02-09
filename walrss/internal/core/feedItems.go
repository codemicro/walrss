package core

import (
	"context"
	"github.com/codemicro/walrss/walrss/internal/db"
	"github.com/codemicro/walrss/walrss/internal/state"
)

func NewFeedItem(st *state.State, feedID, itemID string) (*db.FeedItem, error) {
	fi := &db.FeedItem{
		FeedID: feedID,
		ItemID: itemID,
	}
	return fi, NewFeedItems(st, []*db.FeedItem{fi})
}

func NewFeedItems(st *state.State, fis []*db.FeedItem) error {
	_, err := st.Data.NewInsert().Model(&fis).Exec(context.Background())
	return err
}

func GetFeedItemsForFeed(st *state.State, feedID string) (res []*db.FeedItem, err error) {
	err = st.Data.NewSelect().
		Model(&res).
		Where("feed_id = ?", feedID).
		Scan(context.Background())
	return
}
