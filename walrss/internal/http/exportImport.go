package http

import (
	"errors"
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/urls"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"io"
)

func (s *Server) exportAsOPML(ctx *fiber.Ctx) error {
	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestStandardSignIn(ctx)
	}

	exported, err := core.ExportFeedsForUser(s.state, currentUserID)
	if err != nil {
		return err
	}

	ctx.Set(fiber.HeaderContentType, "application/xml")
	return ctx.Send(exported)
}

func (s *Server) importFromOPML(ctx *fiber.Ctx) error {
	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestFragmentSignIn(ctx, urls.Index)
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		if errors.Is(err, fasthttp.ErrMissingFile) {
			return core.NewUserError("missing file")
		}
		return err
	}

	fileHandle, err := file.Open()
	if err != nil {
		return err
	}

	fileContents, err := io.ReadAll(fileHandle)
	if err != nil {
		return err
	}

	if err := core.ImportFeedsForUser(s.state, currentUserID, fileContents); err != nil {
		return err
	}

	ctx.Set("HX-Refresh", "true")
	ctx.Status(fiber.StatusNoContent)
	return nil
}
