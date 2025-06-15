package controller

import (
	"context"
	"fmt"
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

type InvoiceViewFormat struct {
	Invoice_id       string      `json:"invoice_id"`
	Payment_method   string      `json:"payment_method"`
	Order_id         string      `json:"order_id"`
	Payment_status   *string     `json:"payment_status"`
	Payment_due      interface{} `json:"payment_due"`
	Table_number     interface{} `json:"table_number"`
	Payment_due_date time.Time   `json:"payment_due_date"`
	Order_details    interface{} `json:"order_details"`
}

var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client, "invoice")

func GetInvoices(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	result, err := invoiceCollection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while listing invoice items"})
	}

	var allInvoices []bson.M
	if err := result.All(ctx, &allInvoices); err != nil {
		log.Fatal(err)
	}

	return c.Status(fiber.StatusOK).JSON(allInvoices)
}

func GetInvoice(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	invoiceId := c.Params("invoice_id")
	var invoice models.Invoice

	err := invoiceCollection.FindOne(ctx, bson.M{"invoice_id": invoiceId}).Decode(&invoice)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while listing invoice item"})
	}

	var invoiceView InvoiceViewFormat
	allOrderItems, err := ItemsByOrder(invoice.Order_id)
	if err != nil || len(allOrderItems) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "unable to fetch order items"})
	}

	invoiceView.Order_id = invoice.Order_id
	invoiceView.Payment_due_date = invoice.Payment_due_date
	invoiceView.Payment_method = "null"
	if invoice.Payment_method != nil {
		invoiceView.Payment_method = *invoice.Payment_method
	}
	invoiceView.Invoice_id = invoice.Invoice_id
	invoiceView.Payment_status = invoice.Payment_status
	invoiceView.Payment_due = allOrderItems[0]["payment_due"]
	invoiceView.Table_number = allOrderItems[0]["table_number"]
	invoiceView.Order_details = allOrderItems[0]["order_items"]

	return c.Status(fiber.StatusOK).JSON(invoiceView)
}

func CreateInvoice(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var invoice models.Invoice
	if err := c.BodyParser(&invoice); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var order models.Order
	err := orderCollection.FindOne(ctx, bson.M{"order_id": invoice.Order_id}).Decode(&order)
	if err != nil {
		msg := fmt.Sprintf("message: Order was not found")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}

	status := "PENDING"
	if invoice.Payment_status == nil {
		invoice.Payment_status = &status
	}

	now := time.Now()
	invoice.Payment_due_date = now.AddDate(0, 0, 1)
	invoice.Created_at = now
	invoice.Updated_at = now
	invoice.ID = primitive.NewObjectID()
	invoice.Invoice_id = invoice.ID.Hex()

	if validationErr := validate.Struct(invoice); validationErr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": validationErr.Error()})
	}

	result, insertErr := invoiceCollection.InsertOne(ctx, invoice)
	if insertErr != nil {
		msg := fmt.Sprintf("invoice item was not created")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func UpdateInvoice(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var invoice models.Invoice
	invoiceId := c.Params("invoice_id")

	if err := c.BodyParser(&invoice); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	filter := bson.M{"invoice_id": invoiceId}
	var updateObj primitive.D

	if invoice.Payment_method != nil {
		updateObj = append(updateObj, bson.E{Key: "payment_method", Value: invoice.Payment_method})
	}
	if invoice.Payment_status != nil {
		updateObj = append(updateObj, bson.E{Key: "payment_status", Value: invoice.Payment_status})
	}

	invoice.Updated_at = time.Now()
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: invoice.Updated_at})

	upsert := true
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	result, err := invoiceCollection.UpdateOne(
		ctx,
		filter,
		bson.D{{Key: "$set", Value: updateObj}},
		&opt,
	)
	if err != nil {
		msg := fmt.Sprintf("invoice item update failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}

	return c.Status(fiber.StatusOK).JSON(result)
}