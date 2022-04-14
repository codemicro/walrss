package urls

import (
	"fmt"
	"strings"
)

const (
	Index = "/"

	Auth         = "/auth"
	AuthSignIn   = Auth + "/signin"
	AuthRegister = Auth + "/register"

	Edit               = "/edit"
	EditEnabledState   = Edit + "/enabled"
	EditTimings        = Edit + "/timings"
	EditFeedItem       = Edit + "/feed/:id"
	CancelEditFeedItem = Edit + "/feed/:id/cancel"

	Export       = "/export"
	ExportAsOPML = Export + "/opml"

	New         = "/new"
	NewFeedItem = New + "/feed"

	SendTestEmail = "/send/test"

	Statics = "/statics"
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
