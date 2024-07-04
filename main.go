package main

import (
	database "github.com/DavidIRL/restaurantManagementBackend/database"
	middleware "github.com/DavidIRL/restaurantManagementBackend/middleware"
	routes "github.com/DavidIRL/restaurantManagementBackend/routes"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(middleware.Authentication())
	routes.UserRoutes(router)

	routes.FoodRoutes(router)
	routes.MenuRoutes(router)
	routes.TableRoutes(router)
	routes.OrderRoutes(router)
	routes.OrderItemRoutes(router)
	routes.InvoiceRoutes(router)

	router.Run(":" + port)
}
