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

		var parts []string
		if st.Config.Platform.ContactInformation != "" {
			parts = append(parts, st.Config.Platform.ContactInformation)
		}
		parts = append(parts, "https://github.com/codemicro/walrss")

		o += " (" + strings.Join(parts, ", ") + ")"
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
	userFeeds, err := core.GetFeedsForUser(st, user.ID)
	if err != nil {
		return err
	}

	var interval time.Duration
	if user.ScheduleDay == db.SendDaily || user.ScheduleDay == db.SendDayNever {
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
			pf.Items, err = filterFeedContent(
				st,
				rawFeed,
				f.ID,
			)
			if err != nil {
				return fmt.Errorf("filter for new feed items in %s: %w", f.ID, err)
			}

			// add new items to DB cache
			{
				var newItems []*db.FeedItem
				for _, i := range pf.Items {
					newItems = append(newItems, &db.FeedItem{
						FeedID: f.ID,
						ItemID: i.ID,
					})
				}
				if err := core.NewFeedItems(st, newItems); err != nil {
					return fmt.Errorf("insert new feed items for feed %s: %w", f.ID, err)
				}
			}
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

var feedFetchLock = new(sync.Mutex)

func getFeedContent(st *state.State, f *db.Feed) (*gofeed.Feed, error) {
	feedFetchLock.Lock() // I would like to be able to get rid of this lock, however, in order to do so, a lot of the
	// database infrastructure needs removing and rewriting to use proper transactions. So we'll leave it here for now.
	defer feedFetchLock.Unlock()

	buf := new(bytes.Buffer)

	// If a feed was cached in the last hour, Walrss will not re-query the remote server and will just use the cache.
	hasCachedFeed := f.CachedContent != ""
	cachedFeedIsFresh := !f.LastFetched.IsZero() && time.Now().UTC().Sub(f.LastFetched) < time.Hour

	if hasCachedFeed && cachedFeedIsFresh {
		log.Debug().Msgf("%s using fresh cache (%v)", f.URL, f.LastFetched)
		buf.WriteString(f.CachedContent)
	} else {
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

		f.LastFetched = time.Now().UTC()

		if notModified {
			log.Debug().Msgf("%s not modified", f.URL)
			buf.WriteString(f.CachedContent)
		} else if etag := headers.Get("ETag"); etag != "" {
			log.Debug().Msgf("%s modified (ETag)", f.URL)
			f.SetCacheWithEtag(etag, buf.String())
		} else if lastModified := headers.Get("Last-Modified"); lastModified != "" {
			log.Debug().Msgf("%s modified (Last-Modified)", f.URL)
			f.SetCacheWithLastModified(lastModified, buf.String())
		}

		if err := core.UpdateFeed(st, f); err != nil {
			return nil, fmt.Errorf("update feed after fetch: %v", err)
		}
	}

	feed, err := gofeed.NewParser().Parse(buf)
	if err != nil {
		return nil, err
	}

	return feed, nil
}

type feedItem struct {
	ID          string
	Title       string
	URL         string
	PublishTime time.Time
}

func filterFeedContent(st *state.State, feed *gofeed.Feed, feedID string) ([]*feedItem, error) {
	knownItemsList, err := core.GetFeedItemsForFeed(st, feedID)
	if err != nil {
		return nil, fmt.Errorf("get known feed items: %w", err)
	}

	knownItems := make(map[string]struct{})
	for _, i := range knownItemsList {
		knownItems[i.ItemID] = struct{}{}
	}

	var o []*feedItem

	for _, item := range feed.Items {
		if _, found := knownItems[item.GUID]; !found {
			o = append(o, &feedItem{
				ID:          item.GUID,
				Title:       strings.TrimSpace(item.Title),
				URL:         item.Link,
				PublishTime: *item.PublishedParsed,
			})
		}
	}

	return o, nil
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

	e := &email.Email{
		From:    st.Config.Email.From,
		To:      []string{to},
		Subject: subject,
		Text:    plain,
		HTML:    html,
	}

	var smtpAuth smtp.Auth
	if st.Config.Email.Username != "" || st.Config.Email.Password != "" {
		smtpAuth = smtp.PlainAuth("", st.Config.Email.Username, st.Config.Email.Password, st.Config.Email.Host)
	}

	smtpAddr := fmt.Sprintf("%s:%d", st.Config.Email.Host, st.Config.Email.Port)

	var sendFn func(*email.Email) error

	switch st.Config.Email.TLS {
	case "no":
		sendFn = func(e *email.Email) error {
			return e.Send(smtpAddr, smtpAuth)
		}
	case "tls":
		sendFn = func(e *email.Email) error {
			return e.SendWithTLS(smtpAddr, smtpAuth, &tls.Config{
				ServerName: st.Config.Email.Host,
			})
		}
	case "starttls":
		sendFn = func(e *email.Email) error {
			return e.SendWithStartTLS(smtpAddr, smtpAuth, &tls.Config{
				ServerName: st.Config.Email.Host,
			})
		}
	default:
		return fmt.Errorf("unknown TLS option %s", st.Config.Email.TLS)
	}

	return sendFn(e)
}
