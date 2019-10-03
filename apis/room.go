package apis

import (
	"jumpInto/models"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gin-gonic/gin"
)

// Create a delete a room here

func GetRoomByID(c *gin.Context) {
	rid := c.Param("rid")
	room, err := models.FindRoomByID(rid)
	if err != nil {
		raw, err := c.GetRawData()
		c.JSON(http.StatusBadRequest, gin.H{
			"err":     err,
			"msg":     "Cannot find the room",
			"ReqBody": raw,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"room": room,
	})
}

func GetRooms(c *gin.Context) {
	// Add query

	rooms, err := models.FindRooms(bson.M{})

	if err != nil {
		raw, err := c.GetRawData()
		c.JSON(http.StatusBadRequest, gin.H{
			"err":     err,
			"msg":     "Cannot find the room",
			"ReqBody": raw,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
	})

}

func AddRoom(c *gin.Context) {

	var room *models.Room
	err := c.ShouldBindJSON(room)
	if err != nil {
		raw, err := c.GetRawData()
		c.JSON(http.StatusBadRequest, gin.H{
			"err":     err,
			"msg":     "Cannot bind the input data",
			"ReqBody": raw,
		})
		return
	}

	InsertedID, err := models.AddRoom(room)
	if err != nil {
		raw, err := c.GetRawData()
		c.JSON(http.StatusBadRequest, gin.H{
			"err":     err,
			"msg":     "Cannot Add this room",
			"ReqBody": raw,
		})
		return
	}

	room.ID = InsertedID.(primitive.ObjectID)

	c.JSON(http.StatusOK, gin.H{
		"room": room,
	})
}

func DeleteRoom(c *gin.Context) {
	rid := c.Param("rid")
	err := models.DeleteRoom(rid)
	if err != nil {
		raw, err := c.GetRawData()
		c.JSON(http.StatusBadRequest, gin.H{
			"err":     err,
			"msg":     "Cannot Delete this room",
			"ReqBody": raw,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"deleted": rid,
	})
}

func GetMembersOfRoom(c *gin.Context) {
	var room *models.Room
	var err error

	rid := c.Param("rid")

	room, err = models.FindRoomByID(rid)

	if err != nil {
		raw, err := c.GetRawData()
		c.JSON(http.StatusBadRequest, gin.H{
			"err":     err,
			"msg":     "Cannot Get this room",
			"ReqBody": raw,
			"rid":     rid,
		})
		return
	}

	members, err := models.FindClients(bson.M{"_id": bson.M{"$in": room.Members}})

	c.JSON(http.StatusOK, gin.H{
		"members": members,
		"rid":     rid,
	})
}
