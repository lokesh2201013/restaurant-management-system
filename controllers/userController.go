package controller

import (
	"context"
	//"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"golang-restaurant-management/database"
	 "golang-restaurant-management/helpers"
	"golang-restaurant-management/models"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers(c *fiber.Ctx) error {
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

	matchStage := bson.D{{Key: "$match", Value: bson.D{}}}
	projectStage := bson.D{{Key: "$project", Value: bson.D{
		{Key: "_id", Value: 0},
		{Key: "total_count", Value: 1},
		{Key: "user_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
	}}}

	result, err := userCollection.Aggregate(ctx, mongo.Pipeline{matchStage, projectStage})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error occurred while listing user items"})
	}

	var allUsers []bson.M
	if err := result.All(ctx, &allUsers); err != nil {
		return err
	}
	if len(allUsers) > 0 {
		return c.Status(fiber.StatusOK).JSON(allUsers[0])
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "no users found"})
}

func GetUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	userId := c.Params("user_id")
	var user models.User

	err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "User not found"})
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

func SignUp(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := validate.Struct(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil || count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email already exists"})
	}

	count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
	if err != nil || count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Phone already exists"})
	}

	token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, user.User_id)
	
	hashedPassword := HashPassword(*user.Password)
	user.Password = &hashedPassword
	user.Created_at = time.Now()
	user.Updated_at = time.Now()
	user.ID = primitive.NewObjectID()
	user.User_id = user.ID.Hex()
	user.Token = &token
	user.Refresh_Token = &refreshToken

	_, err = userCollection.InsertOne(ctx, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "User could not be created"})
	}

	return c.Status(fiber.StatusOK).JSON("User exisits")
}

func Login(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user models.User
	var foundUser models.User

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
	if !passwordIsValid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": msg})
	}

	token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)
	helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)

	return c.Status(fiber.StatusOK).JSON(foundUser)
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	if err != nil {
		return false, "Login or password is incorrect"
	}
	return true, ""
}
