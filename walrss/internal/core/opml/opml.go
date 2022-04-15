package opml

import (
	"encoding/xml"
	"github.com/codemicro/walrss/walrss/internal/db"
	"github.com/lithammer/shortuuid/v4"
	"strings"
	"time"
)

type OPML struct {
	XMLName xml.Name `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    struct {
		Title       string    `xml:"title"`
		DateCreated time.Time `xml:"dateCreated,omitempty"`
		OwnerEmail  string    `xml:"ownerEmail,omitempty"`
	} `xml:"head"`
	Body struct {
		Outlines []*Outline `xml:"outline"`
	} `xml:"body"`
}

func FromBytes(x []byte) (*OPML, error) {
	o := new(OPML)
	if err := xml.Unmarshal(x, o); err != nil {
		return nil, err
	}
	return o, nil
}

func FromFeeds(feeds []*db.Feed, userEmailAddress string) *OPML {
	o := new(OPML)
	o.Version = "2.0"
	o.Head.Title = "Walrss feed export"
	o.Head.OwnerEmail = userEmailAddress
	o.Head.DateCreated = time.Now().UTC()

	for _, feed := range feeds {
		o.Body.Outlines = append(o.Body.Outlines, &Outline{
			Text:   feed.Name,
			Title:  feed.Name,
			Type:   "rss",
			XMLURL: feed.URL,
		})
	}

	return o
}

func (o *OPML) ToBytes() ([]byte, error) {
	return xml.Marshal(o)
}

func (o *OPML) ToFeeds() []*db.Feed {
	var out []*db.Feed
	for _, item := range o.Body.Outlines {
		out = append(out, item.ToFeeds()...)
	}
	return out
}

type Outline struct {
	Outlines []*Outline `xml:"outline"`
	Text     string     `xml:"text,attr"`
	Title    string     `xml:"title,attr,omitempty"`
	Type     string     `xml:"type,attr,omitempty"`
	XMLURL   string     `xml:"xmlUrl,attr,omitempty"`
}

func (o *Outline) ToFeeds() []*db.Feed {
	var out []*db.Feed

	if strings.EqualFold(o.Type, "rss") {
		name := o.Text
		if o.Title != "" {
			name = o.Title
		}

		out = append(out, &db.Feed{
			ID:   shortuuid.New(),
			URL:  o.XMLURL,
			Name: name,
		})
	}

	for _, item := range o.Outlines {
		out = append(out, item.ToFeeds()...)
	}

	return out
}
