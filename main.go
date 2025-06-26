package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/logger"
	//"golang-restaurant-management/database"
	"golang-restaurant-management/middleware"
	"golang-restaurant-management/routes"
	"golang-restaurant-management/database"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	app := fiber.New()
    database.DBinstance()
	// Middleware
    app.Use(logger.New())
	app.Use(middleware.Authentication())

	// Register routes
	routes.RegisterRoutes(app)

	// Start server
	log.Fatal(app.Listen(":" + port))
}
