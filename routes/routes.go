package routes

import (
	"github.com/gofiber/fiber/v2"
	"golang-restaurant-management/controllers"
)

func RegisterRoutes(router fiber.Router) {
	// User routes
	user := router.Group("/users")
	user.Get("/", controller.GetUsers)
	user.Get("/:user_id", controller.GetUser)

	// Food routes
	food := router.Group("/foods")
	food.Get("/", controller.GetFoods)
	food.Get("/:food_id", controller.GetFood)
	food.Post("/", controller.CreateFood)
	food.Patch("/:food_id", controller.UpdateFood)

	// Menu routes
	menu := router.Group("/menus")
	menu.Get("/", controller.GetMenus)
	menu.Get("/:menu_id", controller.GetMenu)
	menu.Post("/", controller.CreateMenu)
	menu.Patch("/:menu_id", controller.UpdateMenu)

	// Table routes
	table := router.Group("/tables")
	table.Get("/", controller.GetTables)
	table.Get("/:table_id", controller.GetTable)
	table.Post("/", controller.CreateTable)
	table.Patch("/:table_id", controller.UpdateTable)

	// Order routes
	order := router.Group("/orders")
	order.Get("/", controller.GetOrders)
	order.Get("/:order_id", controller.GetOrder)
	order.Post("/", controller.CreateOrder)
	order.Patch("/:order_id", controller.UpdateOrder)

	// OrderItem routes
	orderItem := router.Group("/orderItems")
	orderItem.Get("/", controller.GetOrderItems)
	orderItem.Get("/:orderItem_id", controller.GetOrderItem)
	orderItem.Get("-order/:order_id", controller.GetOrderItemsByOrder)
	orderItem.Post("/", controller.CreateOrderItem)
	orderItem.Patch("/:orderItem_id", controller.UpdateOrderItem)

	// Invoice routes
	invoice := router.Group("/invoices")
	invoice.Get("/", controller.GetInvoices)
	invoice.Get("/:invoice_id", controller.GetInvoice)
	invoice.Post("/", controller.CreateInvoice)
	invoice.Patch("/:invoice_id", controller.UpdateInvoice)
}
