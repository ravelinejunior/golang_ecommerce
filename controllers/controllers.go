package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ravelinejunior/golang_ecommerce/models"
	"go.mongodb.org/mongo-driver/bson"
)

func HashPassword(password string) string {
	return ""
}

func VerifyPassword(userPassword string, typedPass string) (bool, string) {
	return true, ""
}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON{http.StatusBadRequest, gin.H{"error": err.Error()}}
			return
		}

		validationErr := Validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
			return
		}

		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
		}

		count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})

		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "this phone number is already in use"})
			return
		}

	}
}

func Login(c *gin.Context) gin.HandlerFunc {

}

func AddProduct(c *gin.Context) gin.HandlerFunc {

}

func SearchProduct(c *gin.Context) gin.HandlerFunc {

}

func SearchProductByQuery(c *gin.Context) gin.HandlerFunc {

}
