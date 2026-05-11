package auth

import (
	"log"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type keycloakAuthenticator struct {
	keyfunc keyfunc.Keyfunc
}

func NewKeycloakAuthenticator(url, realm string) Authenticator {
	jwksURL := url + "/realms/" + realm + "/protocol/openid-connect/certs"
	
	kf, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		log.Fatalf("Critical: Could not initialize JWKS: %v", err)
	}

	return &keycloakAuthenticator{
		keyfunc: kf,
	}
}

func (k *keycloakAuthenticator) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token format",
			})
		}

		token, err := jwt.Parse(parts[1], k.keyfunc.Keyfunc)
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid claims",
			})
		}

		c.Locals("user", claims)
		
		return c.Next()
	}
}
