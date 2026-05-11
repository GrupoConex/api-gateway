package proxy

import "github.com/gofiber/fiber/v2"

type Router interface {
	Handle(c *fiber.Ctx) error
}
