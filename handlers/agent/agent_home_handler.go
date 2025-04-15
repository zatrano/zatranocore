package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// AgentHomeHandler agent ana sayfasını işler
func AgentHomeHandler(c *fiber.Ctx) error {
	return c.Render("agent/home/agent_home", fiber.Map{
		"Title": "Aracı Ana Sayfa",
	}, "layouts/agent_layout")
}
