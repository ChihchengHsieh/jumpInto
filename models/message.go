package models

import (
	"context"
	"jumpInto/database"
	"jumpInto/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
)

type Message struct {
	ID            primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Source        string             `json:"source" bson:"source"` // should be the id,
	Destination   string             `json:"destination" bson:"destination"`
	Action        string             `json:"action" bson:"action"`
	SendiningTime time.Time          `json:"sendingTime" bson:"sendingTime"`
	PayLoad       string             `json:"payload" bson:"payload"`
	To            string             `json:"to" bson:"to"`
}

func AddMessageToDB(msg *Message) error {
	//  should contain source and destination

	var chatHistName string
	switch msg.To {
	case "ROOM":
		// Check if the chat hist exist
		chatHistName = msg.Destination
	case "CLEINT":
		chatHistName = utils.GenerateChatNameForTwoUsers(msg.Source, msg.Destination)

	default:
		chatHistName = "UNKNOWN"
	}

	updateOption := options.Update()

	upsertOption := true

	updateOption.Upsert = &upsertOption

	_, err := database.DB.Collection("chatHist").UpdateOne(context.TODO(),
		bson.M{"name": chatHistName},
		bson.M{"$push": bson.M{"messages": msg}},
		updateOption,
	)

	if err != nil {
		return err
	}

	return nil

}

func DeleteMessageFromByID(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return err
	}

	_, err = database.DB.Collection("chatHist").DeleteOne(context.TODO(),
		bson.M{"_id": oid})

	if err != nil {

	}

	return nil
}

func DeleteMessageByName(name string) error {

	_, err := database.DB.Collection("chatHist").DeleteOne(context.TODO(), bson.M{"name": name})

	if err != nil {
		return err
	}
	return nil

}

func DeleteMessagesByClientName(name string) error {
	_, err := database.DB.Collection("chatHist").DeleteMany(context.TODO(),
		bson.M{"name": bson.M{"$regex": ".*" + name + ".*"}})

	if err != nil {
		return err
	}
	return nil
}
