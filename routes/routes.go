package routes

import (
	"github.com/gofiber/fiber/v2"
	controller "golang-restaurant-management/controllers"
)

func RegisterRoutes(app *fiber.App) {
	// User routes
	app.Get("/users", controller.GetUsers)
	app.Get("/users/:user_id", controller.GetUser)
	app.Post("/users/signup", controller.SignUp)
	app.Post("/users/login", controller.Login)

	// Food routes
	app.Get("/foods", controller.GetFoods)
	app.Get("/foods/:food_id", controller.GetFood)
	app.Post("/foods", controller.CreateFood)
	app.Patch("/foods/:food_id", controller.UpdateFood)

	// Menu routes
	app.Get("/menus", controller.GetMenus)
	app.Get("/menus/:menu_id", controller.GetMenu)
	app.Post("/menus", controller.CreateMenu)
	app.Patch("/menus/:menu_id", controller.UpdateMenu)

	// Table routes
	app.Get("/tables", controller.GetTables)
	app.Get("/tables/:table_id", controller.GetTable)
	app.Post("/tables", controller.CreateTable)
	app.Patch("/tables/:table_id", controller.UpdateTable)

	// Order routes
	app.Get("/orders", controller.GetOrders)
	app.Get("/orders/:order_id", controller.GetOrder)
	app.Post("/orders", controller.CreateOrder)
	app.Patch("/orders/:order_id", controller.UpdateOrder)

	// OrderItem routes
	app.Get("/orderItems", controller.GetOrderItems)
	app.Get("/orderItems/:orderItem_id", controller.GetOrderItem)
	app.Get("/orderItems-order/:order_id", controller.GetOrderItemsByOrder)
	app.Post("/orderItems", controller.CreateOrderItem)
	app.Patch("/orderItems/:orderItem_id", controller.UpdateOrderItem)

	// Invoice routes
	app.Get("/invoices", controller.GetInvoices)
	app.Get("/invoices/:invoice_id", controller.GetInvoice)
	app.Post("/invoices", controller.CreateInvoice)
	app.Patch("/invoices/:invoice_id", controller.UpdateInvoice)
}
