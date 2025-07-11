package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"golang-restaurant-management/database"
	"golang-restaurant-management/middleware"
	"golang-restaurant-management/routes"
	"golang-restaurant-management/metrics"
	"golang-restaurant-management/controllers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	app := fiber.New()
	database.DBinstance()

	app.Use(logger.New())

	// Public routes (no auth)
	public := app.Group("/users")
	public.Post("/signup", controller.SignUp)
	public.Post("/login", controller.Login)

	// Register custom Prometheus metrics
	metrics.RegisterCustomMetrics()

	// Protected routes (require JWT)
	protected := app.Group("/", middleware.Authentication())
	routes.RegisterRoutes(protected)

	log.Fatal(app.Listen(":" + port))
}
