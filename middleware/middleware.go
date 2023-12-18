package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	token "github.com/ravelinejunior/golang_ecommerce/tokens"
)

// Authentication is a middleware function that verifies the JWT token sent in the request header
// and sets the user's email and ID in the context. If the token is invalid, the request is aborted
// with an error.
func Authentication() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		// Get the JWT token from the request header
		clientToken := gCtx.Request.Header.Get("token")
		if clientToken == "" {
			gCtx.JSON(http.StatusInternalServerError, gin.H{"error": "No authorization token header"})
			gCtx.Abort()
			return
		}

		// Validate the JWT token
		claims, err := token.ValidateToken(clientToken)
		if err != "" {
			gCtx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			gCtx.Abort()
			return
		}

		// Set the user's email and ID in the context
		gCtx.Set("email", claims.Email)
		gCtx.Set("uid", claims.Uid)

		// Continue processing the request
		gCtx.Next()
	}
}
