package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/ravelinejunior/golang_ecommerce/database"
	"github.com/ravelinejunior/golang_ecommerce/models"
	generate "github.com/ravelinejunior/golang_ecommerce/tokens"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var UserCollection *mongo.Collection = database.UserData(database.Client, "Users")
var ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")
var Validate = validator.New()

// HashPassword godoc
// @Summary Generate a hashed password
// @Description Generate a hashed password using the bcrypt algorithm
// @param password string true "password to be hashed"
// @return string
// @return error
// @example
//
//	password := "password"
//	hashedPassword, err := HashPassword(password)
//	if err != nil {
//		log.Panic(err)
//	}
//	fmt.Println(hashedPassword)
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

// VerifyPassword godoc
// @Summary Verify user password
// @Description Verify user password
// @Tags Auth
// @Accept json
// @Produce json
// @Param user_password path string true "User password"
// @Param typed_pass path string true "User typed password"
// @Success 200 {boolean} true
// @Failure 400 {object} models.Error
// @Failure 500 {object} models.Error
// @Router /auth/verify_password/{user_password}/{typed_pass} [get]
func VerifyPassword(userPassword string, typedPass string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(typedPass), []byte(userPassword))
	valid := true
	message := ""

	if err != nil {
		message = "Login or Password is incorrect"
		valid = false
	}
	return valid, message
}

// Signup godoc
// @Summary Signup user
// @Description Signup user
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body models.User true "User object"
// @Success 201 {object} models.User
// @Failure 400 {object} models.Error
// @Failure 500 {object} models.Error
// @Router /auth/signup [post]
func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

		password := HashPassword(*user.Password)
		user.Password = &password

		user.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_ID = user.ID.Hex()
		token, refreshToken, _ := generate.TokenGenerator(*user.Email, *user.First_Name, *user.Last_Name, user.User_ID)
		user.Token = &token
		user.Refresh_Token = &refreshToken
		user.UserCart = make([]models.ProductUser, 0)
		user.Address_Details = make([]models.Address, 0)
		user.Order_Status = make([]models.Order, 0)
		_, inserterr := UserCollection.InsertOne(ctx, user)
		if inserterr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create the user"})
			return
		}
		defer cancel()

		c.JSON(http.StatusCreated, "Successfully signed in.")
	}
}

// Login godoc
// @Summary Login user
// @Description Login user
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body models.User true "User object"
// @Success 200 {object} models.User
// @Failure 400 {object} models.Error
// @Failure 500 {object} models.Error
// @Router /auth/login [post]
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		var foundUser models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login or password incorrect"})
		}

		PasswordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()

		if !PasswordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			fmt.Println(msg)
			return
		}

		token, refreshToken, _ := generate.TokenGenerator(*foundUser.Email, *foundUser.First_Name, *foundUser.Last_Name, foundUser.User_ID)
		defer cancel()

		generate.UpdateAllTokens(token, refreshToken, foundUser.User_ID)
		c.IndentedJSON(http.StatusOK, foundUser)

	}
}

// ProductViewerAdmin godoc
// @Summary Add a new product to the database
// @Description Adds a new product to the database
// @Tags Products
// @Accept json
// @Produce json
// @Param product body models.Product true "Product object"
// @Success 200 {object} models.Product
// @Failure 400 {object} models.Error
// @Failure 500 {object} models.Error
// @Router /product/admin [post]
func ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var products models.Product
		defer cancel()
		if err := c.BindJSON(&products); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		products.Product_ID = primitive.NewObjectID()
		_, anyerr := ProductCollection.InsertOne(ctx, products)
		if anyerr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Not Created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, "Successfully added our Product Admin!!")
	}
}

// SearchProduct godoc
// @Summary Search for products
// @Description Search for products based on the search query
// @Tags Products
// @Accept json
// @Produce json
// @Param name query string true "Search query"
// @Success 200 {array} models.Product
// @Failure 404 {object} models.Error
// @Failure 500 {object} models.Error
// @Router /product/search [get]
func SearchProduct() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var productList []models.Product
		var contx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		cursor, err := ProductCollection.Find(contx, bson.D{{}})
		if err != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, "Something went wrong, try again!")
			return
		}

		err = cursor.All(contx, &productList)

		if err != nil {
			log.Println(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		defer cursor.Close(ctx)

		if err := cursor.Err(); err != nil {
			log.Println(err)
			ctx.IndentedJSON(http.StatusBadRequest, "invalid search")
			return
		}

		defer cancel()

		ctx.IndentedJSON(http.StatusOK, productList)
	}
}

// SearchProductByQuery godoc
// @Summary Search for products
// @Description Search for products based on the search query
// @Tags Products
// @Accept json
// @Produce json
// @Param name query string true "Search query"
// @Success 200 {array} models.Product
// @Failure 404 {object} models.Error
// @Failure 500 {object} models.Error
// @Router /product/search [get]
func SearchProductByQuery() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var searchProducts []models.Product
		queryParam := ctx.Query("name")

		// check if params is empty
		if queryParam == "" {
			log.Println("query is empty")
			ctx.Header("Content-Type", "application/json")
			ctx.JSON(http.StatusNotFound, gin.H{"Error": "Invalid search index"})
			ctx.Abort()
			return
		}

		var contx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		searchQueryDb, err := ProductCollection.Find(contx, bson.M{"product_name": bson.M{"regex": queryParam}})

		if err != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, "Something went wrong while trying fetch data")
			return
		}

		err = searchQueryDb.All(contx, &searchProducts)
		if err != nil {
			log.Println(err)
			ctx.IndentedJSON(http.StatusInternalServerError, "invalid search")
			return
		}

		defer searchQueryDb.Close(contx)

		if err := searchQueryDb.Err(); err != nil {
			log.Println(err)
			ctx.IndentedJSON(http.StatusInternalServerError, "invalid request")
			return
		}

		defer cancel()
		ctx.IndentedJSON(http.StatusOK, searchProducts)
	}
}
