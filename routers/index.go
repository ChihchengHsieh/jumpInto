package routers

import (
	"github.com/gin-gonic/gin"
)

// InitRouter - Initialise all the router in this function
func InitRouter() *gin.Engine {
	router := gin.Default()
	UserRouterInit(router)
	WebSocketInit(router)
	RoomRouterInit(router)
	// router.GET("/ws", middlewares.LoginAuth(), func(c *gin.Context) {

	// 	client, ok := c.Get("client")
	// 	if !ok {
	// 		c.JSON(http.StatusBadRequest, gin.H{
	// 			"err": "Cannot get the client",
	// 			"msg": "Cannot get the client",
	// 		})
	// 	}

	// 	println("Connection started")
	// 	err := models.SocketHandler(c.Writer, c.Request, client.(*models.Client))

	// 	if err != nil {
	// 		log.Println("Error:", err)
	// 	}
	// 	println("Connetion is settled")

	// })

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, []string{"123", "321"})
	})

	return router
}
