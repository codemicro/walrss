package urls

import (
	"fmt"
	"strings"
)

const (
	Index = "/"

	Auth             = "/auth"
	AuthSignIn       = Auth + "/signin"
	AuthRegister     = Auth + "/register"
	AuthOIDC         = Auth + "/oidc"
	AuthOIDCOutbound = AuthOIDC + "/outbound"
	AuthOIDCCallback = AuthOIDC + "/callback"

	Edit               = "/edit"
	EditEnabledState   = Edit + "/enabled"
	EditTimings        = Edit + "/timings"
	EditFeedItem       = Edit + "/feed/:id"
	CancelEditFeedItem = Edit + "/feed/:id/cancel"

	Export       = "/export"
	ExportAsOPML = Export + "/opml"

	Import         = "/import"
	ImportFromOPML = Import + "/opml"

	New         = "/new"
	NewFeedItem = New + "/feed"
	NewCategory = New + "/category"

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
