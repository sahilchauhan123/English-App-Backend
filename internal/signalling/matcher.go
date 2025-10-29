// package signalling

// import (
// 	"errors"
// 	"fmt"
// 	"github/english-app/internal/types"
// 	"github/english-app/storage"
// 	"log"
// 	"sync"
// 	"time"

// 	"github.com/gorilla/websocket"
// )

// type Message struct {
// 	Type         string     `json:"type"`
// 	User         types.User `json:"user,omitempty"`
// 	Target       int64      `json:"target,omitempty"`
// 	Payload      any        `json:"payload,omitempty"`
// 	From         int64      `json:"from,omitempty"`
// 	FromUserData any        `json:"fromUserData,omitempty"`
// 	RandomCall   bool       `json:"randomCall,omitempty"`
// }

// var (
// 	clients               = make(map[int64]*websocket.Conn) // All clients connections
// 	availableClientsData  = make(map[int64]types.User)      // Online ready for CALL
// 	inCallClientsData     = make(map[int64]types.User)      // Online Incall users
// 	waitingForCallClients = make(map[int64]types.User)      // Waiting For Joinin
// 	allClientsData        = make(map[int64]types.User)      // just for getting data of users
// 	mutex                 sync.RWMutex
// )

// func handleClient(conn *websocket.Conn, db storage.Storage) {
// 	var userID int64

// 	// defer func(uid int64) {
// 	// 	fmt.Println("closing connection for user: ", uid)
// 	// 	delete(clients, uid)
// 	// 	delete(availableClientsData, uid)
// 	// 	delete(inCallClientsData, uid)
// 	// 	delete(waitingForCallClients, uid)
// 	// 	delete(allClientsData, uid)
// 	// 	if err := conn.Close(); err != nil {
// 	// 		fmt.Println("connection close failed : ", err)
// 	// 		// Handle error (e.g., log it)s
// 	// 	}
// 	// }(userID)

// 	for {
// 		var msg Message
// 		if err := conn.ReadJSON(&msg); err != nil {
// 			fmt.Println("invalid request : ", err)
// 			break // Handle other errors (e.g., log them)
// 		}

// 		conn.WriteJSON(map[string]string{
// 			"data": "connected to websocket",
// 		})

// 		switch msg.Type {
// 		case "initialize":
// 			userID = msg.User.Id
// 			clients[userID] = conn
// 			availableClientsData[userID] = msg.User
// 			allClientsData[userID] = msg.User

// 			//precaution is greater than cure
// 			delete(waitingForCallClients, userID)
// 			delete(inCallClientsData, userID)

// 			usersList, err := ShowRelatedUsersList(userID)
// 			if err != nil {
// 				fmt.Println("err in show related user list ", err.Error())
// 			}
// 			err = conn.WriteJSON(map[string]any{
// 				"type":         "initialized",
// 				"user":         msg.User,
// 				"online_users": usersList, // optional function
// 				"message":      "connected successfully",
// 			})
// 			fmt.Println("running")
// 			if err != nil {
// 				fmt.Print("err in initializaion phase : ", err)
// 			}

// 		case "offer", "icecandidate":
// 			targetConn, ok := clients[msg.Target]
// 			fmt.Println("OFFER TYPE : ", msg.RandomCall)
// 			if !ok {
// 				log.Printf("❌ Target user %d not connected", msg.Target)
// 				continue
// 			}
// 			err := targetConn.WriteJSON(Message{
// 				Type:    msg.Type,
// 				From:    msg.From,
// 				Payload: msg.Payload,
// 				FromUserData: map[string]any{
// 					"id":          allClientsData[msg.From].Id,
// 					"full_name":   allClientsData[msg.From].FullName,
// 					"username":    allClientsData[msg.From].Username,
// 					"profile_pic": allClientsData[msg.From].ProfilePic,
// 					"gender":      allClientsData[msg.From].Gender,
// 					"age":         allClientsData[msg.From].Age,
// 				},
// 				RandomCall: msg.RandomCall,
// 				Target:     msg.Target,
// 			})
// 			if err != nil {
// 				log.Printf("🚨 Error writing to WebSocket: %v", err)
// 			}

// 		case "answer":
// 			inCallClientsData[msg.From] = availableClientsData[msg.From]
// 			inCallClientsData[msg.Target] = availableClientsData[msg.Target]

