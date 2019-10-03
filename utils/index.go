package utils

import (
	"os"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/dgrijalva/jwt-go"
)

// GenerateAuthToken - Generate the Auth token for given id
func GenerateAuthToken(id string) (interface{}, error) {
	/*
		Method for generating the token
	*/
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"_id": id,
		// "exp":   time.Now().Add(time.Hour * 2).Unix(),
	})

	authToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		return nil, err
	}

	return authToken, nil
}

// StringInSlice - Checking if the string in the slice
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func DeleteAnElementFromArrayObjectID(element primitive.ObjectID, array []primitive.ObjectID) []primitive.ObjectID {
	var idx = -1

	for i, e := range array {
		if e == element {
			idx = i
			break
		}
	}

	if idx == -1 {
		return array
	}

	return append(array[:idx], array[idx+1:]...)

}

func GenerateChatNameForTwoUsers(idA string, idB string) string {
	if idA > idB {
		return idA + "&" + idB
	}
	return idB + "&" + idA
}
