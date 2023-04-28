package http

import (
	"fmt"
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
		cats, err := core.GetCategoriesForUser(s.state, currentUserID)
		if err != nil {
			return fmt.Errorf("GET newFeedItem: %w", err)
		}
		return ctx.SendString(neoviews.FragmentNewFeed(&neoviews.FragmentNewFeedArgs{
			CurrentCategoryID: ctx.Query("category"),
			Categories:        cats,
		}))
	case fiber.MethodPost:
		categoryID := ctx.FormValue("categoryID")

		_, err := core.NewFeed(
			s.state,
			currentUserID,
			ctx.FormValue("name"),
			ctx.FormValue("url"),
			categoryID,
		)
		if err != nil {
			return err
		}

		// TODO: Set category
		resp, err := neoviews.RenderFeedTabsAndTableForUser(s.state, currentUserID, categoryID, true)
		if err != nil {
			return err
		}
		fragmentEmitSuccess(ctx)
		return ctx.SendString(daz.H("div")() + resp)
	}

	panic("unreachable")
}

func (s *Server) newCategory(ctx *fiber.Ctx) error {
	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestFragmentSignIn(ctx, urls.Index)
	}

	switch ctx.Method() {
	case fiber.MethodGet:
		return ctx.SendString(neoviews.FragmentNewCategory())
	case fiber.MethodPost:
		newCat, err := core.NewCategory(
			s.state,
			currentUserID,
			ctx.FormValue("name"),
		)
		if err != nil {
			return err
		}

		resp, err := neoviews.RenderFeedTabsAndTableForUser(s.state, currentUserID, newCat.ID, true)
		if err != nil {
			return err
		}
		fragmentEmitSuccess(ctx)
		return ctx.SendString(daz.H("div")() + resp)
	}

	panic("unreachable")
}