// 			targetConn, ok := clients[msg.Target]
// 			if !ok {
// 				log.Printf("❌ Target user %d not connected", msg.Target)
// 				continue
// 			}
// 			err := targetConn.WriteJSON(Message{
// 				Type:    msg.Type,
// 				From:    msg.From,
// 				Payload: msg.Payload,
// 				FromUserData: map[string]any{
// 					"id":          allClientsData[msg.From].Id,
// 					"full_name":   allClientsData[msg.From].FullName,
// 					"username":    allClientsData[msg.From].Username,
// 					"profile_pic": allClientsData[msg.From].ProfilePic,
// 					"gender":      allClientsData[msg.From].Gender,
// 					"age":         allClientsData[msg.From].Age,
// 				},
// 				Target: msg.Target,
// 			})

// 			delete(availableClientsData, msg.From)
// 			delete(availableClientsData, msg.Target)
// 			delete(waitingForCallClients, msg.From)
// 			delete(waitingForCallClients, msg.Target)

// 			// add Start call details to Database
// 			// go func() {
// 			// 	mutex.Lock()
// 			// 	defer mutex.Unlock()
// 			// 	callId, err := db.StartCall(msg.Target, msg.From)
// 			// 	if err != nil {
// 			// 		fmt.Println("error in start call ", err)
// 			// 	}
// 			// 	clients[msg.From].WriteJSON(map[string]string{
// 			// 		"type":   "callStarted",
// 			// 		"callId": callId,
// 			// 	})
// 			// 	clients[msg.Target].WriteJSON(map[string]string{
// 			// 		"type":   "callStarted",
// 			// 		"callId": callId,
// 			// 	})
// 			// }()

// 			if err != nil {
// 				log.Printf("🚨 Error writing to WebSocket: %v", err)
// 			}

// 		case "randomCall":
// 			From := msg.From
// 			peerID, err := findAUser(From)
// 			if err != nil {
// 				log.Println("Error in Random Call Func:", err.Error())
// 				continue
// 			}
// 			clients[From].WriteJSON(map[string]any{
// 				"type":   "randomUserFound",
// 				"target": peerID,
// 				"targetUserData": map[string]any{
// 					"id":          allClientsData[peerID].Id,
// 					"full_name":   allClientsData[peerID].FullName,
// 					"username":    allClientsData[peerID].Username,
// 					"profile_pic": allClientsData[peerID].ProfilePic,
// 					"gender":      allClientsData[peerID].Gender,
// 					"age":         allClientsData[peerID].Age,
// 				},
// 			})

// 		// After connecting Ending the Call
// 		case "endCall":
// 			mutex.Lock()
// 			availableClientsData[msg.From] = allClientsData[msg.From]
// 			availableClientsData[msg.Target] = allClientsData[msg.Target]

// 			delete(inCallClientsData, msg.From)
// 			delete(inCallClientsData, msg.Target)
// 			mutex.Unlock()

// 			Target := clients[msg.Target]
// 			Target.WriteJSON(Message{
// 				Type:    "endCall",
// 				From:    msg.From,
// 				Payload: msg.Payload,
// 				FromUserData: map[string]any{
// 					"id":          allClientsData[msg.From].Id,
// 					"full_name":   allClientsData[msg.From].FullName,
// 					"username":    allClientsData[msg.From].Username,
// 					"profile_pic": allClientsData[msg.From].ProfilePic,
// 					"gender":      allClientsData[msg.From].Gender,
// 					"age":         allClientsData[msg.From].Age,
// 				},
// 			})
// 			//Update in database Call Ended
// 			// go func() {

// 			// }()

// 		// user canceled the matching
// 		case "cancelRandomMatch":
// 			mutex.Lock()
// 			availableClientsData[msg.From] = allClientsData[msg.From]
// 			delete(waitingForCallClients, msg.From)
// 			mutex.Unlock()
// 			clients[msg.From].WriteJSON(map[string]any{
// 				"type": "canceledRandomMatch",
// 			})

// 		// Direct Call Rejected
// 		case "rejectCall":
// 			delete(waitingForCallClients, msg.From)
// 			delete(waitingForCallClients, msg.Target)

