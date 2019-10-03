package main

import (
	"jumpInto/database"
	"jumpInto/routers"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func main() {

	// bindAfdress := "localhost:8000"

	gin.ForceConsoleColor()
	database.InitDB()
	r := routers.InitRouter()
	r.Run()

	// Testting Area

	// newMsg := models.Message{
	// 	Action:        "SEND_MESSAGE",
	// 	Destination:   "DESTINATION_ID",
	// 	Source:        "SOURCE_ID",
	// 	PayLoad:       "This is the payload",
	// 	SendiningTime: time.Now(),
	// 	To:            "ROOM",
	// }

	// models.AddMessageToDB(&newMsg)

	// database.DB.Collection("chatHist").InsertOne(context.TODO(), bson.M{"test": "testing"})

}
