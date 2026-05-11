package auth

import (
	"github.com/gofiber/fiber/v2"
)

type Authenticator interface {
	Middleware() fiber.Handler
}
