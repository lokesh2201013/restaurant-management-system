package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"

	"golang-restaurant-management/database"
	"golang-restaurant-management/middleware"
	"golang-restaurant-management/routes"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	app := fiber.New()

	// Middleware
	app.Use(middleware.Logger())
	app.Use(middleware.Authentication())

	// Register routes
	routes.RegisterRoutes(app)

	// Start server
	log.Fatal(app.Listen(":" + port))
}
