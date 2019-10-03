package routers

import (
	"jumpInto/apis"

	"github.com/gin-gonic/gin"
)

// UserRouterInit - Initialise userRouter
func UserRouterInit(router *gin.Engine) {
	userRouter := router.Group("/user")
	{
		userRouter.POST("/signup", apis.SignupClient)
		userRouter.POST("/login", apis.LoginClient)
		userRouter.GET("/single/:cid/rooms", apis.GetClientByID)

	}
}
