package opml

import (
	"encoding/xml"
	"github.com/codemicro/walrss/walrss/internal/db"
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

// Outline holds all information about an outline.
type Outline struct {
	Outlines []*Outline `xml:"outline"`
	Text     string     `xml:"text,attr"`
	Title    string     `xml:"title,attr,omitempty"`
	Type     string     `xml:"type,attr,omitempty"`
	XMLURL   string     `xml:"xmlUrl,attr,omitempty"`
}

func FromBytes(x []byte) (*OPML, error) {
	o := new(OPML)
	if err := xml.Unmarshal(x, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (o *OPML) ToBytes() ([]byte, error) {
	return xml.Marshal(o)
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
