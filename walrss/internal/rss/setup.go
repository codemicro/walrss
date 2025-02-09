package rss

import (
	"fmt"
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/db"
	"github.com/codemicro/walrss/walrss/internal/state"
)

func LoadInitialItems(st *state.State, feed *db.Feed) error {
	content, err := getFeedContent(st, feed)
	if err != nil {
		return core.AsUserError(400, fmt.Errorf("get feed content: %w", err))
	}

	var fis []*db.FeedItem

	for _, item := range content.Items {
		fis = append(fis, &db.FeedItem{FeedID: feed.ID, ItemID: item.GUID})
	}

	return core.NewFeedItems(st, fis)
}