// 			Target := msg.Target
// 			clients[Target].WriteJSON(Message{
// 				Type: "rejectedCall",
// 				From: msg.From,
// 				FromUserData: map[string]any{
// 					"id":          allClientsData[msg.From].Id,
// 					"full_name":   allClientsData[msg.From].FullName,
// 					"username":    allClientsData[msg.From].Username,
// 					"profile_pic": allClientsData[msg.From].ProfilePic,
// 					"gender":      allClientsData[msg.From].Gender,
// 					"age":         allClientsData[msg.From].Age,
// 				},
// 			})

// 		default:
// 			continue
// 		}
// 	}
// }

// func findAUser(id int64) (int64, error) {
// 	mutex.Lock()
// 	defer mutex.Unlock()
// 	for peer, _ := range waitingForCallClients {
// 		if id != peer {
// 			inCallClientsData[peer] = allClientsData[peer]
// 			inCallClientsData[id] = allClientsData[id]
// 			delete(waitingForCallClients, id)
// 			delete(waitingForCallClients, peer)
// 			return peer, nil
// 		}
// 	}
// 	waitingForCallClients[id] = availableClientsData[id]
// 	delete(availableClientsData, id)

// 	go func(uid int64) {
// 		time.Sleep(2 * time.Minute)
// 		mutex.Lock()
// 		defer mutex.Unlock()
// 		availableClientsData[uid] = availableClientsData[uid]
// 		delete(waitingForCallClients, uid)
// 	}(id)

// 	return 0, errors.New("did not find any user")
// }

// func ShowRelatedUsersList(id int64) ([]types.User, error) {
// 	// returning array of users maps
// 	mutex.Lock()
// 	defer mutex.Unlock()
// 	var msg []types.User
// 	maxCounter := 0
// 	fmt.Println("available clients data : ", availableClientsData)
// 	if len(availableClientsData) < 1 {
// 		return msg, fmt.Errorf("Users are not Online at this time")
// 	}

// 	for _, data := range availableClientsData {
// 		if len(msg) < 30 && data.Id != id {
// 			if maxCounter > 30 {
// 				return msg, nil
// 			}
// 			// fmt.Println("user: ",data)
// 			msg = append(msg, data)
// 			maxCounter += 1
// 		}
// 	}

// 	return msg, nil
// }

// package signalling

// import (
// 	"errors"
// 	"fmt"
// 	"github/english-app/internal/types"
// 	"github/english-app/storage"
// 	"log"
// 	"sync"
// 	"time"

// 	"github.com/gorilla/websocket"
// )

// type Message struct {
// 	Type         string     `json:"type"`
// 	User         types.User `json:"user,omitempty"`
// 	Target       int64      `json:"target,omitempty"`
// 	Payload      any        `json:"payload,omitempty"`
// 	From         int64      `json:"from,omitempty"`
// 	FromUserData any        `json:"fromUserData,omitempty"`
// 	RandomCall   bool       `json:"randomCall,omitempty"`
// }

// // Client wraps a websocket connection with a send channel
// type Client struct {
// 	Conn *websocket.Conn
// 	Send chan any
// }

// func (c *Client) writePump() {
// 	for msg := range c.Send {
// 		if err := c.Conn.WriteJSON(msg); err != nil {
// 			log.Println("❌ writePump error:", err)
// 			return
// 		}
// 	}
// }

// var (
// 	clients               = make(map[int64]*Client)    // All clients connections
// 	availableClientsData  = make(map[int64]types.User) // Online ready for CALL
// 	inCallClientsData     = make(map[int64]types.User) // Online Incall users
// 	waitingForCallClients = make(map[int64]types.User) // Waiting For Joinin
// 	allClientsData        = make(map[int64]types.User) // just for getting data of users
// 	mutex                 sync.RWMutex
// )

// func handleClient(conn *websocket.Conn, db storage.Storage) {
// 	var userID int64

// 	client := &Client{
// 		Conn: conn,
// 		Send: make(chan any, 20), // buffered channel to avoid blocking
// 	}
// 	go client.writePump()

// 	defer func(uid int64) {
// 		fmt.Println("closing connection for user:", uid)
// 		close(client.Send) // stop writePump
// 		mutex.Lock()
// 		delete(clients, uid)
// 		delete(availableClientsData, uid)
// 		delete(inCallClientsData, uid)
// 		delete(waitingForCallClients, uid)
// 		delete(allClientsData, uid)
// 		mutex.Unlock()
// 		conn.Close()
// 	}(userID)

