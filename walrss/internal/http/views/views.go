package views

import "github.com/gofiber/fiber/v2"

//go:generate go install github.com/valyala/quicktemplate/qtc@latest
//go:generate qtc -skipLineComments -ext qtpl.html

func SendPage(ctx *fiber.Ctx, page Page) error {
	ctx.Set(fiber.HeaderContentType, "html")
	return ctx.SendString(RenderPage(page))
}

func makePageTitle(p Page) string {
	t := p.Title()
	if t == "" {
		return "Walrss"
	}
	return t + " | Walrss"
}
