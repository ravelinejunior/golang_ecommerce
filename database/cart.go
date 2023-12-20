package database

import (
	"context"
	"errors"
	"log"

	"github.com/ravelinejunior/golang_ecommerce/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrCantFindProduct    = errors.New("can't find the product")
	ErrCantDecodeProducts = errors.New("can't find the product")
	ErrUserIdsNotValid    = errors.New("this user is not valid")
	ErrCantUpdateUser     = errors.New("can't add this product to the cart")
	ErrCantRemoveCartItem = errors.New("can't remove this item from the cart")
	ErrCantGetItem        = errors.New("unable to get item from the cart")
	ErrCantBuyCartItem    = errors.New("can't update the purchase")
)

// AddProductToCart adds a product to the cart of a user
func AddProductToCart(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	// search the product collection for the given product id
	searchFromDB, err := prodCollection.Find(ctx, bson.M{"_id": productID})
	if err != nil {
		log.Println(err)
		return ErrCantFindProduct
	}

	// decode the products into a slice of ProductUser structs
	var productCart []models.ProductUser
	err = searchFromDB.All(ctx, &productCart)
	if err != nil {
		log.Println(err)
		return ErrCantDecodeProducts
	}

	// convert the user id to a primitive.ObjectID
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIdsNotValid
	}

	// create a filter to search for the given user id
	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	// create an update to add the products to the user's cart
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "usercart", Value: bson.D{{Key: "$each", Value: productCart}}}}}}

	// update the user document with the new cart items
	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return ErrCantUpdateUser
	}
	return nil
}

// RemoveCartItem removes an item from the cart of a user
func RemoveCartItem(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	// convert the user id to a primitive.ObjectID
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		// log the error and return that the user id is not valid
		log.Println(err)
		return ErrUserIdsNotValid
	}

	// create a filter to search for the given user id
	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	// create an update to remove the given product id from the user's cart
	update := bson.M{"$pull": bson.M{"usercart": bson.M{"_id": productID}}}

	// update the user document with the new cart items
	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		// log the error and return that the cart item could not be removed
		log.Println(err)
		return ErrCantRemoveCartItem
	}

	// update the product document with the removed cart item
	_, err = prodCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		// log the error and return that the cart item could not be removed
		log.Println(err)
		return ErrCantRemoveCartItem
	}
	return nil
}
func BuyItemFromCart() {

}
func InstantBuyer() {

}
