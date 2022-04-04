package rss

import (
	"bytes"
	"context"
	"fmt"
	"github.com/carlmjohnson/requests"
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/db"
	"github.com/codemicro/walrss/walrss/internal/state"
	"github.com/matcornic/hermes"
	"github.com/mmcdole/gofeed"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	dateFormat = "02Jan06"
	timeFormat = "15:04:05"
)

type processedFeed struct {
	Name  string
	Items []*feedItem
	Error error
}

func ProcessFeeds(st *state.State, day db.SendDay, hour int) error {
	u, e := core.GetUsersBySchedule(st, day, hour)
	for _, ur := range u {
		fmt.Printf("%#v\n", ur)

		userFeeds, err := core.GetFeedsForUser(st, ur.ID)
		if err != nil {
			return err
		}

		var processedFeeds []*processedFeed

		for _, f := range userFeeds {
			pf := new(processedFeed)
			pf.Name = f.Name

			rawFeed, err := getFeedContent(f.URL)
			if err != nil {
				pf.Error = err
			} else {
				pf.Items = filterFeedContent(rawFeed, time.Date(2022, 04, 01, 0, 0, 0, 0, time.UTC))
			}
			processedFeeds = append(processedFeeds, pf)
		}

		plainContent, htmlContent, err := generateEmail(processedFeeds)
		if err != nil {
			return err
		}

		// TODO: Send email
	}

	return nil
}

var (
	feedCache     = cache.New(time.Minute*10, time.Minute*20)
	feedFetchLock = new(sync.Mutex)
)

func getFeedContent(url string) (*gofeed.Feed, error) {
	feedFetchLock.Lock()
	defer feedFetchLock.Unlock()

	if v, found := feedCache.Get(url); found {
		return v.(*gofeed.Feed), nil
	}

	buf := new(bytes.Buffer)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := requests.URL(url).ToBytesBuffer(buf).Fetch(ctx); err != nil {
		return nil, err
	}

	feed, err := gofeed.NewParser().Parse(buf)
	if err != nil {
		return nil, err
	}

	_ = feedCache.Add(url, feed, cache.DefaultExpiration)

	return feed, nil
}

type feedItem struct {
	Title       string
	URL         string
	PublishTime time.Time
}

func filterFeedContent(feed *gofeed.Feed, earliestPublishTime time.Time) []*feedItem {
	var o []*feedItem

	for _, item := range feed.Items {
		if item.PublishedParsed != nil && item.PublishedParsed.After(earliestPublishTime) {
			o = append(o, &feedItem{
				Title:       item.Title,
				URL:         item.Link,
				PublishTime: *item.PublishedParsed,
			})
		}
	}

	return o
}

func generateEmail(processedItems []*processedFeed) (plain, html []byte, err error) {
	sort.Slice(processedItems, func(i, j int) bool {
		pi, pj := processedItems[i], processedItems[j]

		if pi.Error != nil && pj.Error == nil {
			return false
		}

		if pi.Error == nil && pj.Error != nil {
			return true
		}

		return pi.Name < pj.Name
	})

	var sb strings.Builder

	for _, processedItem := range processedItems {

		if len(processedItem.Items) != 0 || processedItem.Error != nil {
			sb.WriteString("* **")
			sb.WriteString(strings.ReplaceAll(processedItem.Name, "*", `\*`))
			sb.WriteString("**\n")
		}

		if processedItem.Error != nil {
			sb.WriteString("  * **Error:** ")
			sb.WriteString(processedItem.Error.Error())
			sb.WriteString("\n")
		} else {
			r := strings.NewReplacer("[", `\[`, "]", `\]`, "*", `\*`)

			for _, item := range processedItem.Items {
				sb.WriteString("  * [**")
				sb.WriteString(r.Replace(item.Title))
				sb.WriteString("**](")
				sb.WriteString(item.URL)
				sb.WriteString(") - ")
				sb.WriteString(strings.ToUpper(item.PublishTime.Format(dateFormat + " " + timeFormat)))
				sb.WriteString("\n")
			}

		}
	}

	e := hermes.Email{
		Body: hermes.Body{
			FreeMarkdown: hermes.Markdown(sb.String()),
		},
	}

	renderer := hermes.Hermes{
		Theme: new(hermes.Flat),
	}

	plainString, err := renderer.GeneratePlainText(e)
	if err != nil {
		return nil, nil, err
	}

	htmlString, err := renderer.GenerateHTML(e)
	if err != nil {
		return nil, nil, err
	}

	return []byte(plainString), []byte(htmlString), nil
}
