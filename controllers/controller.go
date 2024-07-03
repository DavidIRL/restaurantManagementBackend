package controller

import (
	"context"
	"github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"goRestaurantManager/database"
	"goRestaurantManager/models"
	"gopkg.in/bluesuncorp/validator.v5"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"time"
)
var inTimeSpan(start, end, check time.Time) bool {
    return start.After(time.Now()) && end.After(start)
}


//------------------------------- Database Connectors -------------------------------\\
var validate = validator.New()
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")
var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client, "invoice")
var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")
var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")
var tableCollection *mongo.Collection = database.OpenCollection(database.Client, "table")
var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client, "order_item")

//------------------------------- User-based functions -------------------------------\\

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

//------------------------------- Food-based functions -------------------------------\\

func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {
        var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)

        recPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
        if err != nil || recPerPage < 1 {
            recPerPage = 12
        }

        page, err := strconv.Atoi(c.Query("page"))
        if err != nil || page < 1 {
            page = 1
        }

        startIndex := (page-1) * recPerPage
        startIndex, err = strconv.Atoi(c.Query("startIndex"))

        matchBy := bson.D{{"$match", bson.D{{}}}}
        groupBy := bson.D{{"$group", bson.D{
            {"_id", bson.D{{"_id", "null"}}},
            {"total_count", bson.D{{"$sum, 1"}}},{"data", bson.D{{"$push", "$$ROOT"}}} }}}
        projectBy := bson.D{
            {
                "$project", bson.D{
                    {"_id", 0},
                    {"total_count", 1},
                    {"food_items", bson.D{
                        {"$slice", []interface{}{"$data", startIndex, recPerPage}}}}
                }
            }
        }

        result, err := foodCollection.Aggregate(contxt, mongo.Pipeline{
            matchBy, groupBy, projectBy
        })
        defer cancel()
        if err != nil {
            msg := fmt.Sprintf("Listing food items was unsuccessful")
            c.JSON{http.StatusInternalServerError, gin.H{"error": msg}}
        }
        var allFoods []bson.M
        if err = result.All(contxt, &allFoods); err != nil {
            log.Fatal(err)
        }
        c.JSON(http.StatusOK, allFoods[0])
    }
}

func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
		foodId := c.Param("food_id")
		var food models.Food

		err := foodCollection.FindOne(contxt, bson.M{"food_id": foodId}).Decode(&food)
		defer cancel()
		if err != nil {
            msg := fmt.Sprintf("An error occurred while fetching the food item.")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}
		c.JSON(http.StatusOK, food)
	}
}

func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
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
		food.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.ID = primitive.NewObjectID()
		food.Food_id = food.ID.Hex()
		var num = toFixed(*food.Price, 2)
		food.Price = &num

		result, err := foodCollection.InsertOne(contxt, food)
		if err != nil {
			msg := fmt.Sprintf("Food item was not created successfully")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
            return
		}
		defer cancel()
		c.JSON(httb.StatusOK, result)
	}
}

func round(num float64) int {
    return int(num + math.Copysign(0.5, num))
}

fun toFixed(num float64, precision int) float64 {
    output := math.Pow(10, float64(precision))
    return float64(round(num*output)) / output
}

func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
        var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
        var menu models.Menu
        var food models.Food

        foodId := c.Param("food_id")

        if err := c.BindJSON(&food); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        var updateObj primitive.D

        if food.Name != nil {
            updateObj = append(updateObj, bson.E{"name", food.Name})
        }

        if food.Price != nil {
            updateObj = append(updateObj, bson.E{"price", food.Price})
        }

        if food.Food_image != nil {
            updateObj = append(updateObj, bson.E{"food_image", food.Food_image})
        }

        if food.Menu_id != nil {
            err := menuCollection.FindOne(contxt, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
            defer cancel()
            if err != nil {
                msg := fmt.Sprintf("Menu fetch was unsuccessful")
                c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
                return
            }
            food.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
            food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
            food.Id = primitive.NewObjectID()
            food.Food_id = food.ID.Hex()
            var num = toFixed(*food.Price, 2)
            food.Price = &num

            result, insertErr != nil {
                msg := fmt.Sprintf("Food item creation was unsuccessful")
            }

        }

        food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
        updateObj = append(updateObj, bson.E{"updated_at", food.Updated_at})

        upsert := true
        filter := bson.M{"food_id": foodId}
        opt := options.UpdateOptions{
            Upsert: &upsert,
        }

        result, err := foodCollection.UpdateOne(
            contxt,
            filter,
            bson.D{
                {"$set", updateObj},
            },
            &opt,
        )
        if err != nil {
            msg := fmt.Sprintf("Food item update was unsuccessful")
            c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
            return
        }
        c.JSON(http.StatusOK, result)
	}
}