// 	for {
// 		var msg Message
// 		if err := conn.ReadJSON(&msg); err != nil {
// 			fmt.Println("invalid request:", err)
// 			break
// 		}

// 		// client.Send <- map[string]string{"data": "connected to websocket"}

// 		switch msg.Type {
// 		case "initialize":
// 			userID = msg.User.Id
// 			mutex.Lock()
// 			clients[userID] = client
// 			availableClientsData[userID] = msg.User
// 			allClientsData[userID] = msg.User
// 			delete(waitingForCallClients, userID)
// 			delete(inCallClientsData, userID)
// 			mutex.Unlock()

// 			usersList, err := ShowRelatedUsersList(userID)
// 			if err != nil {
// 				fmt.Println("err in show related user list:", err.Error())
// 			}

// 			client.Send <- map[string]any{
// 				"type":         "initialized",
// 				"user":         msg.User,
// 				"online_users": usersList,
// 				"message":      "connected successfully",
// 			}

// 		case "offer", "icecandidate":
// 			mutex.RLock()
// 			target, ok := clients[msg.Target]
// 			mutex.RUnlock()
// 			if !ok {
// 				log.Printf("❌ Target user %d not connected", msg.Target)
// 				continue
// 			}
// 			target.Send <- Message{
// 				Type:    msg.Type,
// 				From:    msg.From,
// 				Payload: msg.Payload,
// 				FromUserData: map[string]any{
// 					"id":          allClientsData[msg.From].Id,
// 					"full_name":   allClientsData[msg.From].FullName,
// 					"username":    allClientsData[msg.From].Username,
// 					"profile_pic": allClientsData[msg.From].ProfilePic,
// 					"gender":      allClientsData[msg.From].Gender,
// 					"age":         allClientsData[msg.From].Age,
// 				},
// 				RandomCall: msg.RandomCall,
// 				Target:     msg.Target,
// 			}

// 		case "answer":
// 			mutex.Lock()
// 			inCallClientsData[msg.From] = availableClientsData[msg.From]
// 			inCallClientsData[msg.Target] = availableClientsData[msg.Target]
// 			delete(availableClientsData, msg.From)
// 			delete(availableClientsData, msg.Target)
// 			delete(waitingForCallClients, msg.From)
// 			delete(waitingForCallClients, msg.Target)
// 			target, ok := clients[msg.Target]
// 			mutex.Unlock()

// 			if !ok {
// 				log.Printf("❌ Target user %d not connected", msg.Target)
// 				continue
// 			}
// 			target.Send <- Message{
// 				Type:    msg.Type,
// 				From:    msg.From,
// 				Payload: msg.Payload,
// 				FromUserData: map[string]any{
// 					"id":          allClientsData[msg.From].Id,
// 					"full_name":   allClientsData[msg.From].FullName,
// 					"username":    allClientsData[msg.From].Username,
// 					"profile_pic": allClientsData[msg.From].ProfilePic,
// 					"gender":      allClientsData[msg.From].Gender,
// 					"age":         allClientsData[msg.From].Age,
// 				},
// 				Target: msg.Target,
// 			}

// 		case "randomCall":
// 			From := msg.From
// 			peerID, err := findAUser(From)
// 			if err != nil {
// 				log.Println("Error in Random Call Func:", err.Error())
// 				continue
// 			}
// 			clients[From].Send <- map[string]any{
// 				"type":           "randomUserFound",
// 				"target":         peerID,
// 				"targetUserData": allClientsData[peerID],
// 			}
// 			fmt.Printf("%s got connected to %s", allClientsData[From].FullName, allClientsData[peerID].FullName)

// 		case "endCall":
// 			mutex.Lock()
// 			availableClientsData[msg.From] = allClientsData[msg.From]
// 			availableClientsData[msg.Target] = allClientsData[msg.Target]
// 			delete(inCallClientsData, msg.From)
// 			delete(inCallClientsData, msg.Target)
// 			target := clients[msg.Target]
// 			mutex.Unlock()

// 			target.Send <- Message{
// 				Type:         "endCall",
// 				From:         msg.From,
// 				Payload:      msg.Payload,
// 				FromUserData: allClientsData[msg.From],
// 			}

// 		case "cancelRandomMatch":
// 			mutex.Lock()
// 			availableClientsData[msg.From] = allClientsData[msg.From]
// 			delete(waitingForCallClients, msg.From)
// 			mutex.Unlock()
// 			clients[msg.From].Send <- map[string]any{"type": "canceledRandomMatch"}

