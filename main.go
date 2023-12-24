package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/gin-gonic/gin"
	"github.com/ravelinejunior/golang_ecommerce/controllers"
	_ "github.com/ravelinejunior/golang_ecommerce/controllers"
	"github.com/ravelinejunior/golang_ecommerce/database"
	_ "github.com/ravelinejunior/golang_ecommerce/database"
	"github.com/ravelinejunior/golang_ecommerce/middleware"
	_ "github.com/ravelinejunior/golang_ecommerce/middleware"
	"github.com/ravelinejunior/golang_ecommerce/routes"
	_ "github.com/ravelinejunior/golang_ecommerce/routes"
)

// main is the entry point of the application
func main() {
	// get the port from the environment variable PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// create a new application instance
	app := controllers.NewApplication(database.ProductData(database.Client, "Products"), database.UserData(database.Client, "Users"))

	// create a new gin router
	router := gin.New()
	// use the gin logger middleware
	router.Use(gin.Logger())

	// register user routes
	routes.UserRoutes(router)
	// use the authentication middleware
	router.Use(middleware.Authentication())

	// register add to cart route
	router.GET("/addtocart", app.AddToCart())
	// register remove item route
	router.GET("/removeitem", app.RemoveItem())
	// register cart checkout route
	router.GET("/cartcheckout", app.BuyFromCart())
	// register instant buy route
	router.GET("/instantbuy", app.InstantBuy())
	router.GET("/listcart", controllers.GetItemFromCart())
	router.POST("/addaddress", controllers.AddAddress())
	router.PUT("/edithomeaddress", controllers.EditHomeAddress())
	router.PUT("/editworkaddress", controllers.EditWorkAddress())
	router.GET("/deleteaddresses", controllers.DeleteAddress())

	// start the server and log any errors
	log.Fatal(router.Run(":" + port))
}
