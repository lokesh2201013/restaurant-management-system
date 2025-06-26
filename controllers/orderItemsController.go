package controller

import (
	"context"
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

	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "order_id", Value: id}}}}
	lookupStage := bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "food"},
		{Key: "localField", Value: "food_id"},
		{Key: "foreignField", Value: "food_id"},
		{Key: "as", Value: "food"},
	}}}
	unwindStage := bson.D{{Key: "$unwind", Value: bson.D{
		{Key: "path", Value: "$food"},
		{Key: "preserveNullAndEmptyArrays", Value: true},
	}}}

	lookupOrderStage := bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "order"},
		{Key: "localField", Value: "order_id"},
		{Key: "foreignField", Value: "order_id"},
		{Key: "as", Value: "order"},
	}}}
	unwindOrderStage := bson.D{{Key: "$unwind", Value: bson.D{
		{Key: "path", Value: "$order"},
		{Key: "preserveNullAndEmptyArrays", Value: true},
	}}}

	lookupTableStage := bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "table"},
		{Key: "localField", Value: "order.table_id"},
		{Key: "foreignField", Value: "table_id"},
		{Key: "as", Value: "table"},
	}}}
	unwindTableStage := bson.D{{Key: "$unwind", Value: bson.D{
		{Key: "path", Value: "$table"},
		{Key: "preserveNullAndEmptyArrays", Value: true},
	}}}

	projectStage := bson.D{{Key: "$project", Value: bson.D{
		{Key: "id", Value: 0},
		{Key: "amount", Value: "$food.price"},
		{Key: "total_count", Value: 1},
		{Key: "food_name", Value: "$food.name"},
		{Key: "food_image", Value: "$food.food_image"},
		{Key: "table_number", Value: "$table.table_number"},
		{Key: "table_id", Value: "$table.table_id"},
		{Key: "order_id", Value: "$order.order_id"},
		{Key: "price", Value: "$food.price"},
		{Key: "quantity", Value: 1},
	}}}

	groupStage := bson.D{{Key: "$group", Value: bson.D{
		{Key: "_id", Value: bson.D{
			{Key: "order_id", Value: "$order_id"},
			{Key: "table_id", Value: "$table_id"},
			{Key: "table_number", Value: "$table_number"},
		}},
		{Key: "payment_due", Value: bson.D{{Key: "$sum", Value: "$amount"}}},
		{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
		{Key: "order_items", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
	}}}

	projectStage2 := bson.D{{Key: "$project", Value: bson.D{
		{Key: "id", Value: 0},
		{Key: "payment_due", Value: 1},
		{Key: "total_count", Value: 1},
		{Key: "table_number", Value: "$_id.table_number"},
		{Key: "order_items", Value: 1},
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
		projectStage2,
	})
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
		updateObj = append(updateObj, bson.E{Key: "unit_price", Value: *orderItem.Unit_price})
	}
	if orderItem.Quantity != nil {
		updateObj = append(updateObj, bson.E{Key: "quantity", Value: *orderItem.Quantity})
	}
	if orderItem.Food_id != nil {
		updateObj = append(updateObj, bson.E{Key: "food_id", Value: *orderItem.Food_id})
	}

	orderItem.Updated_at = time.Now()
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: orderItem.Updated_at})

	upsert := true
	opt := options.UpdateOptions{Upsert: &upsert}

	result, err := orderItemCollection.UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: updateObj}}, &opt)
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

	order.Order_Date = time.Now()
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
