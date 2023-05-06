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
	"net/http"
	"net/smtp"
	"net/textproto"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	dateFormat = "02Jan06"
	timeFormat = "15:04:05"
)

var ua = struct {
	ua   string
	once *sync.Once
}{"", new(sync.Once)}

func getUserAgent(st *state.State) string {
	ua.once.Do(func() {
		o := "walrss"
		if st.Config.Debug {
			o += "/DEV"
		} else if core.Version != "" {
			o += "/" + core.Version
		}
		o += " (" + st.Config.Platform.ContactInformation + ", https://github.com/codemicro/walrss)"
		ua.ua = o
	})
	return ua.ua
}

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
		if err := ProcessUserFeed(st, ur, nil); err != nil {
			log.Warn().Err(err).Str("user", ur.ID).Msg("could not process feeds for user")
		}
	}

	return nil
}

func reportProgress(channel chan string, msg string) {
	if channel == nil {
		return
	}
	channel <- msg
}

func ProcessUserFeed(st *state.State, user *db.User, progressChannel chan string) error {
	defer func() {
		if progressChannel == nil {
			return
		}
		close(progressChannel) // This is important! There's a chance that if this is not done before ProcessUserFeed
		// exits, the caller completely hang on this thread.
	}()

	reportProgress(progressChannel, "Fetching feed list")
	userFeeds, err := core.GetFeeds(st, &core.GetFeedsArgs{UserID: user.ID})
	if err != nil {
		return err
	}

	var interval time.Duration
	if user.Settings.ScheduleDay == db.SendDaily || user.Settings.ScheduleDay == db.SendDayNever {
		interval = time.Hour * 24
	} else {
		interval = time.Hour * 24 * 7
	}

	var processedFeeds []*processedFeed

	startTime := time.Now().UTC()

	reportProgress(progressChannel, "Fetching feed content")

	for i, f := range userFeeds {
		reportProgress(progressChannel, fmt.Sprintf("Fetching feed %d of %d: %s", i+1, len(userFeeds), f.Name))
		pf := new(processedFeed)
		pf.Name = f.Name

		rawFeed, err := getFeedContent(st, f)
		if err != nil {
			pf.Error = err
			reportProgress(progressChannel, "Failed to fetch: "+err.Error())
		} else {
			// Instead of just using the interval, we want to include everything from the earliest day specified by
			// the interval.
			//
			// Say we were running at 5AM and we had an interval of 24 hours. We'd select all the feed items from up to
			// 5AM from the day before. Sometimes, this would end up with feeds published at exactly midnight being
			// ignored, for example.
			//
			// This doesn't explain it well, but I don't quite understand it, so this is what you're getting.

			t := time.Now().UTC().Add(-interval)

			pf.Items = filterFeedContent(
				rawFeed,
				time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC),
			)
		}
		processedFeeds = append(processedFeeds, pf)
	}

	reportProgress(progressChannel, "Finished fetching feed content\nGenerating email")

	plainContent, htmlContent, err := generateEmail(st, processedFeeds, interval, time.Now().UTC().Sub(startTime))
	if err != nil {
		return err
	}

	reportProgress(progressChannel, "Sending email")

	err = sendEmail(
		st,
		plainContent,
		htmlContent,
		user.Email,
		"RSS digest for "+time.Now().UTC().Format(dateFormat),
	)

	reportProgress(progressChannel, "Done!")

	return err
}

var (
	feedCache     = cache.New(time.Minute*10, time.Minute*20)
	feedFetchLock = new(sync.Mutex)
)

func getFeedContent(st *state.State, f *db.Feed) (*gofeed.Feed, error) {
	feedFetchLock.Lock()
	defer feedFetchLock.Unlock()

	if v, found := feedCache.Get(f.URL); found {
		return v.(*gofeed.Feed), nil
	}

	buf := new(bytes.Buffer)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var notModified bool
	headers := make(textproto.MIMEHeader)

	requestBuilder := requests.URL(f.URL).ToBytesBuffer(buf).UserAgent(getUserAgent(st)).CopyHeaders(headers)

	if f.LastEtag != "" || f.LastModified != "" {
		requestBuilder.AddValidator(
			func(resp *http.Response) error {
				if resp.StatusCode == http.StatusNotModified {
					notModified = true
					return nil
				} else {
					return requests.DefaultValidator(resp)
				}
			},
		)

		if f.LastEtag != "" {
			requestBuilder.Header("If-None-Match", f.LastEtag)
		} else if f.LastModified != "" {
			requestBuilder.Header("If-Modified-Since", f.LastModified)
		}

	} else {
		requestBuilder.AddValidator(requests.DefaultValidator) // Since we're using CopyHeaders, we need to add the
		// default validator back ourselves.
	}

	if err := requestBuilder.Fetch(ctx); err != nil {
		return nil, err
	}

	if notModified {
		log.Debug().Msgf("%s not modified", f.URL)
		buf.WriteString(f.CachedContent)
	} else if etag := headers.Get("ETag"); etag != "" {
		log.Debug().Msgf("%s modified (ETag)", f.URL)
		f.CacheWithEtag(etag, buf.String())
		if err := core.UpdateFeed(st, f); err != nil {
			return nil, fmt.Errorf("failed to cache ETag-ed response: %v", err)
		}
	} else if lastModified := headers.Get("Last-Modified"); lastModified != "" {
		log.Debug().Msgf("%s modified (Last-Modified)", f.URL)
		f.CacheWithLastModified(lastModified, buf.String())
		if err := core.UpdateFeed(st, f); err != nil {
			return nil, fmt.Errorf("failed to cache Last-Modified enabled response: %v", err)
		}
	}

	feed, err := gofeed.NewParser().Parse(buf)
	if err != nil {
		return nil, err
	}

	_ = feedCache.Add(f.URL, feed, cache.DefaultExpiration)

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
		if item.PublishedParsed == nil {
			continue
		}

		*item.PublishedParsed = item.PublishedParsed.UTC()

		if item.PublishedParsed.After(earliestPublishTime) || item.PublishedParsed.Equal(earliestPublishTime) {
			o = append(o, &feedItem{
				Title:       strings.TrimSpace(item.Title),
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

	var versionSpecifier string
	if core.Version != "" {
		versionSpecifier = " v" + core.Version
	}

	renderer := hermes.Hermes{
		Product: hermes.Product{
			Name:      "Walrss",
			Link:      st.Config.Server.ExternalURL,
			Logo:      st.Config.Server.ExternalURL + urls.Statics + "/logo_light.png",
			Copyright: fmt.Sprintf("This email was generated in %.2f seconds by Walrss"+versionSpecifier+" - Walrss is open source software licensed under the GNU AGPL v3 - https://github.com/codemicro/walrss", timeToGenerate.Seconds()),
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
	if st.Config.Debug {
		log.Debug().Str("addr", to).Str("subject", subject).Msg("skipping email send due to debug mode")
		return nil
	}

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