// 		case "rejectCall":
// 			mutex.Lock()
// 			delete(waitingForCallClients, msg.From)
// 			delete(waitingForCallClients, msg.Target)
// 			target := clients[msg.Target]
// 			mutex.Unlock()

// 			target.Send <- Message{
// 				Type:         "rejectedCall",
// 				From:         msg.From,
// 				FromUserData: allClientsData[msg.From],
// 			}

// 		default:
// 			continue
// 		}
// 	}
// }

// func findAUser(id int64) (int64, error) {
// 	mutex.Lock()
// 	defer mutex.Unlock()
// 	if len(waitingForCallClients) > 0 {
// 		for peer := range waitingForCallClients {
// 			if id != peer {
// 				inCallClientsData[peer] = allClientsData[peer]
// 				inCallClientsData[id] = allClientsData[id]
// 				delete(waitingForCallClients, id)
// 				delete(waitingForCallClients, peer)
// 				return peer, nil
// 			}
// 		}
// 	}
// 	waitingForCallClients[id] = availableClientsData[id]
// 	delete(availableClientsData, id)

// 	go func(uid int64) {
// 		time.Sleep(2 * time.Minute)
// 		mutex.Lock()
// 		defer mutex.Unlock()
// 		availableClientsData[uid] = allClientsData[uid]
// 		delete(waitingForCallClients, uid)
// 	}(id)

// 	return 0, errors.New("did not find any user")
// }

//	func ShowRelatedUsersList(id int64) ([]types.User, error) {
//		mutex.RLock()
//		defer mutex.RUnlock()
//		var msg []types.User
//		maxCounter := 0
//		if len(availableClientsData) < 1 {
//			return msg, fmt.Errorf("Users are not Online at this time")
//		}
//		for _, data := range availableClientsData {
//			if len(msg) < 30 {
//				if maxCounter > 30 {
//					return msg, nil
//				}
//				msg = append(msg, data)
//				maxCounter++
//			}
//		}
//		return msg, nil
//	}

package signalling

import (
	"errors"
	"fmt"
	"github/english-app/internal/types"
	"github/english-app/storage"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type         string     `json:"type"`
	User         types.User `json:"user,omitempty"`
	Target       int64      `json:"target,omitempty"`
	Payload      any        `json:"payload,omitempty"`
	From         int64      `json:"from,omitempty"`
	FromUserData any        `json:"fromUserData,omitempty"`
	RandomCall   bool       `json:"randomCall,omitempty"`
	CallId       string     `json:"callId,omitempty"`
}

// Client wraps a websocket connection with a send channel
type Client struct {
	Conn       *websocket.Conn
	Send       chan any
	LastActive time.Time
}

func (c *Client) writePump() {
	for msg := range c.Send {
		if err := c.Conn.WriteJSON(msg); err != nil {
			log.Println("❌ writePump error:", err)
			return
		}
	}
}

func (c *Client) isAlive() bool {
	err := c.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(2*time.Second))
	return err == nil
}

var (
	clients               = make(map[int64]*Client)    // All clients connections
	availableClientsData  = make(map[int64]types.User) // Online ready for CALL
	inCallClientsData     = make(map[int64]types.User) // Online Incall users
	waitingForCallClients = make(map[int64]types.User) // Waiting For Joinin
	allClientsData        = make(map[int64]types.User) // just for getting data of users
	mutex                 sync.RWMutex
)

// Helper function to check if user is connected
func isUserConnected(userID int64) bool {
	mutex.RLock()
	defer mutex.RUnlock()
	_, exists := clients[userID]
	return exists
}

// Helper function to send user offline response
func sendUserOfflineResponse(client *Client, targetUserID int64) {
	client.Send <- map[string]any{
		"type":    "userOffline",
		"target":  targetUserID,
		"message": "User is offline",
	}
}

