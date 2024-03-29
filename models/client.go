package models

import (
	"context"
	"errors"
	"jumpInto/database"
	"jumpInto/utils"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline                       = []byte{'\n'}
	space                         = []byte{' '}
	projectionForRemovingPassword = bson.D{
		{"password", 0},
	}
)

type Client struct {
	ID       primitive.ObjectID   `json:"_id" bson:"_id,omitempty"`
	Email    string               `json:"email" bson:"email"`
	Password string               `json:"password" bson:"password"`
	Role     string               `json:"role" bson:"role"`
	Socket   *websocket.Conn      `json:"-" bson:"-"`
	Name     string               `json:"name" bson:"name"`
	Send     chan *Message        `json:"-" bson:"-"`
	Rooms    []primitive.ObjectID `json:"rooms" bson:"rooms"` //The Room that the client living in
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var (
	// Stores all Conn types by their usrId,
	ClientManager = make(map[string]*Client)
)

func HandleData(c *Client, msg *Message) {
	switch msg.Action { // What's the meaning for this event?
	case "JOIN": // action
		c.Join(msg.Destination)
	case "LEAVE": // action
		c.Leave(msg.Destination)
	case "JOINED": // send message to notify that some one has joined
		c.Emit(msg)
	case "LEFT": // send message to notify that someone has left
		c.Emit(msg)

		room := RoomManager[msg.Destination]
		if len(room.Members) == 0 {
			room.Stop()
		}

	case "SENDING_MESSAGE":
		switch msg.To {
		case "ROOM":
			if dst, ok := RoomManager[msg.Destination]; ok {
				dst.Send <- msg
			} else {
				room, err := FindRoomByID(msg.Destination)
				if err != nil {
					log.Println(err)
				}
				RoomManager[room.ID.String()] = room
				err = AddMessageToDB(msg)
				if err != nil {
					log.Println(err)
				}
				room.Send <- msg
			}
		case "CLIENT":
			if dst, ok := ClientManager[msg.Destination]; ok {
				dst.Send <- msg
			} else {
				client, err := FindClientByID(msg.Destination)
				if err != nil {
					log.Println(err)
				}
				ClientManager[client.ID.String()] = client
				err = AddMessageToDB(msg)
				if err != nil {
					log.Println(err)
				}

				// In doesn't have websocket in this case
			}

		default:
			c.Send <- &Message{
				Source:        "Sysytem",
				Destination:   c.ID.String(),
				Action:        "Error",
				PayLoad:       "Cannot Send Out This Message",
				SendiningTime: time.Now(),
			}
		}

	default:
		c.Send <- &Message{
			Source:        "Sysytem",
			Destination:   c.ID.String(),
			Action:        "Error",
			PayLoad:       "Cannot Send Out This Message",
			SendiningTime: time.Now(),
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		for _, roomID := range c.Rooms {
			RoomManager[roomID.String()].Leave(c)
		}
		c.Socket.Close()
	}()
	c.Socket.SetReadLimit(maxMessageSize)
	c.Socket.SetReadDeadline(time.Now().Add(pongWait))
	c.Socket.SetPongHandler(func(string) error {
		c.Socket.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for { // keep reading the message , and keep handling the data
		var msg Message
		err := c.Socket.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %+v", err)
			}
			break
		}
		HandleData(c, &msg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Socket.Close()
	}()
	for {
		select {
		case msg, ok := <-c.Send:
			c.Socket.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Socket.WriteJSON(msg); err != nil {
				log.Printf("Error in Writing: %+v", err)
				c.Socket.WriteJSON(Message{
					Action:        "ERROR",
					Destination:   c.ID.String(),
					PayLoad:       "Error:" + err.Error(),
					SendiningTime: time.Now(),
					Source:        "System",
				})
			}

		case <-ticker.C:
			if err := c.Socket.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Adds the Conn to a Room. If the Room does not exist, it is created.
func (c *Client) Join(id string) {
	var room *Room
	var err error

	// Checking if we can find the existing room
	if _, ok := RoomManager[id]; ok {
		// Check with the DB as well
		room = RoomManager[id]
	} else {
		// Create a room if the room doesn't exist
		room, err = FindRoomByID(id)
		if err != nil {
			log.Println(err)
		}
		// room = NewRoom(id)
	}

	////////////////// Update the DB
	_, err = AddClientToRoomByID(room.ID.String(), c)

	if err != nil {
		log.Println(err) // We can send error throguh websocket
	}

	_, err = AddRoomToClientByID(c.ID.String(), room)

	if err != nil {
		log.Println(err) // We can send error throguh websocket
	}

	// Update Current Server
	room.Members = append(room.Members, c.ID)
	c.Rooms = append(c.Rooms, room.ID)
	RoomManager[id] = room

	room.Join(c) // bi-directional binding
}

// Removes the Conn from a Room.
func (c *Client) Leave(id string) {

	var room *Room
	var ok bool

	if room, ok = RoomManager[id]; ok {
		// Checking if we have this room in currentServer
		room = RoomManager[id]

	} else {

		// If we don't have this room, we retrieve from db
		room, err := FindRoomByID(id)
		if err != nil {
			log.Println(err)
		}
		RoomManager[id] = room
	}

	// Removing from the current Server
	room.Members = utils.DeleteAnElementFromArrayObjectID(c.ID, room.Members)
	c.Rooms = utils.DeleteAnElementFromArrayObjectID(room.ID, c.Rooms)

	// Removing from the database
	_, err := DeleteClientFromMemberByID(room.ID.String(), c.ID.String())

	if err != nil {
		log.Println(err)
	}

	_, err = DeleteRoomFromClientByID(c.ID.String(), room.ID.String())
	if err != nil {
		log.Println(err)
	}

	room.Leave(c)
}

// Broadcasts a Message to all members of a Room.
func (c *Client) Emit(msg *Message) {
	// If we are having this Room in current server
	if room, ok := RoomManager[msg.Destination]; ok {
		room.Emit(msg)
	} else {
		// Else we retrieve it from the database
		room, err := FindRoomByID(msg.Destination)
		if err != nil {
			log.Println(err)
		}
		RoomManager[msg.Destination] = room
		room.Emit(msg)
	}
}

// Upgrades an HTTP connection and creates a new Conn type.
func NewConnection(w http.ResponseWriter, r *http.Request, client *Client) error {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	if err != nil {
		log.Println(err)
		return err
	}

	client.Socket = socket

	ClientManager[client.ID.String()] = client

	return nil
}

// Calls NewConnection, starts the returned Conn's writer, joins the root room, and finally starts the Conn's reader.
func SocketHandler(w http.ResponseWriter, r *http.Request, client *Client) error {

	if r.Method != "GET" {
		return errors.New("Method not allowed 405")
	}
	if err := NewConnection(w, r, client); err != nil {
		return err
	}

	go client.writePump()
	client.Join("root")
	go client.readPump()

	return nil

}

func AddRoomToClientByID(id string, inputRoom *Room) (interface{}, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		log.Println(err)
		return nil, err

	}

	result, err := database.DB.Collection("client").UpdateOne(context.TODO(),
		bson.M{"_id": oid},
		bson.M{"$push": bson.M{"rooms": inputRoom.ID}},
	)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return result.UpsertedID, nil
}

func DeleteRoomFromClientByID(clientID string, roomID string) (interface{}, error) {

	clientOID, err := primitive.ObjectIDFromHex(clientID)
	if err != nil {
		return nil, err
	}

	roomOID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		return nil, err
	}

	result, err := database.DB.Collection("client").UpdateOne(context.TODO(),
		bson.M{"_id": clientOID},
		bson.M{"$pull": bson.M{"rooms": roomOID}})

	if err != nil {
		return nil, err
	}

	return result.UpsertedID, nil
}

func AddClient(inputClient *Client) (interface{}, error) {
	result, err := database.DB.Collection("client").InsertOne(context.TODO(), inputClient)
	return result.InsertedID, err
}

func UpdateClientByID(id string, updateDetail map[string]interface{}) (interface{}, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	result, err := database.DB.Collection("client").UpdateOne(context.TODO(), bson.M{"_id": oid}, bson.M{"$set": updateDetail})

	if err != nil {
		return nil, err
	}

	return result.UpsertedID, nil
}

func FindClientByID(id string) (*Client, error) {
	var client *Client
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}
	err = database.DB.Collection("client").FindOne(context.TODO(),
		bson.M{"_id": oid},
		options.FindOne().SetProjection(projectionForRemovingPassword)).Decode(&client)
	if err != nil {
		return nil, err
	}

	client.Send = make(chan *Message)
	return client, nil
}

func FindClientByEmail(email string) (*Client, error) {
	var client *Client
	err := database.DB.Collection("client").FindOne(context.TODO(),
		bson.M{"email": email},
		options.FindOne().SetProjection(projectionForRemovingPassword)).Decode(&client)
	if err != nil {
		return nil, err
	}

	client.Send = make(chan *Message)
	return client, nil
}

func CheckingTheAuth(email string, password string) (*Client, error) {
	var client Client
	err := database.DB.Collection("client").FindOne(context.TODO(),
		bson.M{"email": email}).Decode(&client)

	err = bcrypt.CompareHashAndPassword([]byte(client.Password), []byte(password))

	if err != nil {
		return nil, err
	}
	return &client, nil
}

func FindClients(filterDetail bson.M) ([]*Client, error) {
	var clients []*Client
	result, err := database.DB.Collection("client").Find(context.TODO(),
		filterDetail,
		options.Find().SetProjection(projectionForRemovingPassword))

	if err != nil {
		return nil, err
	}
	defer result.Close(context.TODO())

	for result.Next(context.TODO()) {
		var elem Client
		err := result.Decode(&elem)
		if err != nil {
			return nil, err
		}
		elem.Send = make(chan *Message)
		clients = append(clients, &elem)
	}
	return clients, nil
}
