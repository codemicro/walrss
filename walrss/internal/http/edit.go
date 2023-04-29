package http

import (
	"github.com/codemicro/walrss/walrss/internal/core"
	"github.com/codemicro/walrss/walrss/internal/db"
	"github.com/codemicro/walrss/walrss/internal/http/views"
	"github.com/codemicro/walrss/walrss/internal/urls"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"strings"
)

func (s *Server) editEnabledState(ctx *fiber.Ctx) error {
	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestFragmentSignIn(ctx, urls.Index)
	}

	user, err := core.GetUserByID(s.state, currentUserID)
	if err != nil {
		return err
	}

	if strings.ToLower(ctx.FormValue("enable", "off")) == "on" {
		user.Active = true
	} else {
		user.Active = false
	}

	if err := core.UpdateUser(s.state, user); err != nil {
		return err
	}

	fragmentEmitSuccess(ctx)
	return ctx.SendString((&views.MainPage{
		EnableDigests: user.Active,
		SelectedDay:   user.ScheduleDay,
		SelectedTime:  user.ScheduleHour,
	}).RenderScheduleCard())
}

func (s *Server) editTimings(ctx *fiber.Ctx) error {
	currentUserID := getCurrentUserID(ctx)
	if currentUserID == "" {
		return requestFragmentSignIn(ctx, urls.Index)
	}

	user, err := core.GetUserByID(s.state, currentUserID)
	if err != nil {
		return err
	}

	if n, err := strconv.ParseInt(ctx.FormValue("day"), 10, 32); err != nil {
		return core.AsUserError(fiber.StatusBadRequest, err)
	} else {
		x := db.SendDay(n)
		if x > db.SendOnSunday || x < 0 {
			return core.NewUserError("invalid day: out of range 0<=x<=%d", int(db.SendOnSunday))
		}
		user.ScheduleDay = x
	}

	if n, err := strconv.ParseInt(ctx.FormValue("time"), 10, 8); err != nil {
		return core.AsUserError(fiber.StatusBadRequest, err)
	} else {
		x := int(n)
		if x > 23 || x < 0 {
			return core.NewUserError("invalid time: out of range 0<=x<=23")
		}
		user.ScheduleHour = x
	}

	if err := core.UpdateUser(s.state, user); err != nil {
		return err
	}

	fragmentEmitSuccess(ctx)
	return ctx.SendString((&views.MainPage{
		EnableDigests: user.Active,
		SelectedDay:   user.ScheduleDay,
		SelectedTime:  user.ScheduleHour,
	}).RenderScheduleCard())
}
