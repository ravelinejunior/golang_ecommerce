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

func NewApplication(prodCollection, userCollection *mongo.Collection) *Application {
	return &Application{
		prodCollection: prodCollection,
		userCollection: userCollection,
	}
}

func (app *Application) AddToCart() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productQueryID := ctx.Query("id")
		if productQueryID == "" {
			log.Println("product id is empty")

			_ = ctx.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}

		userQueryID := ctx.Query("userID")
		if userQueryID == "" {
			log.Println("user id is empty")
			_ = ctx.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)

		if err != nil {
			log.Println(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var contx, cancel = context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		err = database.AddProductToCart(contx, app.prodCollection, app.userCollection, productID, userQueryID)
		if err != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, err)
		}
		ctx.IndentedJSON(http.StatusOK, "Product Successfully Added To The Cart")

	}
}

func (app *Application) RemoveItem() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productQueryID := ctx.Query("id")
		if productQueryID == "" {
			log.Println("product id is empty")

			_ = ctx.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}

		userQueryID := ctx.Query("userID")
		if userQueryID == "" {
			log.Println("user id is empty")
			_ = ctx.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)

		if err != nil {
			log.Println(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var contx, cancel = context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		database.RemoveCartItem(contx, app.prodCollection, app.userCollection, productID, userQueryID)

		if err != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		ctx.IndentedJSON(http.StatusOK, "Successfully removed item from cart")

	}
}

func GetItemFromCart() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		userID := gCtx.Query("id")

		if userID == "" {
			gCtx.Header("Content-Type", "application/json")
			gCtx.JSON(http.StatusNotFound, gin.H{"erro": "invalid id"})
			gCtx.Abort()
			return
		}

		usertId, _ := primitive.ObjectIDFromHex(userID)

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var filledCart models.User
		err := UserCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: usertId}}).Decode(&filledCart)

		if err != nil {
			log.Println(err)
			gCtx.JSON(http.StatusInternalServerError, "Not found")
			return
		}

		filterMatch := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: usertId}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$usercart"}}}}
		grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$usercart.price"}}}}}}
		pointCursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{filterMatch, unwind, grouping})

		if err != nil {
			log.Println(err)
		}
		var listing []bson.M
		if err = pointCursor.All(ctx, &listing); err != nil {
			log.Println(err)
			gCtx.AbortWithStatus(http.StatusInternalServerError)
		}
		for _, json := range listing {
			gCtx.IndentedJSON(http.StatusOK, json["total"])
			gCtx.IndentedJSON(http.StatusOK, filledCart.UserCart)
		}

		ctx.Done()

	}
}

func (app *Application) BuyFromCart() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userQueryID := ctx.Query("id")

		if userQueryID == "" {
			log.Panicln("user id is empty")
			_ = ctx.AbortWithError(http.StatusBadRequest, errors.New("UserID is empty"))
		}

		var contx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		err := database.BuyItemFromCart(contx, app.userCollection, userQueryID)
		if err != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, err)
		}

		ctx.IndentedJSON(http.StatusOK, "Successfully placed the order")

	}
}

func (app *Application) InstantBuy() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productQueryID := ctx.Query("id")
		if productQueryID == "" {
			log.Println("product id is empty")

			_ = ctx.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}

		userQueryID := ctx.Query("userID")
		if userQueryID == "" {
			log.Println("user id is empty")
			_ = ctx.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)

		if err != nil {
			log.Println(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var contx, cancel = context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		err = database.InstantBuyer(contx, app.prodCollection, app.userCollection, productID, userQueryID)

		if err != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		ctx.IndentedJSON(http.StatusOK, "Successfully placed to order")
	}
}