func handleClient(conn *websocket.Conn, db storage.Storage) {
	var userID int64

	client := &Client{
		Conn: conn,
		Send: make(chan any, 20), // buffered channel to avoid blocking
	}
	go client.writePump()

	// defer func(uid int64) {
	// 	fmt.Println("closing connection for user:", uid)
	// 	close(client.Send) // stop writePump
	// 	mutex.Lock()
	// 	delete(clients, uid)
	// 	delete(availableClientsData, uid)
	// 	delete(inCallClientsData, uid)
	// 	delete(waitingForCallClients, uid)
	// 	delete(allClientsData, uid)
	// 	mutex.Unlock()
	// 	conn.Close()
	// }(userID)

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			fmt.Println("invalid request:", err)
			break
		}

		switch msg.Type {
		case "initialize":
			userID = msg.User.Id
			mutex.Lock()
			clients[userID] = client
			availableClientsData[userID] = msg.User
			allClientsData[userID] = msg.User
			delete(waitingForCallClients, userID)
			delete(inCallClientsData, userID)
			mutex.Unlock()

			usersList, err := ShowRelatedUsersList(userID)
			if err != nil {
				fmt.Println("err in show related user list:", err.Error())
			}

			client.Send <- map[string]any{
				"type":         "initialized",
				"user":         msg.User,
				"online_users": usersList,
				"message":      "connected successfully",
			}

		case "offer", "icecandidate":
			// Check if target user is connected
			if !isUserConnected(msg.Target) {
				log.Printf("❌ Target user %d is offline", msg.Target)
				sendUserOfflineResponse(client, msg.Target)
				continue
			}

			mutex.RLock()
			target := clients[msg.Target]
			mutex.RUnlock()

			target.Send <- Message{
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
				RandomCall: msg.RandomCall,
				Target:     msg.Target,
			}

		case "answer":
			// Check if target user is connected
			if !isUserConnected(msg.Target) {
				log.Printf("❌ Target user %d is offline", msg.Target)
				sendUserOfflineResponse(client, msg.Target)
				continue
			}

			mutex.Lock()
			inCallClientsData[msg.From] = availableClientsData[msg.From]
			inCallClientsData[msg.Target] = availableClientsData[msg.Target]
			delete(availableClientsData, msg.From)
			delete(availableClientsData, msg.Target)
			delete(waitingForCallClients, msg.From)
			delete(waitingForCallClients, msg.Target)
			target := clients[msg.Target]
			from := clients[msg.From]
			mutex.Unlock()

			target.Send <- Message{
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
			}
			// add Start call details to Database
			go func() {
				mutex.Lock()
				defer mutex.Unlock()
				callId, err := db.StartCall(msg.Target, msg.From)
				if err != nil {
					fmt.Println("error in start call ", err)
				}
				from.Send <- map[string]string{
					"type":   "callStarted",
					"callId": callId,
				}
				target.Send <- map[string]string{
					"type":   "callStarted",
					"callId": callId,
				}
			}()

		case "randomCall":
			From := msg.From
			peerID, err := findAUser(From)
			if err != nil {
				log.Println("Error in Random Call Func:", err.Error())
				client.Send <- map[string]any{
					"type":    "randomCallError",
					"message": err.Error(),
				}
				continue
			}
			clients[From].Send <- map[string]any{
				"type":           "randomUserFound",
				"target":         peerID,
				"targetUserData": allClientsData[peerID],
			}
			fmt.Printf("%s got connected to %s", allClientsData[From].FullName, allClientsData[peerID].FullName)

		case "endCall":
			fmt.Println("End Call Message Received for Call ID:", msg.CallId)
			// Check if target user is connected
			if !isUserConnected(msg.Target) {
				log.Printf("❌ Target user %d is offline during endCall", msg.Target)
				// Still update the caller's state even if target is offline
				mutex.Lock()
				availableClientsData[msg.From] = allClientsData[msg.From]
				delete(inCallClientsData, msg.From)
				delete(inCallClientsData, msg.Target)
				mutex.Unlock()
				continue
			}

			mutex.Lock()
			availableClientsData[msg.From] = allClientsData[msg.From]
			availableClientsData[msg.Target] = allClientsData[msg.Target]
			delete(inCallClientsData, msg.From)
			delete(inCallClientsData, msg.Target)
			target := clients[msg.Target]
			mutex.Unlock()

			target.Send <- Message{
				Type:         "endCall",
				From:         msg.From,
				Payload:      msg.Payload,
				FromUserData: allClientsData[msg.From],
			}
			//Update in database Call Ended
			go func(uid string) {
				err := db.EndCall(uid)
				if err != nil {
					fmt.Println("error in end call ", err)
				}
			}(msg.CallId)

		case "cancelRandomMatch":
			mutex.Lock()
			availableClientsData[msg.From] = allClientsData[msg.From]
			delete(waitingForCallClients, msg.From)
			mutex.Unlock()
			clients[msg.From].Send <- map[string]any{"type": "canceledRandomMatch"}

		case "rejectCall":
			// Check if target user is connected
			if !isUserConnected(msg.Target) {
				log.Printf("❌ Target user %d is offline during rejectCall", msg.Target)
				// Still clean up the states
				mutex.Lock()
				delete(waitingForCallClients, msg.From)
				delete(waitingForCallClients, msg.Target)
				mutex.Unlock()
				continue
			}

			mutex.Lock()
			delete(waitingForCallClients, msg.From)
			delete(waitingForCallClients, msg.Target)
			target := clients[msg.Target]
			mutex.Unlock()

			target.Send <- Message{
				Type:         "rejectedCall",
				From:         msg.From,
				FromUserData: allClientsData[msg.From],
			}

		case "refreshList":
			id := msg.From
			target := clients[id]
			usersList, err := ShowRelatedUsersList(id)
			if err != nil {
				target.Send <- map[string]any{
					"type":       "newUsersList",
					"usersCount": len(usersList),
					"error":      err.Error(),
				}
				fmt.Println("err in refresh list :", err.Error())
				break
			}
			target.Send <- map[string]any{
				"type":       "newUsersList",
				"usersCount": len(usersList),
				"usersList":  usersList,
			}

		default:
			continue
		}
	}
}

