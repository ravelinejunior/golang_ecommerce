package controllers

import (
	"context"
	"fmt"
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

		var addressinfo []bson.M
		if err = pointCursor.All(ctx, &addressinfo); err != nil {
			panic(err)
		}

		var size int32
		for _, addressNo := range addressinfo {
			count := addressNo["count"]
			size = count.(int32)
		}
		if size < 2 {
			filter := bson.D{primitive.E{Key: "_id", Value: address}}
			update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}
			_, err := UserCollection.UpdateOne(ctx, filter, update)

			if err != nil {
				fmt.Println(err)
			}

		} else {
			gCtx.IndentedJSON(http.StatusNotFound, "Not Allowed")
		}

		defer cancel()
		ctx.Done()
	}
}

func EditHomeAddress() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		userId := gCtx.Query("id")

		if userId == "" {
			gCtx.Header("Content-Type", "application/json")
			gCtx.JSON(http.StatusNotFound, gin.H{"Error": "Invalid Search Index"})
			gCtx.Abort()
			return
		}

		usertId, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			gCtx.IndentedJSON(http.StatusInternalServerError, "Internal Server Error")
		}

		var editAddress models.Address
		if err := gCtx.BindJSON(&editAddress); err != nil {
			gCtx.IndentedJSON(http.StatusBadRequest, err.Error())
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		filter := bson.D{primitive.E{Key: "_id", Value: usertId}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address.0.house_name", Value: editAddress.House}, {Key: "address.0.street_name", Value: editAddress.Street}, {Key: "address.0.city_name", Value: editAddress.City}, {Key: "address.0.pin_code", Value: editAddress.PinCode}}}}
		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			gCtx.IndentedJSON(http.StatusInternalServerError, "Something went wrong while updating Home Address")
			return
		}
		defer cancel()
		ctx.Done()

		gCtx.IndentedJSON(http.StatusOK, "Home address successfully updated")
	}
}

func EditWorkAddress() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		userId := gCtx.Query("id")

		if userId == "" {
			gCtx.Header("Content-Type", "application/json")
			gCtx.JSON(http.StatusNotFound, gin.H{"Error": "Invalid Search Index"})
			gCtx.Abort()
			return
		}

		usertId, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			gCtx.IndentedJSON(http.StatusInternalServerError, "Internal Server Error")
		}

		var editAddress models.Address
		if err := gCtx.BindJSON(&editAddress); err != nil {
			gCtx.IndentedJSON(http.StatusBadRequest, err.Error())
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		filter := bson.D{primitive.E{Key: "_id", Value: usertId}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address.1.house_name", Value: editAddress.House}, {Key: "address.1.street_name", Value: editAddress.Street}, {Key: "address.1.city_name", Value: editAddress.City}, {Key: "address.1.pin_code", Value: editAddress.PinCode}}}}
		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			gCtx.IndentedJSON(http.StatusInternalServerError, "Something went wrong while updating work address")
			return
		}
		defer cancel()
		ctx.Done()

		gCtx.IndentedJSON(http.StatusOK, "Work address successfully updated")
	}
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
