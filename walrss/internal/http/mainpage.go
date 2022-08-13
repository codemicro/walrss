package http

import (
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/http/views"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) mainPage(ctx *fiber.Ctx) error {
	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestStandardSignIn(ctx)
	}

	user, err := core.GetUserByID(s.state, currentUserID)
	if err != nil {
		return err
	}

	feeds, err := core.GetFeedsForUser(s.state, currentUserID)
	if err != nil {
		return err
	}

	return views.SendPage(ctx, &views.MainPage{
		EnableDigests: user.Active,
		SelectedDay:   user.ScheduleDay,
		SelectedTime:  user.ScheduleHour,
		Feeds:         feeds,
	})
}
