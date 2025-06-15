package middleware

import (
	//"fmt"
	helper "golang-restaurant-management/helpers"
	"github.com/gofiber/fiber/v2"
)

func Authentication() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientToken := c.Get("token")
		if clientToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "No Authorization header provided",
			})
		}

		claims, err := helper.ValidateToken(clientToken)
		if err != "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err,
			})
		}

		// Set user data in context locals
		c.Locals("email", claims.Email)
		c.Locals("first_name", claims.First_name)
		c.Locals("last_name", claims.Last_name)
		c.Locals("uid", claims.Uid)

		return c.Next()
	}
}
