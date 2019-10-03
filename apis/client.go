package apis

import (
	"jumpInto/models"
	"jumpInto/utils"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// SignupClient - Singup the client by "email", "password" and  "code"
func SignupClient(c *gin.Context) {
	// Extract required fields, including "email", "password" and "code"
	email, password, code := c.PostForm("email"), c.PostForm("password"), c.PostForm("code")

	log.Println("In the singup client")

	// Checking if the password or password is empty
	if email == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    "You must provide email and password",
			"msg":      "You must provide email and password",
			"email":    email,
			"password": password,
		})
		return
	}

	// Checking if the register code is correct
	// if code != os.Getenv("REGISTER_CODE") && code != os.Getenv("ADMIN_REGISTER_CODE") {
	// 	c.JSON(http.StatusBadRequest, gin.H{
	// 		"err":  "The Given Rigister Code is not correct",
	// 		"msg":  "The Given Rigister Code is not correct",
	// 		"code": code,
	// 	})
	// 	return
	// }
	// Checking if the client already exist
	if client, err := models.FindClientByEmail(email); client != nil {
		c.JSON(http.StatusConflict, gin.H{
			"err":    err,
			"msg":    "The client already exists.",
			"client": client,
		})
		return
	}

	// Using the register code to define the role
	role := "normal"
	if code == os.Getenv("ADMIN_REGISTER_CODE") {
		role = "admin"
	}

	// Hash the given password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": err,
			"msg": "Cannot hash the given password",
		})
		return
	}

	// Create the newClient
	newClient := models.Client{
		Email:    email,
		Password: string(hashedPassword),
		Role:     role,
		Name:     "",
		Rooms:    []primitive.ObjectID{},
	}

	// Add this Client
	insertedID, err := models.AddClient(&newClient)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     err,
			"msg":       "Cannot register this client",
			"newClient": newClient,
		})
		return
	}

	newClient.ID = insertedID.(primitive.ObjectID)

	newClient.Password = ""

	authToken, err := utils.GenerateAuthToken(newClient.ID.Hex())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err,
			"msg":    "Cannot generate the auth token for this client",
			"client": newClient,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"client": newClient,
		"token":  authToken,
	})

}

// LoginClient - Login the client through email and password
func LoginClient(c *gin.Context) {
	inputEmail, inputPassword := c.PostForm("email"), c.PostForm("password")

	if inputEmail == "" || inputPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Must Provide Email and Password",
			"msg":        "Must Provide Email and Password",
			"inputEmail": inputEmail,
		})
		return
	}

	client, err := models.CheckingTheAuth(inputEmail, inputPassword)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      err,
			"msg":        "Email or Password is not correct",
			"client":     client,
			"inputEmail": inputEmail,
		})
		return
	}

	authToken, err := utils.GenerateAuthToken(client.ID.Hex())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      err,
			"msg":        "Cannot Generate the Auth Token",
			"inputEmail": inputEmail,
			"client":     client,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":  authToken,
		"client": client,
	})

}

func GetTheRoomsOfClient(c *gin.Context) {

	var client *models.Client
	var err error

	cid := c.Param("cid")

	if cid == "" {
		clientStr, ok := c.Get("client")
		if !ok {
			raw, err := c.GetRawData()
			c.JSON(http.StatusBadRequest, gin.H{
				"err":     err,
				"msg":     "Cannot find the room",
				"cid":     cid,
				"ReqBody": raw,
			})
			return
		}

		client = clientStr.(*models.Client)
		cid = client.ID.String()
	} else {
		client, err = models.FindClientByID(cid)
		if err != nil {
			raw, err := c.GetRawData()
			c.JSON(http.StatusBadRequest, gin.H{
				"err":     err,
				"msg":     "Cannot find rooms for given user",
				"cid":     cid,
				"ReqBody": raw,
			})
			return
		}

	}

	rooms, err := models.FindRooms(bson.M{"_id": bson.M{"$in": client.Rooms}})
	if err != nil {
		raw, err := c.GetRawData()
		c.JSON(http.StatusBadRequest, gin.H{
			"err":     err,
			"msg":     "Cannot find rooms for given user",
			"cid":     cid,
			"ReqBody": raw,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
		"cid":   cid,
	})
}

func GetClientByID(c *gin.Context) {
	cid := c.Param("cid")
	client, err := models.FindClientByID(cid)
	if err != nil {
		raw, err := c.GetRawData()
		c.JSON(http.StatusBadRequest, gin.H{
			"err":     err,
			"msg":     "Cannot find rooms for given user",
			"cid":     cid,
			"ReqBody": raw,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client": client,
	})

}
