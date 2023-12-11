package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ravelinejunior/golang_ecommerce/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddAddress() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		userID := gCtx.Query("id")

		if userID == "" {
			gCtx.Header("Content-Type", "application/json")
			gCtx.JSON(http.StatusNotFound, gin.H{"erro": "invalid code"})
			gCtx.Abort()
			return
		}

		address, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			gCtx.IndentedJSON(http.StatusInternalServerError, "Internal Server Error")
		}

		var addresses models.Address

		addresses.Address_ID = primitive.NewObjectID()

		if err = gCtx.BindJSON(&addresses); err != nil {
			gCtx.IndentedJSON(http.StatusNotAcceptable, err.Error())
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		matchFilter := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: address}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$address"}}}}
		group := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "address_id"}, {Key: "count", Value: bson.D{primitive.E{Key: "$sum", Value: 1}}}}}}
		pointCursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{matchFilter, unwind, group})

		if err != nil {
			gCtx.IndentedJSON(http.StatusInternalServerError, "Internal Server Error")
		}
	}
}

func EditHomeAddress() gin.HandlerFunc {

}

func EditWorkAddress() gin.HandlerFunc {

}

func DeleteAddress() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		userId := gCtx.Query("id")

		if userId == "" {
			gCtx.Header("Content-Type", "application/json")
			gCtx.JSON(http.StatusNotFound, gin.H{"Error": "Invalid Search Index"})
			gCtx.Abort()
			return
		}

		addresses := make([]models.Address, 0)
		usertId, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			gCtx.IndentedJSON(http.StatusInternalServerError, "Ineternal Server Error")
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		filter := bson.D{primitive.E{Key: "_id", Value: usertId}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}
		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			gCtx.IndentedJSON(http.StatusNotFound, "Wrong command")
			return
		}
		defer cancel()
		ctx.Done()
		gCtx.IndentedJSON(http.StatusOK, "Successfully Deleted")
	}
}
