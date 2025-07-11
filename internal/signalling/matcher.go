package signalling

import (
	"errors"
	"fmt"
	"github/english-app/internal/types"
	"github/english-app/storage"
	"log"
	"sync"

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

var clients = make(map[int64]*websocket.Conn)          // All clients connections
var availableClientsData = make(map[int64]types.User)  // Online ready for CALL
var inCallClientsData = make(map[int64]types.User)     // Online Incall users
var waitingForCallClients = make(map[int64]types.User) // Waiting For Joinin

func handleClient(conn *websocket.Conn, storage *storage.Storage) {
	var userID int64
	var mutex *sync.Mutex
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
			clients[userID].WriteMessage("connected succenssfuly")
			continue

		case "offer", "icecandidate":
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

		case "answer":
			inCallClientsData[msg.From] = availableClientsData[msg.From]
			inCallClientsData[msg.Target] = availableClientsData[msg.Target]

			delete(availableClientsData, msg.From)
			delete(availableClientsData, msg.Target)

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

			// add Start call details to Database
			go func() {
				mutex.Lock()

				mutex.Unlock()
			}()
			if err != nil {
				log.Printf("🚨 Error writing to WebSocket: %v", err)
			}

		case "endCall":
			availableClientsData[msg.From] = inCallClientsData[msg.From]
			availableClientsData[msg.Target] = inCallClientsData[msg.Target]

			delete(inCallClientsData, msg.From)
			delete(inCallClientsData, msg.Target)

			Target := clients[msg.Target]
			Target.WriteJSON(Message{
				Type:    "offer",
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
			//Update in database Call Ending
			go func() {

			}()

		case "randomCall":
			From := msg.From
			waitingForCallClients[From] = availableClientsData[From]
			peerID, err := findAUser(From)
			if err != nil {
				fmt.Errorf("Error in Random Call Func ", err.Error())
				continue
			}
			clients[peerID].WriteJSON(Message{
				Type:    "offer",
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
			delete(availableClientsData, From)

		case "rejectCall":
			Target := msg.Target
			clients[Target].WriteJSON(Message{
				Type: "rejectCall",
				From: msg.From,
				FromUserData: map[string]any{
					"id":          availableClientsData[msg.From].Id,
					"full_name":   availableClientsData[msg.From].FullName,
					"username":    availableClientsData[msg.From].Username,
					"profile_pic": availableClientsData[msg.From].ProfilePic,
					"gender":      availableClientsData[msg.From].Gender,
					"age":         availableClientsData[msg.From].Age,
				},
			})

		default:
			continue
		}
	}
}

func findAUser(id int64) (int64, error) {
	for peer := range waitingForCallClients {
		if id != peer {
			return peer, nil
		}
	}
	return 0, errors.New("did not find any user")
}

func ShowRelatedUsersList() ([]types.User, error) {
	// returning array of users maps
	var msg []types.User

	if len(availableClientsData) < 1 {
		return msg, fmt.Errorf("Users are not Online at this time")
	}

	for _, data := range availableClientsData {
		if len(msg) < 30 {
			msg = append(msg, data)
		} else {
			break
		}

	}
	return msg, nil
}
