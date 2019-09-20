package routers

import (
	"blog/apis"

	"github.com/gin-gonic/gin"
)

// UserRouterInit - Initialise userRouter
func UserRouterInit(router *gin.Engine) {
	userRouter := router.Group("/user")
	{
		userRouter.POST("/signup", apis.SignupUser)
		userRouter.POST("/login", apis.LoginUser)
	}
}
