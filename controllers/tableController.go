package controller

import (
	"context"
	//"fmt"
	"golang-restaurant-management/database"
	"golang-restaurant-management/models"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var tableCollection *mongo.Collection = database.OpenCollection(database.Client, "table")

func GetTables(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	result, err := tableCollection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while listing tables"})
	}

	var allTables []bson.M
	if err := result.All(ctx, &allTables); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(allTables)
}

func GetTable(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	tableId := c.Params("table_id")
	var table models.Table

	err := tableCollection.FindOne(ctx, bson.M{"table_id": tableId}).Decode(&table)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while fetching the table"})
	}
	return c.JSON(table)
}

func CreateTable(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var table models.Table
	if err := c.BodyParser(&table); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := validate.Struct(table); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	table.ID = primitive.NewObjectID()
	table.Table_id = table.ID.Hex()
	now := time.Now().UTC()
	table.Created_at = now
	table.Updated_at = now

	result, err := tableCollection.InsertOne(ctx, table)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "table item was not created"})
	}
	return c.JSON(result)
}

func UpdateTable(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	tableId := c.Params("table_id")
	var table models.Table
	if err := c.BodyParser(&table); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var updateObj primitive.D
	if table.Number_of_guests != nil {
		updateObj = append(updateObj, bson.E{"number_of_guests", table.Number_of_guests})
	}
	if table.Table_number != nil {
		updateObj = append(updateObj, bson.E{"table_number", table.Table_number})
	}
	table.Updated_at = time.Now().UTC()
	updateObj = append(updateObj, bson.E{"updated_at", table.Updated_at})

	opts := options.Update().SetUpsert(true)
	result, err := tableCollection.UpdateOne(ctx, bson.M{"table_id": tableId}, bson.D{{"$set", updateObj}}, opts)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "table update failed"})
	}
	return c.JSON(result)
}
