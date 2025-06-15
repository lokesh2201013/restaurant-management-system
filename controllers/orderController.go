package controller

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"golang-restaurant-management/database"
	"golang-restaurant-management/models"
)

var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")

func GetOrders(c *fiber.Ctx) error {
	result, err := orderCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while listing order items"})
	}

	var allOrders []bson.M
	if err = result.All(ctx, &allOrders); err != nil {
		log.Fatal(err)
	}
	return c.Status(fiber.StatusOK).JSON(allOrders)
}

func GetOrder(c *fiber.Ctx) error {
	orderId := c.Params("order_id")
	var order models.Order

	err := orderCollection.FindOne(ctx, bson.M{"order_id": orderId}).Decode(&order)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while fetching the order"})
	}
	return c.Status(fiber.StatusOK).JSON(order)
}

func CreateOrder(c *fiber.Ctx) error {
	var table models.Table
	var order models.Order

	if err := c.BodyParser(&order); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	validationErr := validate.Struct(order)
	if validationErr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": validationErr.Error()})
	}

	if order.Table_id != nil {
		err := tableCollection.FindOne(ctx, bson.M{"table_id": order.Table_id}).Decode(&table)
		if err != nil {
			msg := fmt.Sprintf("message: Table was not found")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": msg})
		}
	}

	order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	order.ID = primitive.NewObjectID()
	order.Order_id = order.ID.Hex()

	result, insertErr := orderCollection.InsertOne(ctx, order)
	if insertErr != nil {
		msg := fmt.Sprintf("order was not created")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func UpdateOrder(c *fiber.Ctx) error {
	var table models.Table
	var order models.Order
	var updateObj primitive.D

	orderId := c.Params("order_id")
	if err := c.BodyParser(&order); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if order.Table_id != nil {
		err := tableCollection.FindOne(ctx, bson.M{"table_id": order.Table_id}).Decode(&table)
		if err != nil {
			msg := fmt.Sprintf("message: Table was not found")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": msg})
		}
		updateObj = append(updateObj, bson.E{Key: "table_id", Value: order.Table_id})
	}

	order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: order.Updated_at})

	upsert := true
	filter := bson.M{"order_id": orderId}
	opt := options.UpdateOptions{Upsert: &upsert}

	result, err := orderCollection.UpdateOne(
		ctx,
		filter,
		bson.D{{Key: "$set", Value: updateObj}},
		&opt,
	)
	if err != nil {
		msg := fmt.Sprintf("order update failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func OrderItemOrderCreator(order models.Order) string {
	order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.ID = primitive.NewObjectID()
	order.Order_id = order.ID.Hex()

	orderCollection.InsertOne(ctx, order)
	return order.Order_id
}
