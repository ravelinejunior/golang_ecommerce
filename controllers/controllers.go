package controllers

import (
	"github.com/gin-gonic/gin"
)

func HashPassword(password string) string {
	return ""
}

func VerifyPassword(userPassword string, typedPass string) (bool, string) {
	return true, ""
}

func Signup(c *gin.Context) gin.HandlerFunc {

}
func Login(c *gin.Context) gin.HandlerFunc {

}
func AddProduct(c *gin.Context) gin.HandlerFunc {

}
func SearchProduct(c *gin.Context) gin.HandlerFunc {

}
func SearchProductByQuery(c *gin.Context) gin.HandlerFunc {

}
