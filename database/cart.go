package database

import (
	"context"
	"errors"
	"log"
	"time"

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

// BuyItemFromCart fetches the cart of the user, finds the cart total, creates an order with the items, adds the order to the user collection, adds the items in the cart to the order list, and empties the cart.
func BuyItemFromCart(ctx context.Context, userCollection *mongo.Collection, userID string) error {
	// Convert the user ID to a primitive.ObjectID.
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIdsNotValid
	}

	// Initialize variables to hold the retrieved cart items and the order.
	var getCartItems models.User
	var orderCart models.Order

	// Set the order ID, ordered at time, and initialize the order cart.
	orderCart.Order_ID = primitive.NewObjectID()
	orderCart.Ordered_At = time.Now()
	orderCart.Order_Cart = make([]models.ProductUser, 0)
	orderCart.Payment_Method.COD = true

	// Unwind the user cart and group by the user ID to find the cart total.
	unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$usercart"}}}}
	grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$usercart.price"}}}}}}

	// Run the aggregation pipeline to retrieve the cart items and total.
	currentResults, err := userCollection.Aggregate(ctx, mongo.Pipeline{unwind, grouping})
	ctx.Done()
	if err != nil {
		panic(err)
	}

	// Initialize a slice to hold the retrieved cart items.
	var getUserCart []bson.M
	if err = currentResults.All(ctx, &getUserCart); err != nil {
		panic(err)
	}

	// Initialize a variable to hold the total cart price.
	var totalPrice int32

	// Iterate through the retrieved cart items to find the total price.
	for _, userItem := range getUserCart {
		price := userItem["total"]
		totalPrice = price.(int32)
	}

	// Set the order price to the total cart price.
	orderCart.Price = int(totalPrice)

	// Create a filter to search for the given user ID.
	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	// Create an update to add the order to the user's orders.
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "orders", Value: orderCart}}}}

	// Update the user document with the new order.
	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Println(err)
	}

	// Find the user document with the given user ID.
	userCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&getCartItems)
	if err != nil {
		log.Println(err)
	}

	// Create a filter to search for the given user ID again.
	secondFilter := bson.D{primitive.E{Key: "_id", Value: id}}

	// Create an update to add the cart items to the order list.
	secondUpdate := bson.M{"$push": bson.M{"orders.$[].order_list": bson.M{"$each": getCartItems.UserCart}}}

	// Update the user document with the new order list.
	_, err = userCollection.UpdateOne(ctx, secondFilter, secondUpdate)
	if err != nil {
		log.Println(err)
	}

	// Initialize an empty slice for the user cart.
	userCartEmpty := make([]models.ProductUser, 0)

	// Create a filter to search for the given user ID a third time.
	thirdFilter := bson.D{primitive.E{Key: "_id", Value: id}}

	// Create an update to set the user cart to the empty slice.
	thirdUpdate := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "usercart", Value: userCartEmpty}}}}

	// Update the user document with the empty cart.
	_, err = userCollection.UpdateOne(ctx, thirdFilter, thirdUpdate)
	if err != nil {
		return ErrCantBuyCartItem
	}

	return nil
}

// InstantBuyer adds a product to the cart of a user
func InstantBuyer(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	// Convert the user ID to a primitive.ObjectID.
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIdsNotValid
	}

	// Initialize variables to hold the retrieved cart items and the order.
	var getCartItems models.User
	var orderCart models.Order

	// Set the order ID, ordered at time, and initialize the order cart.
	orderCart.Order_ID = primitive.NewObjectID()
	orderCart.Ordered_At = time.Now()
	orderCart.Order_Cart = make([]models.ProductUser, 0)
	orderCart.Payment_Method.COD = true

	// Unwind the user cart and group by the user ID to find the cart total.
	unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$usercart"}}}}
	grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$usercart.price"}}}}}}

	// Run the aggregation pipeline to retrieve the cart items and total.
	currentResults, err := userCollection.Aggregate(ctx, mongo.Pipeline{unwind, grouping})
	ctx.Done()
	if err != nil {
		panic(err)
	}

	// Initialize a slice to hold the retrieved cart items.
	var getUserCart []bson.M
	if err = currentResults.All(ctx, &getUserCart); err != nil {
		panic(err)
	}

	// Initialize a variable to hold the total cart price.
	var totalPrice int32

	// Iterate through the retrieved cart items to find the total price.
	for _, userItem := range getUserCart {
		price := userItem["total"]
		totalPrice = price.(int32)
	}

	// Set the order price to the total cart price.
	orderCart.Price = int(totalPrice)

	// Create a filter to search for the given user ID.
	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	// Create an update to add the order to the user's orders.
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "orders", Value: orderCart}}}}

	// Update the user document with the new order.
	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Println(err)
	}

	// Find the user document with the given user ID.
	userCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&getCartItems)
	if err != nil {
		log.Println(err)
	}

	// Create a filter to search for the given user ID again.
	secondFilter := bson.D{primitive.E{Key: "_id", Value: id}}

	// Create an update to add the cart items to the order list.
	secondUpdate := bson.M{"$push": bson.M{"orders.$[].order_list": bson.M{"$each": getCartItems.UserCart}}}

	// Update the user document with the new order list.
	_, err = userCollection.UpdateOne(ctx, secondFilter, secondUpdate)
	if err != nil {
		log.Println(err)
	}

	// Initialize an empty slice for the user cart.
	userCartEmpty := make([]models.ProductUser, 0)

	// Create a filter to search for the given user ID a third time.
	thirdFilter := bson.D{primitive.E{Key: "_id", Value: id}}

	// Create an update to set the user cart to the empty slice.
	thirdUpdate := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "usercart", Value: userCartEmpty}}}}

	// Update the user document with the empty cart.
	_, err = userCollection.UpdateOne(ctx, thirdFilter, thirdUpdate)
	if err != nil {
		return ErrCantBuyCartItem
	}

	return nil
}
