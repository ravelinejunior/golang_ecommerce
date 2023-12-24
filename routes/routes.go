package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ravelinejunior/golang_ecommerce/controllers"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("users/signup", controllers.Signup())
	incomingRoutes.POST("users/login", controllers.Login())
	incomingRoutes.POST("/admin/add_product", controllers.ProductViewerAdmin())
	incomingRoutes.GET("/users/product_view", controllers.SearchProduct())
	incomingRoutes.GET("/users/search", controllers.SearchProductByQuery())
}
