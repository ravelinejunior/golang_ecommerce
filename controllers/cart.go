package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/ravelinejunior/golang_ecommerce/database"
	"github.com/ravelinejunior/golang_ecommerce/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var UserCollection *mongo.Collection = database.UserData(database.Client, "Users")
var ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")
var Validate = validator.New()

type Application struct {
	prodCollection *mongo.Collection
	userCollection *mongo.Collection
}

// NewApplication initializes the application
func NewApplication(prodCollection, userCollection *mongo.Collection) *Application {
	return &Application{
		prodCollection: prodCollection,
		userCollection: userCollection,
	}
}

// AddToCart adds a product to the cart
func (app *Application) AddToCart() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// productQueryID is the id of the product to be added to the cart
		productQueryID := ctx.Query("id")
		if productQueryID == "" {
			log.Println("product id is empty")

			// return an error with status code 400 Bad Request
			_ = ctx.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}

		// userQueryID is the id of the user who is adding the product to the cart
		userQueryID := ctx.Query("userID")
		if userQueryID == "" {
			log.Println("user id is empty")
			// return an error with status code 400 Bad Request
			_ = ctx.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		// convert the product id to a primitive.ObjectID type
		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Println(err)
			// return an internal server error
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// add the product to the cart
		err = database.AddProductToCart(context.Background(), app.prodCollection, app.userCollection, productID, userQueryID)
		if err != nil {
			// return an internal server error
			ctx.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		// return a status code of 200 OK and a message indicating that the product was added to the cart
		ctx.IndentedJSON(http.StatusOK, "Product Successfully Added To The Cart")
	}
}

// RemoveItem removes an item from the cart
func (app *Application) RemoveItem() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// productQueryID is the id of the product to be removed from the cart
		productQueryID := ctx.Query("id")
		if productQueryID == "" {
			log.Println("product id is empty")

			// return an error with status code 400 Bad Request
			_ = ctx.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}

		// userQueryID is the id of the user who is removing the product from the cart
		userQueryID := ctx.Query("userID")
		if userQueryID == "" {
			log.Println("user id is empty")
			// return an error with status code 400 Bad Request
			_ = ctx.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		// convert the product id to a primitive.ObjectID type
		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Println(err)
			// return an internal server error
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// remove the product from the cart
		err = database.RemoveCartItem(context.Background(), app.prodCollection, app.userCollection, productID, userQueryID)
		if err != nil {
			// return an internal server error
			ctx.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		// return a status code of 200 OK and a message indicating that the product was removed from the cart
		ctx.IndentedJSON(http.StatusOK, "Product Successfully Removed From The Cart")
	}
}

// GetItemFromCart returns the items in the cart of a user
func GetItemFromCart() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		// userID is the id of the user whose cart items are to be returned
		userID := gCtx.Query("id")

		// if the user id is empty, return an error
		if userID == "" {
			gCtx.Header("Content-Type", "application/json")
			gCtx.JSON(http.StatusNotFound, gin.H{"erro": "invalid id"})
			gCtx.Abort()
			return
		}

		// convert the user id to a primitive.ObjectID type
		usertId, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			log.Println(err)
			gCtx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		// create a context with a timeout of 100 seconds
		var ctx context.Context
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// find the user in the database
		var filledCart models.User
		err = UserCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: usertId}}).Decode(&filledCart)
		if err != nil {
			// if the user is not found, return an error
			if err == mongo.ErrNoDocuments {
				gCtx.JSON(http.StatusNotFound, "Not found")
				return
			}
			log.Println(err)
			gCtx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// create a pipeline to filter, unwind, and group the cart items
		filterMatch := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: usertId}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$usercart"}}}}
		grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$usercart.price"}}}}}}

		// run the pipeline and retrieve the results
		pointCursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{filterMatch, unwind, grouping})
		if err != nil {
			log.Println(err)
			gCtx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// create a slice to store the results
		var listing []bson.M
		if err = pointCursor.All(ctx, &listing); err != nil {
			log.Println(err)
			gCtx.AbortWithStatus(http.StatusInternalServerError)
		}

		// iterate over the results and return them in JSON format
		for _, json := range listing {
			gCtx.IndentedJSON(http.StatusOK, json["total"])
			gCtx.IndentedJSON(http.StatusOK, filledCart.UserCart)
		}
	}
}

// BuyFromCart handles the buying process of an item from the cart
func (app *Application) BuyFromCart() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// userQueryID is the id of the user who is buying the product
		userQueryID := ctx.Query("id")

		// if the user id is empty, return an error
		if userQueryID == "" {
			log.Panicln("user id is empty")
			_ = ctx.AbortWithError(http.StatusBadRequest, errors.New("UserID is empty"))
		}

		// create a context with a timeout of 100 seconds
		var contx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		// buy the product from the cart
		err := database.BuyItemFromCart(contx, app.userCollection, userQueryID)
		if err != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, err)
		}

		// return a status code of 200 OK and a message indicating that the product was bought successfully
		ctx.IndentedJSON(http.StatusOK, "Successfully placed the order")
	}
}

// InstantBuy handles the buying process of an item from the cart
func (app *Application) InstantBuy() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// productQueryID is the id of the product to be bought
		productQueryID := ctx.Query("id")
		if productQueryID == "" {
			log.Println("product id is empty")

			// return an error with status code 400 Bad Request
			_ = ctx.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}

		// userQueryID is the id of the user who is buying the product
		userQueryID := ctx.Query("userID")
		if userQueryID == "" {
			log.Println("user id is empty")
			// return an error with status code 400 Bad Request
			_ = ctx.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		// convert the product id to a primitive.ObjectID type
		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Println(err)
			// return an internal server error
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// create a context with a timeout of 5 seconds
		var contx, cancel = context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		// buy the product from the cart
		err = database.InstantBuyer(contx, app.prodCollection, app.userCollection, productID, userQueryID)
		if err != nil {
			// return an internal server error
			ctx.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		// return a status code of 200 OK and a message indicating that the product was bought successfully
		ctx.IndentedJSON(http.StatusOK, "Successfully placed the order")
	}
}
