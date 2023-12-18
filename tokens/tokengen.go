package tokens

import (
	"context"
	"log"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt"
	"github.com/ravelinejunior/golang_ecommerce/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var UserData *mongo.Collection = database.UserData(database.Client, "users")
var SECRET_KEY = os.Getenv("SECRET_KEY")

type SignedDetails struct {
	Email      string
	First_Name string
	Last_Name  string
	Uid        string
	jwt.StandardClaims
}

// TokenGenerator generates a JWT and a refresh JWT
func TokenGenerator(email string, firstName string, lastName string, uid string) (signedToken string, signedRefreshToken string, err error) {
	// create a new instance of SignedDetails
	claims := &SignedDetails{
		Email:      email,
		First_Name: firstName,
		Last_Name:  lastName,
		Uid:        uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	// create a new instance of SignedDetails for refresh token
	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	// create a new JWT using the claims and the secret key
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		return "", "", err
	}

	// create a new refresh JWT using the refresh claims and the secret key
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS384, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}

// ValidateToken takes a signed JWT token and returns the decoded claims and an error message
func ValidateToken(signedToken string) (claims *SignedDetails, message string) {
	// token is a jwt.Token struct that holds the parsed JWT token
	token, err := jwt.ParseWithClaims(signedToken, &SignedDetails{}, func(t *jwt.Token) (interface{}, error) {
		// SECRET_KEY is a string containing the secret key used to sign the JWT token
		return []byte(SECRET_KEY), nil
	})

	// if there is an error parsing the JWT token, return the error message
	if err != nil {
		message = err.Error()
		return
	}

	// claims is a pointer to a struct that holds the decoded JWT claims
	claims, ok := token.Claims.(*SignedDetails)
	// if the claims are not of type *SignedDetails, return an error message
	if !ok {
		message = "the token is invalid!"
		return
	}

	// check if the token has expired
	if claims.ExpiresAt < time.Now().Local().Unix() {
		message = "token is expired"
		return
	}

	// return the decoded claims and an empty error message
	return claims, message
}

// UpdateAllTokens updates all the tokens of a user
func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	// create a context with a timeout of 100 seconds
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	// create a variable to store the update object
	var updateObject primitive.D
	// add the "token" field to the update object with the given value
	updateObject = append(updateObject, bson.E{Key: "token", Value: signedToken})
	// add the "refresh_token" field to the update object with the given value
	updateObject = append(updateObject, bson.E{Key: "refresh_token", Value: signedRefreshToken})
	// parse the current time into a time.Time object with the RFC3339 format
	updateAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	// add the "updateat" field to the update object with the parsed time
	updateObject = append(updateObject, bson.E{Key: "updateat", Value: updateAt})
	// set the upsert flag to true
	upsert := true
	// create a filter based on the "user_id" field
	filter := bson.M{"user_id": userId}
	// create an update options object with the upsert flag
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	// update the user document with the given filter and update object using the update options
	_, err := UserData.UpdateOne(ctx, filter, bson.D{
		{Key: "$set", Value: updateObject},
	}, &opt)
	// defer the cancellation of the context
	defer cancel()
	// check for any errors
	if err != nil {
		log.Panic(err)
		return
	}
}
