package routers

import (
	"jumpInto/apis"

	"github.com/gin-gonic/gin"
)

func RoomRouterInit(router *gin.Engine) {
	roomRouter := router.Group("/room")
	{
		roomRouter.GET("/", apis.GetRooms)

		roomRouter.POST("/", apis.AddRoom)
		roomRouter.DELETE("/single/:rid", apis.DeleteRoom)
		roomRouter.GET("/single/:rid", apis.GetRoomByID)
		roomRouter.GET("/single/:rid/members", apis.GetMembersOfRoom)

	}
}
