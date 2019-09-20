package routers

import (
	"jumpInto/middlewares"
	"jumpInto/models"
	"log"

	"github.com/gin-gonic/gin"
)

// InitRouter - Initialise all the router in this function
func InitRouter() *gin.Engine {
	router := gin.Default()
	UserRouterInit(router)

	router.GET("/ws", middlewares.LoginAuth(), func(c *gin.Context) {

		client := c.Get("client").(*models.Client)
		println("Connection started")
		err := models.SocketHandler(c.Writer, c.Request)

		if err != nil {
			log.Println("Error:", err)
		}
		println("Connetion is settled")

	})

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, []string{"123", "321"})
	})

	return router
}