//------------------------------- Invoice-based functions -------------------------------\\

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

//------------------------------- Menu-based functions -------------------------------\\

func GetMenus() gin.HandlerFunc {
	return func(c *gin.Context) {
        var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
        result, err := menuCollection.Find(context.TODO(), bson.M{})
        defer cancel()
        if err != nil {
            msg := fmt.Sprintf("error occurred while fetching menu items.")
            c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
        }
        var allMenus []bson.m
        if err = result.All(contxt, &allMenus); err != nil {
            log.Fatal(err)
        }
        c.JSON(http.StatusOK, allMenus)
	}
}

func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
        var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
        menuId := c.Param("menu_id")
        var menu models.Menu

        err := menuCollection.FindOne(contxt, bson.M{"menu_id": menuId})
        defer cancel()
        if err != nil {
            msg := Sprintf("error occurred whie fetching requested menu")
            c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
        }
        c.JSON(http.StatusOK, menu)
	}
}

func CreateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
        var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
        var menu models.Menu

        if err := c.BindJSON(&menu); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        validationErr := validate.Struct(menu)
        if validationErr != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
            return
        }
        
        menu.Created_at, _ = time.Parse(time.RFC3330, time.Now().Format(time.RFC3339))
        menu.Updated_at, _ = time.Parse(time.RFC3330, time.Now().Format(time.RFC3339))
        menu.ID = primitive.NewObjectID()
        menu.Menu_id = menu.ID.Hex()

        result, err := menuCollection.InsertOne(contxt, menu)
        if err != nil {
            msg := fmt.Sprintf("Menu item was not created successfully")
            c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
            return
        }
        defer cancel()
        c.JSON(http.StatusOK, result)
        defer cancel()


	}
}

func UpdateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
        var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
        var menu models.Menu

        if err := c.BindJSON(&menu); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        menuID := c.Param("menu_id")
        filter := bson.M{"menu_id": menuID}

        var updateObj primitive.D

        if menu.Start_Date != nil && menu.End_Date != nil {
            if !inTimeSpan(*menu.Start_Date, *menu.End_Date, time.Now()) {
                msg := "Incorrect time provided"
                c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
                defer cancel()
                return
            }

            updateObj = append(updateObj, bson.E{"start_date", menu.Start_Date})
            updateObj = append(updateObj, bson.E{"end_date", menu.End_Date})

            if menu.Name != "" {
                updateObj = append(updateObj, bson.E{"name", menu.Name})
            }
            if menu.Category != "" {
                updateObj = append(updateObj, bson.E{"name", menu.Name})
            }
             
		    menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
            updateObj = append(updateObj bson.E{"updated_at", menu.Updated_at})

            upsert := true
            opt := options.UpdateOptions{
                Upsert : &upsert,
            }

            result, err := menuCollection.UpdateOne(
                contxt,
                filter,
                bson.D{
                    {"$set", updateObj},
                },
                &opt,
            )
            if err != nil {
                msg := "Menu update was unsuccessful"
                c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
            }
            defer cancel()
            c.JSON(http.StatusOK, result)
        }   
	}
}

//------------------------------- Order-based functions -------------------------------\\

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

//------------------------------- Table-based functions -------------------------------\\

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

//------------------------------- OrderItem-based functions -------------------------------\\

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
