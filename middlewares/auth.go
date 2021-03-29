package middlewares

import (
	"strings"

	"github.com/faizainur/ipfs-api/services"
	"github.com/gofiber/fiber/v2"
)

type AuthMiddleware struct {
	AuthService *services.AuthService
}

func (a *AuthMiddleware) ValidateJwtToken(c *fiber.Ctx) error {
	authHeader := strings.Trim(c.Get("Authorization"), " ")
	if len(authHeader) < 2 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":  fiber.StatusUnauthorized,
			"error": "No JWT token provided",
		})
	}

	authToken := strings.Split(authHeader, " ")[1]
	isValid, data, err := a.AuthService.ValidateJwt(authToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	if !isValid {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized Access")
	}
	c.Locals("email", data.Email)
	c.Locals("userUid", data.UserUid)
	return c.Next()
}

func (a *AuthMiddleware) IntrospectAccessToken(c *fiber.Ctx) error {
	authHeader := strings.Trim(c.Get("Authorization"), " ")
	if len(authHeader) < 2 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":  fiber.StatusUnauthorized,
			"error": "No access token provided",
		})
	}

	accessToken := strings.Split(authHeader, " ")[1]
	isActive, data, err := a.AuthService.IntrospectTokenOauth2(accessToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	if !isActive {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized Access")
	}
	c.Locals("clientId", data.ClientID)
	c.Locals("scopes", data.Scope)
	c.Locals("email", data.Sub)
	return c.Next()
}