func findAUser(id int64) (int64, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if len(waitingForCallClients) > 0 {
		for peer := range waitingForCallClients {
			if id != peer {
				// Double check if the peer is still connected
				if _, exists := clients[peer]; !exists {
					// Remove disconnected user from waiting list
					delete(waitingForCallClients, peer)
					continue
				}

				inCallClientsData[peer] = allClientsData[peer]
				inCallClientsData[id] = allClientsData[id]
				delete(waitingForCallClients, id)
				delete(waitingForCallClients, peer)
				return peer, nil
			}
		}
	}
	waitingForCallClients[id] = availableClientsData[id]
	delete(availableClientsData, id)

	go maximumWaitingTimeExceededChecker(id)

	return 0, errors.New("did not find any user")
}

func maximumWaitingTimeExceededChecker(uid int64) {
	// Maximum wait time of 2 minutes in random matching
	time.Sleep(2 * time.Minute)
	mutex.Lock()
	defer mutex.Unlock()
	// Check if user is still connected before moving back to available
	if _, stillWaiting := waitingForCallClients[uid]; stillWaiting {
		delete(waitingForCallClients, uid)
		if _, isConnected := clients[uid]; isConnected {
			// User disconnected while waiting
			availableClientsData[uid] = allClientsData[uid]
			fmt.Println("User", uid, "timed out after 2 mins, moved to idle.")
		}
	}
}
func ShowRelatedUsersList(id int64) ([]types.User, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	var msg []types.User
	maxCounter := 0
	fmt.Println("online users : ", len(availableClientsData))
	if len(availableClientsData) < 2 {
		fmt.Println("No Users are Online at this time")
		return msg, fmt.Errorf(" No Users are Online at this time")
	}

	// Only show users that are actually connected
	for userID, data := range availableClientsData {
		if len(msg) >= 30 || maxCounter > 30 {
			break
		}

		// Verify user is still connected
		_, exists := clients[userID]

		if exists && userID != id {
			msg = append(msg, data)
			maxCounter++
		}
	}
	return msg, nil
}

func StartUserChecker() {
	go func() {
		for {
			time.Sleep(15 * time.Second)
			if len(clients) <= 0 {
				continue
			}
			for id, client := range clients {
				if time.Since(client.LastActive) > 15*time.Second {
					go func(uid int64, c *Client) {
						fmt.Println("checking user after 15 sec", allClientsData[id].FullName)
						if !c.isAlive() {
							fmt.Println("removing user after 15 sec", allClientsData[id].FullName)
							mutex.Lock()
							defer mutex.Unlock()
							delete(clients, uid)
							delete(availableClientsData, uid)
							delete(inCallClientsData, uid)
							delete(waitingForCallClients, uid)
							delete(allClientsData, uid)
							c.Conn.Close()
							close(c.Send)
						}
					}(id, client)
				}
			}
		}
	}()

}
