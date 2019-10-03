package routers

import (
	"jumpInto/middlewares"
	"jumpInto/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func WebSocketInit(router *gin.Engine) {
	router.GET("/ws", middlewares.LoginAuth(), func(c *gin.Context) {

		client, ok := c.Get("client")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"err": "Cannot get the client",
				"msg": "Cannot get the client",
			})
		}

		println("Connection started")
		err := models.SocketHandler(c.Writer, c.Request, client.(*models.Client))

		if err != nil {
			log.Println("Error:", err)
		}
		println("Connetion is settled")

	})

}
