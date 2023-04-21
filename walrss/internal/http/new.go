package http

import (
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/http/neoviews"
	"github.com/codemicro/walrss/walrss/internal/urls"
	"github.com/gofiber/fiber/v2"
	"github.com/stevelacy/daz"
)

func (s *Server) newFeedItem(ctx *fiber.Ctx) error {

	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestFragmentSignIn(ctx, urls.Index)
	}

	switch ctx.Method() {
	case fiber.MethodGet:
		return ctx.SendString(neoviews.FragmentNewFeed())
		//return ctx.SendString(views.RenderNewFeedItemRow())
	case fiber.MethodPost:
		_, err := core.NewFeed(
			s.state,
			currentUserID,
			ctx.FormValue("name"),
			ctx.FormValue("url"),
		)

		if err != nil {
			return err
		}

		feeds, err := core.GetFeedsForUser(s.state, currentUserID)
		if err != nil {
			return err
		}

		fragmentEmitSuccess(ctx)
		// TODO: Set activeCategory
		return ctx.SendString(daz.H("div")() + neoviews.RenderFeedTabsAndTable(feeds, "", true))
	}

	panic("unreachable")
}
