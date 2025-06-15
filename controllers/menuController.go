package controller

import (
	"context"
	//"fmt"
	"golang-restaurant-management/database"
	"golang-restaurant-management/models"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")

func GetMenus(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	result, err := menuCollection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while listing the menu items"})
	}

	var allMenus []bson.M
	if err := result.All(ctx, &allMenus); err != nil {
		log.Fatal(err)
	}

	return c.Status(fiber.StatusOK).JSON(allMenus)
}

func GetMenu(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	menuId := c.Params("menu_id")
	var menu models.Menu

	err := menuCollection.FindOne(ctx, bson.M{"menu_id": menuId}).Decode(&menu)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while fetching the menu"})
	}

	return c.Status(fiber.StatusOK).JSON(menu)
}

func CreateMenu(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var menu models.Menu

	if err := c.BodyParser(&menu); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if validationErr := validate.Struct(menu); validationErr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": validationErr.Error()})
	}

	menu.Created_at = time.Now()
	menu.Updated_at = time.Now()
	menu.ID = primitive.NewObjectID()
	menu.Menu_id = menu.ID.Hex()

	result, err := menuCollection.InsertOne(ctx, menu)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Menu item was not created"})
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func inTimeSpan(start, end, check time.Time) bool {
	return start.After(time.Now()) && end.After(start)
}

func UpdateMenu(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var menu models.Menu
	if err := c.BodyParser(&menu); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	menuId := c.Params("menu_id")
	filter := bson.M{"menu_id": menuId}
	var updateObj primitive.D

	if menu.Start_Date != nil && menu.End_Date != nil {
		if !inTimeSpan(*menu.Start_Date, *menu.End_Date, time.Now()) {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "kindly retype the time"})
		}
		updateObj = append(updateObj, bson.E{Key: "start_date", Value: menu.Start_Date})
		updateObj = append(updateObj, bson.E{Key: "end_date", Value: menu.End_Date})
	}

	if menu.Name != "" {
		updateObj = append(updateObj, bson.E{Key: "name", Value: menu.Name})
	}
	if menu.Category != "" {
		updateObj = append(updateObj, bson.E{Key: "category", Value: menu.Category})
	}

	menu.Updated_at = time.Now()
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: menu.Updated_at})

	upsert := true
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	result, err := menuCollection.UpdateOne(
		ctx,
		filter,
		bson.D{{Key: "$set", Value: updateObj}},
		&opt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Menu update failed"})
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
