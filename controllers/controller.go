package controller

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	database "goRestaurantManager/database"
	helper "goRestaurantManager/helpers"
	models "goRestaurantManager/models"
	"log"
	"math"
	"net/http"
	"primitive"
	"strconv"
	"time"
)

func inTimeSpan(start, end, check time.Time) bool {
	return start.After(time.Now()) && end.After(start)
}

// ------------------------------- Database Connectors -------------------------------\\
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
	bytes, err := bcrypt.GeneratFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(expectedPass string, providedPass string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPass), []byte(expectedPass))
	check := true
	msg := ""
	if err != nil {
		msg = fmt.Sprintf("login and/or password are incorrect")
		check = false
	}
	return check, msg
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
		var user models.User

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 12
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchBy := bson.D{{"$match", bson.D{{}}}}
		projectBy := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items", bson.D{
					{"$slice", []interface{}{"$data", startIndex, recordPerPage}},
				}}}}}

		result, err := userCollection.Aggregate(contxt, mongo.Pipeline{
			matchBy,
			projectBy,
		})
		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("Listing all users was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}

		var allUsers []bson.M
		if err = result.All(contxt, &allUsers); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allUsers[0])
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
		userId := c.Param("user_id")

		var user models.User
		err := userCollection.FindOne(contxt, bson.M{"user_id": userId}).Decode(&user)

		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("Fetching user was unsuccessful")
			c.JSON(http.StatusInternalServerError, git.H{"error": msg})
		}
		c.JSON(http.StatusOK, user)
	}
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		pwd := HashPassword(*user.Password)
		user.Password = &pwd

		count, err := userCollection.CountDocuments(contxt, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			msg := fmt.Sprintf("Finding phone number was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if count > 0 {
			msg := fmt.Sprintf("Email or phone number already exists")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		token, refreshToken, _ := helper.GenerateAllTokens(
			*user.Email,
			*user.First_Name,
			*user.Last_Name,
			*user.User_id,
		)
		user.Token = &token
		user.Refresh_Token = &refreshToken

		result, err := userCollection.InsertOne(contxt, user)
		if err != nil {
			msg := fmt.Sprintf("User creation was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, result)

	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(contxt, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("User not found. Login incorrect")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		isValidPwd, msg = VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
		if isValidPwd != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}

		token, refreshToken, _ := helper.GenerateAllTokens(
			*foundUser.Email,
			*foundUser.First_Name,
			*foundUser.Last_Name,
			*foundUser.User_id,
		)

		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)
		c.JSON(http.StatusOK, foundUser)
	}
}

//------------------------------- Food-based functions -------------------------------\\

func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 12
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchBy := bson.D{{"$match", bson.D{{}}}}
		groupBy := bson.D{
			{"$group", bson.D{
				{"_id", bson.D{{"_id", "null"}}},
				{"total_count", bson.D{{"$sum", 1}}},
				{"data", bson.D{{"$push", "$$ROOT"}}},
			},
			},
		}

		projectBy := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"food_items", bson.D{
					{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
			},
			},
		}

		result, err := foodCollection.Aggregate(contxt, mongo.Pipeline{
			matchBy, groupBy, projectBy,
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

func toFixed(num float64, precision int) float64 {
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

			if err != nil {
				msg := fmt.Sprintf("Food item creation was unsuccessful")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
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
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)

		result, err := invoiceCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("Listing of invoices was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}

		var allInvoices []bson.M
		if err = result.All(contxt, &allInvoices); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allInvoices)
	}
}

func GetInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
		invoiceId := c.Param("invoice_id")

		var invoice models.Invoice

		err := invoiceCollection.FindOne(contxt, bson.M{"invoice_id": invoiceId}).Decode(&invoice)
		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("Listing of invoice details was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}

		var invoiceView InvoiceViewFormat

		allOrderItems, err := ItemByOrder(invoice.Order_id)
		invoiceView.Order_id = invoice.Order_id
		invoiceView.Payment_due_date = invoice.Payment_due_date

		invoiceView.Payment_method = "null"
		if invoice.Payment_method != nil {
			invoiceView.Payment_method = *invoice.Payment_method
		}

		invoiceView.Invoice_id = invoice.Invoice_id
		invoiceView.Payment_status = *&invoice.Payment_status
		invoiceView.Amount_due = allOrderItems[0]["amount_due"]
		invoiceView.Table_number = allOrderItems[0]["table_number"]
		invoiceView.Order_details = allOrderItems[0]["order_items"]

		c.JSON(http.StatusOK, invoiceView)
	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
		var invoice models.Invoice

		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var order models.Order

		err := orderCollection.FindOne(contxt, bson.M{"order_id": invoice.Order_id}).Decode(&order)
		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("Order was not found")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if invoice.Payment_status == nil {
			invoice.Payment_status = "PENDING"
		}

		invoice.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.Payment_due_date, _ = time.Parse(time.RFC3339, time.Now().AddDate(0, 0, 1).Format(time.RFC3339))
		invoice.ID = primitive.NewObjectID()
		invoice.Invoice_id = invoice.ID.Hex()

		validationErr := validate.Struct(invoice)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		result, err := invoiceCollection.InsertOne(contxt, invoice)
		if err != nil {
			msg := fmt.Sprintf("Invoice creation was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
		invoice_id := c.Param("invoice_id")

		var invoice models.Invoice

		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var updateObj primitive.D

		if invoice.Payment_method != nil {
			updateObj = append(updateObj, bson.E{"payment_method", invoice.Payment_method})
		}

		if invoice.Payment_status != nil {
			updateObj = append(updateObj, bson.E{"payment_status", invoice.Payment_status})
		}

		invoice.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", invoice.Updated_at})

		upsert := true
		filter := bson.M{"invoice_id": invoice_id}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		if invoice.Payment_status == nil {
			invoice.Payment_status = "PENDING"
		}

		result, err := invoiceCollection.UpdateOne(
			contxt,
			filter,
			bson.D{
				{"$set", updateObj},
			},
			&opt,
		)
		if err != nil {
			msg := fmt.Sprintf("Invoice update was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)
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
			updateObj = append(updateObj, bson.E{"updated_at", menu.Updated_at})

			upsert := true
			opt := options.UpdateOptions{
				Upsert: &upsert,
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
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)

		result, err := orderCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("Listing of orders was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}
		var allOrders []bson.M
		if err = result.All(contxt, &allOrders); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allOrders)

	}
}

func GetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
		orderId := c.Param("order_id")

		err := orderCollection.FindOne(contxt, bson.M{"order_id": orderId}).Decode(&order)
		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("Fetching order details was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}
		c.JSON(http.StatusOK, order)
	}
}

func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var table models.Table
		var order models.Order

		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(order)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		}

		if order.Table_id != nil {
			err := tableCollection.FindOne(contxt, bson.M{"table_id": order.Table_id}).Decode(&table)
			defer cancel()
			if err != nil {
				msg := fmt.Sprintf("Table was not found")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
		}

		order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		order.ID = primitive.NewObjectID()
		order.Order_id = order.ID.Hex()

		result, err := orderCollection.InsertOne(contxt, order)

		if err != nil {
			msg := fmt.Sprintf("Order creation was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}

func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var table models.Table
		var order models.Order

		var updateObj primitive.D

		orderId := c.Param("order_id")
		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if order.Table_id != nil {
			err := menuCollection.FindOne(contxt, bson.M{"table_id": food.Table_id}).Decode(&table)
			defer cancel()
			if err != nil {
				msg := fmt.Sprintf("Menu was not found")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			}

			order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			updateObj = append(updateObj, bson.E{"updated_at", food.Updated_at})

			upsert := true
			filter := bson.M{"order_id": orderId}
			opt := options.UpdateOptions{
				Upsert: &upsert,
			}

			result, err := orderCollection.UpdateOne(
				contxt,
				filter,
				bson.D{
					{"$set", updateObj},
				},
				&opt,
			)

			if err != nil {
				msg := fmt.Sprintf("Order update was unsuccessful")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}

			defer cancel()
			c.JSON(http.StatusOK, result)
		}
	}
}

//------------------------------- Table-based functions -------------------------------\\

func GetTables() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)

		result, err := orderCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("Listing tables was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}

		var allTables []bson.M
		if err = result.All(contxt, &allTables); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allTables)
	}
}

func GetTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
		tableId := c.Param("_id")
		var table models.Table

		err := tableCollection.FindOne(context, bson.M{"table_id": table_id}).Decode(&table)
		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("Fetching table was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}
		c.JSON(http.StatusOK, table)
	}
}

func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
		var table models.Table

		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(table)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		table.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		table.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		table.ID = primitive.NewObjectID()
		table.Order_id = table.ID.Hex()

		result, err := tableCollection.InsertOne(contxt, table)
		if err != nil {
			msg := fmt.Sprintf("Table creation was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, result)

	}
}

func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
		var table models.Table

		var updateObj primitive.D

		tableId := c.Param("table_id")
		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if table.Number_of_guests != nil {
			updateObj = append(updateObj, bson.E{"number_of_guests", table.Number_of_guests})
		}

		if table.Table_number != nil {
			updateObj = append(updateObj, bson.E{"table_number", table.Table_number})

		}

		table.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", table.Updated_at})

		upsert := true
		filter := bson.M{"table_id": tableId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := tableCollection.UpdateOne(
			contxt,
			filter,
			bson.D{
				{"$set", updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := fmt.Sprintf("Table update was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)

	}
}

//------------------------------- OrderItem-based functions -------------------------------\\

func orderItemCreator()

func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)

		result, err := orderItemCollection.Find(context.TODO(), bson.M{})

		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("Listing of order items was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		var allOrderItems []bson.M
		if err = result.All(contxt, &allOrderItems); err != nil {
			log.Fatal(err)
			return
		}
		c.JSON(http.StatusOK, allOrderItems)
	}
}

func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)

		orderItemId := c.Param("order_item_id")
		var orderItem models.OrderItem

		err := orderItemCollection.FindOne(contxt, bson.M{"orderItem_id": orderItemId}).Decode(&orderItem)
		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("Listing ordered items was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderId := c.Param("order_id")

		allOrderItems, err := ItemByOrder(orderId)
		if err != nil {
			msg := fmt.Sprintf("Listing of order items by order was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func ItemByOrder(id string) (OrderItems []primitive.M, err error) {
	var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)

	matchBy := bson.D{{"$match": bson.D{{"order_id", id}}}}
	lookupBy := bson.D{{"$lookup", bson.D{{"from", "food"},
		{"localField", "food_id"}, {"foreignField", "food_id"}, {"as", "food"}}}}
	unwind := bson.D{{"$unwind", bson.D{{"path", "$food"}, {"preserveNullAndEmptyArrays", true}}}}

	lookupOrder := bson.D{{"$lookup", bson.D{{"from", "order"},
		{"localField", "order_id"}, {"foreignField", "order_id"}, {"as", "order"}}}}
	unwindOrder := bson.D{{"$unwind", bson.D{{"path", "$order"}, {"preserveNullAndEmptyArrays", true}}}}

	lookupTable := bson.D{{"$lookup", bson.D{{"from", "table"},
		{"localField", "order.table_id"}, {"foreignField", "table_id"}, {"as", "table"}}}}
	unwindTable := bson.D{{"$unwind", bson.D{{"path", "$table"}, {"preserveNullAndEmptyArrays", true}}}}

	tableProject := bson.D{
		{"$project", bson.D{
			{"id", 0},
			{"amount", "$food.price"},
			{"total_count", 1},
			{"food_name", "$food.name"},
			{"food_image", "food.food_image"},
			{"table_number", "$table.table_number"},
			{"table_id", "$table.table_id"},
			{"order_id", "$order.order_id"},
			{"price", "$food.price"},
			{"quantity", 1},
		},
		},
	}

	groupBy := bson.D{{"$group", bson.D{{"_id", bson.D{
		{"order_id", "$order_id"}, {"table_id", "$table_id"},
		{"table_number", "$table_number"}}},
		{"amount_due", bson.D{{"$sum", "$amount"}}},
		{"total_count", bson.D{{"$sum", 1}}},
		{"order_items", bson.D{{"$push", 1}}},
	}}}

	groupProject := bson.D{
		{"$project", bson.D{
			{"id", 0},
			{"amount_due", 1},
			{"total_count", 1},
			{"table_number", "$_id.table_number"},
			{"order_items", 1},
		},
		},
	}

	result, err := orderItemCollection.Aggregate(contxt, mongo.Pipeline{
		matchBy,
		lookupBy,
		unwind,
		lookupOrder,
		unwindOrder,
		lookupTable,
		unwindTable,
		tableProject,
		groupBy,
		groupProject,
	})
	if err != nil {
		panic(err)
	}

	if err = result.All(contxt, &OrderItems); err != nil {
		panic(err)
	}

	defer cancel()

	return OrderItems, err

}

func CreateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)

		var orderItemPack orderItemPack
		var order models.Order

		if err := c.BindJSON(&orderItemPack); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		order.Order_Date, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		orderItemsToInsert := []interface{}{}
		order.Table_id = orderItemPack.Table_id
		order_id := OrderItemCreator(order)

		for _, orderItem := range orderItemPack.Order_items {
			orderItem.Order_id = order_id

			validationErr := validate.Struct(orderItem)
			if validationErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
				return
			}

			orderItem.ID = primitive.NewObjectID()
			orderItem.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.Order_item_id = orderItem.ID.Hex()

			var num = toFixed(*orderItem.Unit_price, 2)
			orderItem.Unit_price = &num
			orderItemsToInsert = append(orderItemsToInsert, orderItem)
		}

		insertedOrderItems, err := orderItemCollection.InsertMany(contxt, orderItemsToInsert)
		if err != nil {
			log.Fatal(err)
		}
		defer cancel()
		c.JSON(http.StatusOK, insertedOrderItems)
	}
}

func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)

		var orderItem models.OrderItem
		orderItemId := c.Param("order_item")
		var updateObj primitive.D

		if orderItem.Unit_price != nil {
			updateObj = append(updateObj, bson.E{"unit_price", *&orderItem.Unit_price})
		}

		if orderItem.Quantity != nil {
			updateObj = append(updateObj, bson.E{"quantity", *&orderItem.Quantity})
		}

		if orderItem.Food_id != nil {
			updateObj = append(updateObj, bson.E{"food_id", *&orderItem.Food_id})
		}

		orderItem.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", orderItem.Updated_at})

		upsert := true
		filter := bson.M{"order_item_id": orderItemId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := orderItemCollection.UpdateOne(
			contxt,
			filter,
			bson.D{
				{"$set", updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := fmt.Sprintf("Order item update was unsuccessful")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}
