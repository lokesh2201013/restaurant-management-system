package controller

import (
	"context"
	//"fmt"
	"golang-restaurant-management/database"
	"golang-restaurant-management/models"
	//"log"
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

var validate = validator.New()

func GetFoods(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	recordPerPage, err := strconv.Atoi(c.Query("recordPerPage", "10"))
	if err != nil || recordPerPage < 1 {
		recordPerPage = 10
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	startIndex := (page - 1) * recordPerPage
	if queryStartIndex := c.Query("startIndex"); queryStartIndex != "" {
		if idx, err := strconv.Atoi(queryStartIndex); err == nil {
			startIndex = idx
		}
	}

	matchStage := bson.D{
    {Key: "$match", Value: bson.D{}},
}

groupStage := bson.D{
    {Key: "$group", Value: bson.D{
        {Key: "_id", Value: nil},
        {Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
        {Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
    }},
}

projectStage := bson.D{
    {Key: "$project", Value: bson.D{
        {Key: "_id", Value: 0},
        {Key: "total_count", Value: 1},
        {Key: "food_items", Value: bson.D{
            {Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}},
        }},
    }},
}

	result, err := foodCollection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while listing food items"})
	}

	var allFoods []bson.M
	if err := result.All(ctx, &allFoods); err != nil || len(allFoods) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error decoding food data"})
	}

	return c.Status(fiber.StatusOK).JSON(allFoods[0])
}

func GetFood(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	foodId := c.Params("food_id")
	var food models.Food

	err := foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&food)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while fetching the food item"})
	}

	return c.Status(fiber.StatusOK).JSON(food)
}

func CreateFood(c *fiber.Ctx) error {
	var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var menu models.Menu
	var food models.Food

	if err := c.BodyParser(&food); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if validationErr := validate.Struct(food); validationErr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": validationErr.Error()})
	}

	err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "menu was not found"})
	}

	now := time.Now()
	food.Created_at = now
	food.Updated_at = now
	food.ID = primitive.NewObjectID()
	food.Food_id = food.ID.Hex()
	num := toFixed(*food.Price, 2)
	food.Price = &num

	result, insertErr := foodCollection.InsertOne(ctx, food)
	if insertErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "food item was not created"})
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func UpdateFood(c *fiber.Ctx) error {
	var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var menu models.Menu
	var food models.Food

	foodId := c.Params("food_id")

	if err := c.BodyParser(&food); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var updateObj primitive.D

	if food.Name != nil {
		updateObj = append(updateObj, bson.E{Key: "name", Value: food.Name})
	}
	if food.Price != nil {
		updateObj = append(updateObj, bson.E{Key: "price", Value: food.Price})
	}
	if food.Food_image != nil {
		updateObj = append(updateObj, bson.E{Key: "food_image", Value: food.Food_image})
	}
	if food.Menu_id != nil {
		err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "menu was not found"})
		}
		updateObj = append(updateObj, bson.E{Key: "menu", Value: food.Price})
	}

	food.Updated_at = time.Now()
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: food.Updated_at})

	upsert := true
	filter := bson.M{"food_id": foodId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	result, err := foodCollection.UpdateOne(
		ctx,
		filter,
		bson.D{{Key: "$set", Value: updateObj}},
		&opt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "food item update failed"})
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
