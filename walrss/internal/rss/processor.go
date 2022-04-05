package rss

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/carlmjohnson/requests"
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/db"
	"github.com/codemicro/walrss/walrss/internal/state"
	"github.com/codemicro/walrss/walrss/internal/urls"
	"github.com/jordan-wright/email"
	"github.com/matcornic/hermes"
	"github.com/mmcdole/gofeed"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
	"net/smtp"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	dateFormat = "02JAN06"
	timeFormat = "15:04:05"
)

type processedFeed struct {
	Name  string
	Items []*feedItem
	Error error
}

func ProcessFeeds(st *state.State, day db.SendDay, hour int) error {
	u, err := core.GetUsersBySchedule(st, day, hour)
	if err != nil {
		return err
	}

	for _, ur := range u {
		if err := ProcessUserFeed(st, ur); err != nil {
			log.Warn().Err(err).Str("user", ur.ID).Msg("could not process feeds for user")
		}
	}

	return nil
}

func ProcessUserFeed(st *state.State, user *db.User) error {
	userFeeds, err := core.GetFeedsForUser(st, user.ID)
	if err != nil {
		return err
	}

	var interval time.Duration
	if user.Schedule.Day == db.SendDaily {
		interval = time.Hour * 24
	} else {
		interval = time.Hour * 24 * 7
	}

	var processedFeeds []*processedFeed

	startTime := time.Now().UTC()

	for _, f := range userFeeds {
		pf := new(processedFeed)
		pf.Name = f.Name

		rawFeed, err := getFeedContent(f.URL)
		if err != nil {
			pf.Error = err
		} else {
			pf.Items = filterFeedContent(rawFeed, time.Now().UTC().Add(-interval))
		}
		processedFeeds = append(processedFeeds, pf)
	}

	plainContent, htmlContent, err := generateEmail(st, processedFeeds, interval, time.Now().UTC().Sub(startTime))
	if err != nil {
		return err
	}

	return sendEmail(
		st,
		plainContent,
		htmlContent,
		user.Email,
		"RSS digest for "+time.Now().UTC().Format(dateFormat),
	)
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

func generateEmail(st *state.State, processedItems []*processedFeed, interval, timeToGenerate time.Duration) (plain, html []byte, err error) {
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

	sb.WriteString("Here are the updates to the feeds you're subscribed to that have been published in the last ")

	if interval.Hours() == 24 {
		sb.WriteString("24 hours")
	} else {
		sb.WriteString(fmt.Sprintf("%.0f days", interval.Hours()/24))
	}

	sb.WriteString(".\n\n")

	if len(processedItems) == 0 {
		sb.WriteString("*There's nothing to show here right now.*\n\n")
	}

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
				sb.WriteString(item.PublishTime.Format(dateFormat + " " + timeFormat))
				sb.WriteString("\n")
			}

		}
	}

	e := hermes.Email{
		Body: hermes.Body{
			Title:     "Hi there!",
			Signature: "Have a good one",
			Outros: []string{
				"You can edit the feeds that you're subscribed to and your delivery settings here: " + st.Config.Server.ExternalURL,
			},
			FreeMarkdown: hermes.Markdown(sb.String()),
		},
	}

	renderer := hermes.Hermes{
		Product: hermes.Product{
			Name:      "Walrss",
			Link:      st.Config.Server.ExternalURL,
			Logo:      st.Config.Server.ExternalURL + urls.Statics + "/logo_light.png",
			Copyright: fmt.Sprintf("This email was generated in %.2f seconds by Walrss - Walrss is open source software licensed under the GNU AGPL v3 - https://github.com/codemicro/walrss", timeToGenerate.Seconds()),
		},
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

func sendEmail(st *state.State, plain, html []byte, to, subject string) error {
	return (&email.Email{
		From:    st.Config.Email.From,
		To:      []string{to},
		Subject: subject,
		Text:    plain,
		HTML:    html,
	}).SendWithStartTLS(
		fmt.Sprintf("%s:%d", st.Config.Email.Host, st.Config.Email.Port),
		smtp.PlainAuth("", st.Config.Email.Username, st.Config.Email.Password, st.Config.Email.Host),
		&tls.Config{
			ServerName: st.Config.Email.Host,
		},
	)
}
