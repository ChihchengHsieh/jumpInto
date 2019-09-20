package middlewares

import (
	"fmt"
	"jumpInto/models"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// LoginAuth - Login Middleware for checking the auth
func LoginAuth() gin.HandlerFunc {
	return func(c *gin.Context) {

		tokenStr := c.GetHeader("Authorization")

		// log.Printf("tokenStr: %+v", tokenStr)

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "No token Provided",
				"msg":   "Token is needed",
				"token": tokenStr,
			})
			return
		}

		// Check if it use Bearer

		if s := strings.Split(tokenStr, " "); len(s) == 2 {
			tokenStr = s[1]
		}

		// Problem: token is generatted but not valid

		// Add Claim Latter
		token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error":    "Invalid Token",
					"msg":      "Cannot parse the given token",
					"token":    token,
					"tokenStr": tokenStr,
				})
				return nil, fmt.Errorf("Invalid Token")
			}

			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		// fmt.Printf("The token:\n %+v\n", token)

		fmt.Printf("Claims:\n%+v\n", token.Claims.(jwt.MapClaims))

		// Find the Client and store in c

		fmt.Printf("Point1")

		claims := token.Claims.(jwt.MapClaims)

		fmt.Printf("Point2")

		inputClaim := claims["_id"].(string)

		fmt.Printf("PointSp")

		client, err := models.FindClientByID(inputClaim)

		fmt.Printf("Point3")

		if err != nil {
			log.Print(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"err":        err,
				"msg":        "Cannot find this client",
				"inputClaim": inputClaim,
				"claims":     claims,
				"token":      token,
				"tokenStr":   tokenStr,
			})
			return
		}

		c.Set("client", client)

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"tokenStr": tokenStr,
				"token":    token,
				"error":    "Token is not valid",
				"msg":      "The token is not valid",
			})
			return
		}

		c.Next()
	}
}

// SoftAuth - only checking the token but will not abbort
func SoftAuth() gin.HandlerFunc {
	return func(c *gin.Context) {

		tokenStr := c.GetHeader("Authorization")

		if strings.TrimSpace(tokenStr) == "" {
			c.Set("client", nil)
			c.Next()
			return
		}

		// log.Printf("tokenStr: %+v", tokenStr)

		if s := strings.Split(tokenStr, " "); len(s) == 2 {
			tokenStr = s[1]
		}

		// Problem: token is generatted but not valid

		// Add Claim Latter
		token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error":    "Invalid Token",
					"msg":      "Cannot parse the given token",
					"token":    token,
					"tokenStr": tokenStr,
				})
				return nil, fmt.Errorf("Invalid Token")
			}

			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		// fmt.Printf("The token:\n %+v\n", token)

		fmt.Printf("Claims:\n%+v\n", token.Claims.(jwt.MapClaims))

		// Find the Client and store in c

		fmt.Printf("Point1")

		claims := token.Claims.(jwt.MapClaims)

		fmt.Printf("Point2")

		inputClaim := claims["_id"].(string)

		fmt.Printf("PointSp")

		client, err := models.FindClientByID(inputClaim)

		fmt.Printf("Point3")

		if err != nil {
			log.Print(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"err":        err,
				"msg":        "Cannot find this client",
				"inputClaim": inputClaim,
				"claims":     claims,
				"token":      token,
				"tokenStr":   tokenStr,
			})
			return
		}

		c.Set("client", client)

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"tokenStr": tokenStr,
				"token":    token,
				"error":    "Token is not valid",
				"msg":      "The token is not valid",
			})
			return
		}

		c.Next()
	}
}
