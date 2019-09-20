package utils

import (
	"os"

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

func GenerateChatNameForTwoUsers(idA string, idB string) string {
	if idA > idB {
		return idA + "&" + idB
	}
	return idB + "&" + idA
}
