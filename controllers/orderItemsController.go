package controller

import (
	"context"
	//"fmt"
	"log"
	"net/http"
	"time"

	"golang-restaurant-management/database"
	"golang-restaurant-management/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderItemPack struct {
	Table_id    *string             `json:"table_id"`
	Order_items []models.OrderItem `json:"order_items"`
}

var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client, "orderItem")

func GetOrderItems(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	result, err := orderItemCollection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while listing ordered items"})
	}
	var allOrderItems []bson.M
	if err = result.All(ctx, &allOrderItems); err != nil {
		log.Fatal(err)
		return err
	}
	return c.JSON(allOrderItems)
}

func GetOrderItemsByOrder(c *fiber.Ctx) error {
	orderId := c.Params("order_id")

	allOrderItems, err := ItemsByOrder(orderId)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while listing order items by order ID"})
	}
	return c.JSON(allOrderItems)
}

func ItemsByOrder(id string) (OrderItems []primitive.M, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	matchStage := bson.D{{"$match", bson.D{{"order_id", id}}}}
	lookupStage := bson.D{{"$lookup", bson.D{{"from", "food"}, {"localField", "food_id"}, {"foreignField", "food_id"}, {"as", "food"}}}}
	unwindStage := bson.D{{"$unwind", bson.D{{"path", "$food"}, {"preserveNullAndEmptyArrays", true}}}}

	lookupOrderStage := bson.D{{"$lookup", bson.D{{"from", "order"}, {"localField", "order_id"}, {"foreignField", "order_id"}, {"as", "order"}}}}
	unwindOrderStage := bson.D{{"$unwind", bson.D{{"path", "$order"}, {"preserveNullAndEmptyArrays", true}}}}

	lookupTableStage := bson.D{{"$lookup", bson.D{{"from", "table"}, {"localField", "order.table_id"}, {"foreignField", "table_id"}, {"as", "table"}}}}
	unwindTableStage := bson.D{{"$unwind", bson.D{{"path", "$table"}, {"preserveNullAndEmptyArrays", true}}}}

	projectStage := bson.D{{"$project", bson.D{
		{"id", 0},
		{"amount", "$food.price"},
		{"total_count", 1},
		{"food_name", "$food.name"},
		{"food_image", "$food.food_image"},
		{"table_number", "$table.table_number"},
		{"table_id", "$table.table_id"},
		{"order_id", "$order.order_id"},
		{"price", "$food.price"},
		{"quantity", 1},
	}}}

	groupStage := bson.D{{"$group", bson.D{{"_id", bson.D{{"order_id", "$order_id"}, {"table_id", "$table_id"}, {"table_number", "$table_number"}}}, {"payment_due", bson.D{{"$sum", "$amount"}}}, {"total_count", bson.D{{"$sum", 1}}}, {"order_items", bson.D{{"$push", "$$ROOT"}}}}}}

	projectStage2 := bson.D{{"$project", bson.D{
		{"id", 0},
		{"payment_due", 1},
		{"total_count", 1},
		{"table_number", "$_id.table_number"},
		{"order_items", 1},
	}}}

	result, err := orderItemCollection.Aggregate(ctx, mongo.Pipeline{
		matchStage,
		lookupStage,
		unwindStage,
		lookupOrderStage,
		unwindOrderStage,
		lookupTableStage,
		unwindTableStage,
		projectStage,
		groupStage,
		projectStage2})

	if err != nil {
		return nil, err
	}

	err = result.All(ctx, &OrderItems)
	return OrderItems, err
}

func GetOrderItem(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	orderItemId := c.Params("order_item_id")
	var orderItem models.OrderItem

	err := orderItemCollection.FindOne(ctx, bson.M{"order_item_id": orderItemId}).Decode(&orderItem)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while listing ordered item"})
	}
	return c.JSON(orderItem)
}

func UpdateOrderItem(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	orderItemId := c.Params("order_item_id")
	var orderItem models.OrderItem

	if err := c.BodyParser(&orderItem); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	filter := bson.M{"order_item_id": orderItemId}
	var updateObj primitive.D

	if orderItem.Unit_price != nil {
		updateObj = append(updateObj, bson.E{"unit_price", *orderItem.Unit_price})
	}
	if orderItem.Quantity != nil {
		updateObj = append(updateObj, bson.E{"quantity", *orderItem.Quantity})
	}
	if orderItem.Food_id != nil {
		updateObj = append(updateObj, bson.E{"food_id", *orderItem.Food_id})
	}

	orderItem.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{"updated_at", orderItem.Updated_at})

	upsert := true
	opt := options.UpdateOptions{Upsert: &upsert}

	result, err := orderItemCollection.UpdateOne(ctx, filter, bson.D{{"$set", updateObj}}, &opt)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Order item update failed"})
	}
	return c.JSON(result)
}

func CreateOrderItem(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var orderItemPack OrderItemPack
	var order models.Order

	if err := c.BodyParser(&orderItemPack); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	order.Order_Date, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.Table_id = orderItemPack.Table_id
	order_id := OrderItemOrderCreator(order)

	orderItemsToBeInserted := []interface{}{}

	for _, orderItem := range orderItemPack.Order_items {
		orderItem.Order_id = order_id
		if err := validate.Struct(orderItem); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		orderItem.ID = primitive.NewObjectID()
		orderItem.Created_at = time.Now()
		orderItem.Updated_at = time.Now()
		orderItem.Order_item_id = orderItem.ID.Hex()
		num := toFixed(*orderItem.Unit_price, 2)
		orderItem.Unit_price = &num
		orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)
	}

	insertedOrderItems, err := orderItemCollection.InsertMany(ctx, orderItemsToBeInserted)
	if err != nil {
		log.Fatal(err)
	}

	return c.JSON(insertedOrderItems)
}
