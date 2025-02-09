package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"strings"
	"time"
)

func New(filename string) (*bun.DB, error) {
	dsn := filename
	log.Info().Msg("connecting to database")
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1) // https://github.com/mattn/go-sqlite3/issues/274#issuecomment-191597862

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("setting PRAGMA foreign_keys = ON: %w", err)
	}

	return bun.NewDB(db, sqlitedialect.New()), nil
}

type User struct {
	bun.BaseModel `bun:"table:users"`

	ID       string `bun:"id,pk"`
	Email    string `bun:"email,notnull,unique"`
	Password []byte `bun:"password"`
	Salt     []byte `bun:"salt"`

	Active       bool    `bun:"active,notnull"`
	ScheduleDay  SendDay `bun:"schedule_day"`
	ScheduleHour int     `bun:"schedule_hour"`
}

type Feed struct {
	bun.BaseModel `bun:"table:feeds"`

	ID     string `bun:"id,pk"`
	URL    string `bun:"url,notnull"`
	Name   string `bun:"name,notnull"`
	UserID string `bun:"user_id,notnull"`

	LastFetched   time.Time `bun:"last_fetched,nullzero"`
	LastEtag      string    `bun:"last_etag,nullzero"`
	LastModified  string    `bun:"last_modified,nullzero"`
	CachedContent string    `bun:"cached_content,nullzero"`

	User *User `bun:",rel:belongs-to,join:user_id=id"`
}

func (f *Feed) SetCacheWithEtag(etag, content string) {
	f.LastModified = ""
	f.LastEtag = etag
	f.CachedContent = content
}

func (f *Feed) SetCacheWithLastModified(lastModified, content string) {
	f.LastEtag = ""
	f.LastModified = lastModified
	f.CachedContent = content
}

func (f *Feed) ClearCache() {
	f.LastEtag = ""
	f.LastModified = ""
	f.CachedContent = ""
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

type FeedItem struct {
	bun.BaseModel `bun:"table:feed_items"`

	FeedID string `bun:"feed_id,notnull"`
	ItemID string `bun:"item_id,notnull"`

	// Feed *Feed `bun:",rel:belongs-to,join:feed_id=id"` // don't think this is needed but here in case??
}
