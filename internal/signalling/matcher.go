package signalling

import (
	"fmt"
	"github/english-app/internal/types"
	"log"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type         string     `json:"type"`
	User         types.User `json:"user,omitempty"`
	Target       int64      `json:"target,omitempty"`
	Payload      any        `json:"payload,omitempty"`
	From         int64      `json:"from,omitempty"`
	FromUserData any        `json:"fromUserData,omitempty"`
}

var clients = make(map[int64]*websocket.Conn)         // All clients connections
var availableClientsData = make(map[int64]types.User) // Online ready for CALL
var inCallClientsData = make(map[int64]types.User)    // Online Incall users

func handleClient(conn *websocket.Conn) {
	var userID int64

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			fmt.Println("invalid request : ", err)
			break // Handle other errors (e.g., log them)
		}

		switch msg.Type {
		case "initialize":
			userID = msg.User.Id
			clients[userID] = conn
			availableClientsData[userID] = msg.User

			defer func(uid int64) {
				delete(clients, uid)
				delete(availableClientsData, uid)
				if err := conn.Close(); err != nil {
					fmt.Println("connection close failed : ", err)
					// Handle error (e.g., log it)s
				}
			}(userID)
			continue

		case "offer", "answer", "icecandidate":
			targetConn, ok := clients[msg.Target]

			if !ok {
				log.Printf("❌ Target user %d not connected", msg.Target)
				continue
			}
			err := targetConn.WriteJSON(Message{
				Type:    msg.Type,
				From:    msg.From,
				Payload: msg.Payload,
				FromUserData: map[string]any{
					"id":          availableClientsData[msg.From].Id,
					"full_name":   availableClientsData[msg.From].FullName,
					"username":    availableClientsData[msg.From].Username,
					"profile_pic": availableClientsData[msg.From].ProfilePic,
					"gender":      availableClientsData[msg.From].Gender,
					"age":         availableClientsData[msg.From].Age,
				},
			})
			if err != nil {
				log.Printf("🚨 Error writing to WebSocket: %v", err)
			}

		case "randomCall":

		default:
			continue
		}
	}
}

// func matcher() {

// }

func findAUser(id int64) int64 {
	return id
}
