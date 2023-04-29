package urls

import (
	"fmt"
	"strings"
)

const (
	Index = "/"

	Feeds = "/feeds"

	FeedsFragments   = Feeds + "/-"
	FeedsCategoryTab = FeedsFragments + "/list-tab"
	FeedsNewFeed     = FeedsFragments + "/feed"
	FeedsNewCategory = FeedsFragments + "/category"
	FeedsFeed        = FeedsFragments + "/feed/:id"
	FeedsCategory    = FeedsFragments + "/category/:id"

	Auth             = "/auth"
	AuthSignIn       = Auth + "/signin"
	AuthRegister     = Auth + "/register"
	AuthOIDC         = Auth + "/oidc"
	AuthOIDCOutbound = AuthOIDC + "/outbound"
	AuthOIDCCallback = AuthOIDC + "/callback"

	Settings = "/settings"

	Edit               = "/edit"
	EditEnabledState   = Edit + "/enabled"
	EditTimings        = Edit + "/timings"
	CancelEditFeedItem = Edit + "/feed/:id/cancel"

	Export       = "/export"
	ExportAsOPML = Export + "/opml"

	Import         = "/import"
	ImportFromOPML = Import + "/opml"

	SendTestEmail   = "/send/test"
	TestEmailStatus = SendTestEmail + "/status"

	CancelModal = "/cancelmodal"

	Statics = "/assets"
)

func Expand(template string, replacements ...interface{}) string {
	spt := strings.Split(template, "/")
	for i, part := range spt {
		if len(part) == 0 {
			continue
		}
		if part[0] == ':' {
			spt[i] = "%s"
		}
	}
	return fmt.Sprintf(strings.Join(spt, "/"), replacements...)
}
