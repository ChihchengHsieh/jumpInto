package models

import (
	"context"
	"jumpInto/database"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Room struct {
	ID        primitive.ObjectID   `json:"_id" bson:"_id"`
	Name      string               `json:"name" bson:"name"`       // Name of the room
	Members   []primitive.ObjectID `json:"members" bson:"members"` // Members of client in this room
	StopChan  chan bool            `json:"-" bson:"-"`             // Meaning of this?
	JoinChan  chan *Client         `json:"-" bson:"-"`
	LeaveChan chan *Client         `json:"-" bson:"-"`
	Send      chan *Message        `json:"-" bson:"-"`
}

var RoomManager = make(map[string]*Room)

func (r *Room) Stop() {
	r.StopChan <- true
}

// Adds a Client to the Room.
func (r *Room) Join(c *Client) {
	r.JoinChan <- c
}

// Removes a Client from the Room.
func (r *Room) Leave(c *Client) {
	r.LeaveChan <- c
}

// Broadcasts data to all members of the Room.
func (r *Room) Emit(m *Message) {
	r.Send <- m
}

// Creates a new Room type and starts it.
func NewRoom(name string) *Room {

	// Check if we can find a room in database

	r := &Room{
		Name:      name,
		Members:   []primitive.ObjectID{},
		StopChan:  make(chan bool),
		JoinChan:  make(chan *Client),
		LeaveChan: make(chan *Client),
		Send:      make(chan *Message),
	}

	// Check it from outside
	// if oldR, ok := RoomManager[name]; ok {
	// 	r.Members = oldR.Members // Keep the members data
	// }

	InsertedID, err := AddRoom(r)
	if err != nil {
		log.Println(err)
	}
	r.ID = InsertedID.(primitive.ObjectID)
	RoomManager[name] = r
	go r.Start()
	return r
}

// AddRoom -
func AddRoom(newRoom *Room) (interface{}, error) {
	// Accepting the room but only save as the Members
	result, err := database.DB.Collection("room").InsertOne(context.TODO(), newRoom)
	return result.InsertedID, err
}

func DeleteRoom(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = database.DB.Collection("room").DeleteOne(context.TODO(), bson.M{"_id": oid})
	if err != nil {
		return err
	}

	return nil
}

func UpdateRoomByID(id string, updateDetail map[string]interface{}) (interface{}, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	result, err := database.DB.Collection("room").UpdateOne(context.TODO(),
		bson.M{"_id": oid},
		bson.M{"$set": updateDetail})

	if err != nil {
		return nil, err
	}

	return result.UpsertedID, nil

}

func FindRoomByID(id string) (*Room, error) {
	var room *Room
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}
	err = database.DB.Collection("room").FindOne(context.TODO(),
		bson.M{"_id": oid},
		options.FindOne().SetProjection(projectionForRemovingPassword)).Decode(&room)
	if err != nil {
		return nil, err
	}
	room.StopChan = make(chan bool)
	room.JoinChan = make(chan *Client)
	room.LeaveChan = make(chan *Client)
	room.Send = make(chan *Message)

	return room, nil
}

func FindRooms(filterDetail bson.M) ([]*Room, error) {
	var rooms []*Room
	result, err := database.DB.Collection("room").Find(context.TODO(),
		filterDetail,
		options.Find().SetProjection(projectionForRemovingPassword))

	if err != nil {
		return nil, err
	}
	defer result.Close(context.TODO())

	for result.Next(context.TODO()) {
		var elem Room
		err := result.Decode(&elem)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, &elem)
	}
	return rooms, nil
}

func AddClientToRoomByID(id string, inputClient *Client) (interface{}, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("transfering ID problem")
		return nil, err
	}
	result, err := database.DB.Collection("room").UpdateOne(context.TODO(),
		bson.M{"_id": oid},
		bson.M{"$push": bson.M{"members": inputClient}})
	if err != nil {
		log.Printf("fail to add the the room")
		return nil, err
	}
	return result.UpsertedID, nil
}

func UpdateClientToRoomByID(roomID string, clientID string, newClient *Client) (interface{}, error) {
	roomOid, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		return nil, err
	}

	clientOid, err := primitive.ObjectIDFromHex(clientID)
	if err != nil {
		return nil, err
	}

	result, err := database.DB.Collection("room").UpdateOne(context.TODO(),
		bson.M{"_id": roomOid, "members._id": clientOid},
		bson.M{"$set": bson.M{"members.$": newClient}})
	// bson.M{"$set": bson.M{"clients.$": new}})
	if err != nil {
		return nil, err
	}

	return result.UpsertedID, nil

}

func DeleteClientFromMemberByID(roomID string, clientID string) (interface{}, error) {
	roomOID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		return nil, err
	}

	clientOid, err := primitive.ObjectIDFromHex(clientID)
	if err != nil {
		return nil, err
	}

	result, err := database.DB.Collection("room").UpdateOne(context.TODO(),
		bson.M{"_id": roomOID},
		bson.M{"$pull": bson.M{"members": bson.M{"_id": clientOid}}})

	if err != nil {
		return nil, err
	}

	return result.UpsertedID, nil
}

// Delete room
// Delete room from a room

func (r *Room) Start() {
	for {
		select {
		case c := <-r.JoinChan:

			c.Send <- &Message{
				Action:        "JOINED",
				Destination:   r.Name,
				PayLoad:       "You have joined the room",
				SendiningTime: time.Now(),
				Source:        "System",
				To:            "CLIENT",
			}
		case c := <-r.LeaveChan:

			c.Send <- &Message{
				Action:        "LEFT",
				Destination:   r.Name,
				PayLoad:       "You have left the room",
				SendiningTime: time.Now(),
				Source:        "System",
				To:            "CLIENT",
			}

		case msg := <-r.Send:
			// we store the message to server
			for _, cid := range r.Members {
				if cid.String() == msg.Source {
					continue
				}

				// If the client is on the server, then we use websocket to pass the meesage
				if client, ok := ClientManager[cid.String()]; ok {
					client.Send <- msg
				}

			}

		case <-r.StopChan:
			delete(RoomManager, r.Name)
			return
		}
	}
}

// Create the Room Route in here

// Should be able to get the room manager from here

/* Saving Content:
1. Members in the room
2. ChatHist of this room
*/
