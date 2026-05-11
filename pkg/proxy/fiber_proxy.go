package proxy

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

type fiberRouter struct {
	routes map[string]string
}

func NewFiberRouter(routes map[string]string) Router {
	return &fiberRouter{
		routes: routes,
	}
}

func (r *fiberRouter) Handle(c *fiber.Ctx) error {
	path := c.Params("*")
	parts := strings.SplitN(path, "/", 2)
	
	if len(parts) == 0 || parts[0] == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Service not specified"})
	}

	serviceName := strings.ToLower(parts[0])
	targetBase, ok := r.routes[serviceName]
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Service not found"})
	}

	remainingPath := ""
	if len(parts) > 1 {
		remainingPath = "/" + parts[1]
	}

	url := targetBase + remainingPath
	return proxy.Do(c, url)
}
