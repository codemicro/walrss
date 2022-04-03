package http

import (
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/http/views"
	"github.com/codemicro/walrss/walrss/internal/urls"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) newFeedItem(ctx *fiber.Ctx) error {

	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestFragmentSignIn(ctx, urls.Index)
	}

	switch ctx.Method() {
	case fiber.MethodGet:
		return ctx.SendString(views.RenderNewFeedItemRow())
	case fiber.MethodPost:
		feed, err := core.NewFeed(
			s.state,
			currentUserID,
			ctx.FormValue("name"),
			ctx.FormValue("url"),
		)

		if err != nil {
			return err
		}

		return ctx.SendString(views.RenderFeedRow(feed.ID, feed.Name, feed.URL))
	}

	panic("unreachable")
}
