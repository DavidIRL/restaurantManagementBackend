package routes

import(
    "github.com/gin-gonic/gin"
    controller "goRestaurantManager"
)


func FoodRoutes(incomingRoutes *gin.engine) {
    incomingRoutes.GET("/foods", controller.GetFoods())
    incomingRoutes.GET("/foods/:food_id", controller.GetFood())
    incomingRoutes.POST("/foods", controller.CreateFood())
    incomingRoutes.GET("/foods/:food_id", controller.UpdateFood())
}
