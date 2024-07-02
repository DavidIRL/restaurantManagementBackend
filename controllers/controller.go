package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"goRestaurantManager/database"
	"goRestaurantManager/models"
	"gopkg.in/bluesuncorp/validator.v5"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"time"
)

var validate = validator.New()
var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

//------------------------------- User-based functions-------------------------------\\

func HashPassword(password string) string {

}

func VerifyPassword(expectedPass string, providedPass string) (bool, string) {

}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

//------------------------------- Food-based functions-------------------------------\\

func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		foodId := c.Param("food_id")
		var food models.Food

		err := foodCollection.FindOne(contxt, bson.M{"food_id": foodId}).Decode(&food)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while fetching the food item"})
		}
		c.JSON(http.StatusOK, food)
	}
}

func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu
		var food models.Food

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(food)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
		if err != nil {
			msg := fmt.Sprintf("menu was not found")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}
		food.Created_at, _ = time.Parse(time.RFC3339, time.Now()).Format(time.RFC3339)
		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now()).Format(time.RFC3339)
		food.ID = primitive.NewObjectID()
		food.Food_id = food.ID.Hex()
		var num = toFixed(*food.Price, 2)
		food.Price = &num

		result, err := foodCollection.InsertOne(contxt, food)
		if err != nil {
			msg := fmt.Sprintf("Food item was not created successfully")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}
		defer cancel()
		c.JSON(httb.StatusOK, result)
	}
}

func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

//------------------------------- Invoice-based functions-------------------------------\\

func GetInvoices() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func GetInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

//------------------------------- Menu-based functions-------------------------------\\

func GetMenus() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func CreateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func UpdateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

//------------------------------- Order-based functions-------------------------------\\

func GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func GetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

//------------------------------- Table-based functions-------------------------------\\

func GetTables() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func GetTable() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

//------------------------------- OrderItem-based functions-------------------------------\\

func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func ItemByOrder(id string) (OrderItems []primitive.M, err error) {
	return
}

func CreateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
