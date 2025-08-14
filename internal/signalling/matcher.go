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
var allClientsData = make(map[int64]types.User)        // just for getting data of users

func handleClient(conn *websocket.Conn, db storage.Storage) {
	var userID int64
	var mutex = &sync.Mutex{}

	// defer func(uid int64) {
	// 	fmt.Println("closing connection for user: ", uid)
	// 	delete(clients, uid)
	// 	delete(availableClientsData, uid)
	// 	delete(inCallClientsData, uid)
	// 	delete(waitingForCallClients, uid)
	// 	delete(allClientsData, uid)
	// 	if err := conn.Close(); err != nil {
	// 		fmt.Println("connection close failed : ", err)
	// 		// Handle error (e.g., log it)s
	// 	}
	// }(userID)

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			fmt.Println("invalid request : ", err)
			break // Handle other errors (e.g., log them)
		}

		conn.WriteJSON(map[string]string{
			"data": "connected to websocket",
		})

		switch msg.Type {
		case "initialize":
			userID = msg.User.Id
			clients[userID] = conn
			availableClientsData[userID] = msg.User
			allClientsData[userID] = msg.User

			usersList, err := ShowRelatedUsersList(userID)
			if err != nil {
				fmt.Println("err in show related user list ", err.Error())
			}
			err = conn.WriteJSON(map[string]any{
				"type":         "initialized",
				"user":         msg.User,
				"online_users": usersList, // optional function
				"message":      "connected successfully",
			})
			fmt.Println("running")
			if err != nil {
				fmt.Print("err in initializaion phase : ", err)
			}

		case "offer", "icecandidate":
			targetConn, ok := clients[msg.Target]
			fmt.Println("offer received : ", msg)
			if !ok {
				log.Printf("❌ Target user %d not connected", msg.Target)
				continue
			}
			err := targetConn.WriteJSON(Message{
				Type:    msg.Type,
				From:    msg.From,
				Payload: msg.Payload,
				FromUserData: map[string]any{
					"id":          allClientsData[msg.From].Id,
					"full_name":   allClientsData[msg.From].FullName,
					"username":    allClientsData[msg.From].Username,
					"profile_pic": allClientsData[msg.From].ProfilePic,
					"gender":      allClientsData[msg.From].Gender,
					"age":         allClientsData[msg.From].Age,
				},
				Target: msg.Target,
			})
			if err != nil {
				log.Printf("🚨 Error writing to WebSocket: %v", err)
			}

		case "answer":
			inCallClientsData[msg.From] = availableClientsData[msg.From]
			inCallClientsData[msg.Target] = availableClientsData[msg.Target]

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
					"id":          allClientsData[msg.From].Id,
					"full_name":   allClientsData[msg.From].FullName,
					"username":    allClientsData[msg.From].Username,
					"profile_pic": allClientsData[msg.From].ProfilePic,
					"gender":      allClientsData[msg.From].Gender,
					"age":         allClientsData[msg.From].Age,
				},
				Target: msg.Target,
			})

			delete(availableClientsData, msg.From)
			delete(availableClientsData, msg.Target)
			delete(waitingForCallClients, msg.From)
			delete(waitingForCallClients, msg.Target)

			// add Start call details to Database
			go func() {
				mutex.Lock()
				defer mutex.Unlock()
				callId, err := db.StartCall(msg.Target, msg.From)
				if err != nil {
					fmt.Println("error in start call ", err)
				}
				clients[msg.From].WriteJSON(map[string]string{
					"type":   "callStarted",
					"callId": callId,
				})
				clients[msg.Target].WriteJSON(map[string]string{
					"type":   "callStarted",
					"callId": callId,
				})
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
				Type:    "endCall",
				From:    msg.From,
				Payload: msg.Payload,
				FromUserData: map[string]any{
					"id":          allClientsData[msg.From].Id,
					"full_name":   allClientsData[msg.From].FullName,
					"username":    allClientsData[msg.From].Username,
					"profile_pic": allClientsData[msg.From].ProfilePic,
					"gender":      allClientsData[msg.From].Gender,
					"age":         allClientsData[msg.From].Age,
				},
			})
			//Update in database Call Ended
			// go func() {

			// }()

		case "randomCall":
			From := msg.From
			waitingForCallClients[From] = availableClientsData[From]
			peerID, err := findAUser(From)
			if err != nil {
				log.Println("Error in Random Call Func:", err.Error())
				continue
			}
			// in frontend we will check if randomcallOffer received and user is in random match then  we will immediately send answer in response
			// and other peer will receive answer and we will check if user is in random match then we will immeditely process to webrtc connection
			clients[peerID].WriteJSON(Message{
				Type:    "randomCallOffer",
				From:    msg.From,
				Payload: msg.Payload,
				FromUserData: map[string]any{
					"id":          allClientsData[msg.From].Id,
					"full_name":   allClientsData[msg.From].FullName,
					"username":    allClientsData[msg.From].Username,
					"profile_pic": allClientsData[msg.From].ProfilePic,
					"gender":      allClientsData[msg.From].Gender,
					"age":         allClientsData[msg.From].Age,
				},
			})
			delete(availableClientsData, From)
		case "cancelRandomMatch":
			availableClientsData[msg.From] = waitingForCallClients[msg.From]
			delete(waitingForCallClients, msg.From)
			clients[msg.From].WriteJSON(map[string]any{
				"type": "canceledRandomMatch",
			})
		case "rejectCall":
			delete(waitingForCallClients, msg.From)
			delete(waitingForCallClients, msg.Target)

			Target := msg.Target
			clients[Target].WriteJSON(Message{
				Type: "rejectedCall",
				From: msg.From,
				FromUserData: map[string]any{
					"id":          allClientsData[msg.From].Id,
					"full_name":   allClientsData[msg.From].FullName,
					"username":    allClientsData[msg.From].Username,
					"profile_pic": allClientsData[msg.From].ProfilePic,
					"gender":      allClientsData[msg.From].Gender,
					"age":         allClientsData[msg.From].Age,
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

func ShowRelatedUsersList(id int64) ([]types.User, error) {
	// returning array of users maps
	var msg []types.User
	fmt.Println("available clients data : ", availableClientsData)
	if len(availableClientsData) < 1 {
		return msg, fmt.Errorf("Users are not Online at this time")
	}

	for _, data := range availableClientsData {
		if len(msg) < 30 && data.Id != id {
			// fmt.Println("user: ",data)
			msg = append(msg, data)
		} else {
			return msg, nil
		}

	}
	return msg, nil
}
