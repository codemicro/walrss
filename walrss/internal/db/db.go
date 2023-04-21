package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"strings"
)

func New(filename string) (*bun.DB, error) {
	dsn := filename
	log.Info().Msg("connecting to database")
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1) // https://github.com/mattn/go-sqlite3/issues/274#issuecomment-191597862

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

	ID         string `bun:"id,pk"`
	URL        string `bun:"url,notnull"`
	Name       string `bun:"name,notnull"`
	UserID     string `bun:"user_id,notnull"`
	CategoryID string `bun:"category_id,nullzero"`

	LastEtag      string `bun:"last_etag,nullzero"`
	LastModified  string `bun:"last_modified,nullzero"`
	CachedContent string `bun:"cached_content,nullzero"`

	User     *User     `bun:",rel:belongs-to,join:user_id=id"`
	Category *Category `bun:",rel:belongs-to,join:category_id=id"`
}

func (f *Feed) CacheWithEtag(etag, content string) {
	f.LastModified = ""
	f.LastEtag = etag
	f.CachedContent = content
}

func (f *Feed) CacheWithLastModified(lastModified, content string) {
	f.LastEtag = ""
	f.LastModified = lastModified
	f.CachedContent = content
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

type Category struct {
	bun.BaseModel `bun:"table:categories"`

	ID     string `bun:"id,pk"`
	Name   string `bun:"name,notnull"`
	UserID string `bun:"user_id,notnull"`

	User *User `bun:",rel:belongs-to,join:user_id=id"`
}
