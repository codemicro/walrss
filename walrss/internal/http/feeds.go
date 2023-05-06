package http

import (
	"fmt"
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/http/neoviews"
	"github.com/codemicro/walrss/walrss/internal/urls"
	"github.com/gofiber/fiber/v2"
	"github.com/stevelacy/daz"
)

func (s *Server) feedsPage(ctx *fiber.Ctx) error {
	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestStandardSignIn(ctx)
	}

	user, err := core.GetUserByID(s.state, currentUserID)
	if err != nil {
		return err
	}

	var x string
	feeds, err := core.GetFeeds(s.state, &core.GetFeedsArgs{UserID: currentUserID, CategoryID: &x})
	if err != nil {
		return err
	}

	cats, err := core.GetCategoriesForUser(s.state, currentUserID)
	if err != nil {
		return err
	}

	ctx.Type("html")
	return ctx.SendString(neoviews.FeedsPage(&neoviews.FeedsPageArgs{
		DigestsEnabled: user.Settings.DigestsActive,
		Feeds:          feeds,
		Categories:     cats,
	}))
}

func (s *Server) getFeedsTab(ctx *fiber.Ctx) error {
	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestFragmentSignIn(ctx, urls.Index)
	}
	category := ctx.Query("category")

	resp, err := neoviews.RenderFeedTabsAndTableForUser(s.state, currentUserID, category, false)
	if err != nil {
		return err
	}
	return ctx.SendString(resp)
}

func (s *Server) editFeedItem(ctx *fiber.Ctx) error {
	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestFragmentSignIn(ctx, urls.Index)
	}

	feedID := ctx.Params("id")
	categoryID := ctx.Query("category")

	feed, err := core.GetFeed(s.state, feedID)
	if err != nil {
		return err
	}

	switch ctx.Method() {
	case fiber.MethodGet:
		categories, err := core.GetCategoriesForUser(s.state, currentUserID)
		if err != nil {
			return fmt.Errorf("GET editFeedItem: %w", err)
		}

		return ctx.SendString(neoviews.FragmentEditFeed(&neoviews.FragmentEditFeedArgs{
			Feed:              feed,
			CurrentCategoryID: categoryID,
			Categories:        categories,
		}))
	case fiber.MethodDelete:
		if err := core.DeleteFeed(s.state, feed.ID); err != nil {
			return err
		}
	case fiber.MethodPut:
		feed.Name = ctx.FormValue("name")
		feed.URL = ctx.FormValue("url")
		feed.CategoryID = ctx.FormValue("categoryID")

		categoryID = feed.CategoryID

		if err := core.UpdateFeed(s.state, feed); err != nil {
			return err
		}
	}

	resp, err := neoviews.RenderFeedTabsAndTableForUser(s.state, currentUserID, categoryID, true)
	if err != nil {
		return err
	}
	fragmentEmitSuccess(ctx)
	return ctx.SendString(daz.H("div")() + resp)
}

func (s *Server) editCategory(ctx *fiber.Ctx) error {
	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestFragmentSignIn(ctx, urls.Index)
	}

	categoryID := ctx.Params("id")
	displayCategoryID := categoryID

	category, err := core.GetCategory(s.state, categoryID)
	if err != nil {
		return err
	}

	switch ctx.Method() {
	case fiber.MethodGet:
		return ctx.SendString(neoviews.FragmentEditCategory(&neoviews.FragmentEditCategoryArgs{
			Category: category,
		}))
	case fiber.MethodDelete:
		if err := core.DeleteCategory(s.state, categoryID); err != nil {
			return err
		}
		displayCategoryID = ""
	case fiber.MethodPut:
		category.Name = ctx.FormValue("name")
		if err := core.UpdateCategory(s.state, category); err != nil {
			return err
		}
	}

	resp, err := neoviews.RenderFeedTabsAndTableForUser(s.state, currentUserID, displayCategoryID, true)
	if err != nil {
		return err
	}
	fragmentEmitSuccess(ctx)
	return ctx.SendString(daz.H("div")() + resp)
}

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
