package controller

import (
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/mongo"
    "goRestaurantManager/database"
    "goRestaurantManager/models"
    "time"
    "context"
)


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
        defer cancel ()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while fetching the food item"})
        }
        c.JSON(http.StatusOK, food)
    }
}


func CreateFood() gin.HandlerFunc {
    return func(c *gin.Context) {

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

