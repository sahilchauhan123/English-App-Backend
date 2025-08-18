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
}

// Client wraps a websocket connection with a send channel
type Client struct {
	Conn *websocket.Conn
	Send chan any
}

func (c *Client) writePump() {
	for msg := range c.Send {
		if err := c.Conn.WriteJSON(msg); err != nil {
			log.Println("❌ writePump error:", err)
			return
		}
	}
}

var (
	clients               = make(map[int64]*Client)    // All clients connections
	availableClientsData  = make(map[int64]types.User) // Online ready for CALL
	inCallClientsData     = make(map[int64]types.User) // Online Incall users
	waitingForCallClients = make(map[int64]types.User) // Waiting For Joinin
	allClientsData        = make(map[int64]types.User) // just for getting data of users
	mutex                 sync.RWMutex
)

func handleClient(conn *websocket.Conn, db storage.Storage) {
	var userID int64

	client := &Client{
		Conn: conn,
		Send: make(chan any, 20), // buffered channel to avoid blocking
	}
	go client.writePump()

	defer func(uid int64) {
		fmt.Println("closing connection for user:", uid)
		close(client.Send) // stop writePump
		mutex.Lock()
		delete(clients, uid)
		delete(availableClientsData, uid)
		delete(inCallClientsData, uid)
		delete(waitingForCallClients, uid)
		delete(allClientsData, uid)
		mutex.Unlock()
		conn.Close()
	}(userID)

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			fmt.Println("invalid request:", err)
			break
		}

		client.Send <- map[string]string{"data": "connected to websocket"}

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
			mutex.RLock()
			target, ok := clients[msg.Target]
			mutex.RUnlock()
			if !ok {
				log.Printf("❌ Target user %d not connected", msg.Target)
				continue
			}
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
			mutex.Lock()
			inCallClientsData[msg.From] = availableClientsData[msg.From]
			inCallClientsData[msg.Target] = availableClientsData[msg.Target]
			delete(availableClientsData, msg.From)
			delete(availableClientsData, msg.Target)
			delete(waitingForCallClients, msg.From)
			delete(waitingForCallClients, msg.Target)
			target, ok := clients[msg.Target]
			mutex.Unlock()

			if !ok {
				log.Printf("❌ Target user %d not connected", msg.Target)
				continue
			}
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

		case "randomCall":
			From := msg.From
			peerID, err := findAUser(From)
			if err != nil {
				log.Println("Error in Random Call Func:", err.Error())
				continue
			}
			clients[From].Send <- map[string]any{
				"type":           "randomUserFound",
				"target":         peerID,
				"targetUserData": allClientsData[peerID],
			}

		case "endCall":
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

		case "cancelRandomMatch":
			mutex.Lock()
			availableClientsData[msg.From] = allClientsData[msg.From]
			delete(waitingForCallClients, msg.From)
			mutex.Unlock()
			clients[msg.From].Send <- map[string]any{"type": "canceledRandomMatch"}

		case "rejectCall":
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

		default:
			continue
		}
	}
}

func findAUser(id int64) (int64, error) {
	mutex.Lock()
	defer mutex.Unlock()
	for peer := range waitingForCallClients {
		if id != peer {
			inCallClientsData[peer] = allClientsData[peer]
			inCallClientsData[id] = allClientsData[id]
			delete(waitingForCallClients, id)
			delete(waitingForCallClients, peer)
			return peer, nil
		}
	}
	waitingForCallClients[id] = availableClientsData[id]
	delete(availableClientsData, id)

	go func(uid int64) {
		time.Sleep(2 * time.Minute)
		mutex.Lock()
		defer mutex.Unlock()
		availableClientsData[uid] = allClientsData[uid]
		delete(waitingForCallClients, uid)
	}(id)

	return 0, errors.New("did not find any user")
}

func ShowRelatedUsersList(id int64) ([]types.User, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	var msg []types.User
	maxCounter := 0
	if len(availableClientsData) < 1 {
		return msg, fmt.Errorf("Users are not Online at this time")
	}
	for _, data := range availableClientsData {
		if len(msg) < 30 && data.Id != id {
			if maxCounter > 30 {
				return msg, nil
			}
			msg = append(msg, data)
			maxCounter++
		}
	}
	return msg, nil
}

// package signalling

// import (
// 	"errors"
// 	"fmt"
// 	"log"
// 	"sync"
// 	"time"

// 	"github/english-app/internal/types"
// 	"github/english-app/storage"

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

// // Move user to "waiting" state
// func setUserWaiting(id int64, user types.User) {
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	// Remove from other maps
// 	delete(availableClientsData, id)
// 	delete(inCallClientsData, id)

// 	// Add to waiting map
// 	waitingForCallClients[id] = user
// }

// // Move user to "available" state
// func setUserAvailable(id int64, user types.User) {
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	// Remove from other maps
// 	delete(waitingForCallClients, id)
// 	delete(inCallClientsData, id)

// 	// Add to available map
// 	availableClientsData[id] = user
// }

// // Move user to "in call" state
// func setUserInCall(id int64, user types.User) {
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	// Remove from other maps
// 	delete(waitingForCallClients, id)
// 	delete(availableClientsData, id)

// 	// Add to in-call map
// 	inCallClientsData[id] = user
// }

// func handleClient(conn *websocket.Conn, db storage.Storage) {
// 	var userID int64

// 	defer func(uid int64) {
// 		fmt.Println("closing connection for user: ", uid)
// 		delete(clients, uid)
// 		delete(availableClientsData, uid)
// 		delete(inCallClientsData, uid)
// 		delete(waitingForCallClients, uid)
// 		delete(allClientsData, uid)
// 		if err := conn.Close(); err != nil {
// 			fmt.Println("connection close failed : ", err)
// 			// Handle error (e.g., log it)s
// 		}
// 	}(userID)

// 	for {
// 		var msg Message
// 		if err := conn.ReadJSON(&msg); err != nil {
// 			fmt.Println("invalid request : ", err)
// 			break // Handle other errors (e.g., log them)
// 		}

// 		switch msg.Type {
// 		case "initialize":
// 			userID = msg.User.Id
// 			clients[userID] = conn
// 			setUserAvailable(userID, msg.User)
// 			allClientsData[userID] = msg.User

// 			usersList, err := ShowRelatedUsersList(userID)
// 			if err != nil {
// 				fmt.Println("err in show related user list ", err.Error())
// 			}
// 			err = conn.WriteJSON(map[string]any{
// 				"type":         "initialized",
// 				"user":         msg.User,
// 				"online_users": usersList,
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
// 			setUserInCall(msg.From, allClientsData[msg.From])
// 			setUserInCall(msg.Target, allClientsData[msg.Target])

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

// 		case "endCall":
// 			setUserAvailable(msg.From, allClientsData[msg.From])
// 			setUserAvailable(msg.Target, allClientsData[msg.Target])

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

// 		case "cancelRandomMatch":
// 			setUserAvailable(msg.From, allClientsData[msg.From])
// 			clients[msg.From].WriteJSON(map[string]any{
// 				"type": "canceledRandomMatch",
// 			})

// 		case "rejectCall":
// 			if _, ok := waitingForCallClients[msg.From]; !ok {
// 				setUserAvailable(msg.From, allClientsData[msg.From])
// 			}
// 			setUserAvailable(msg.Target, allClientsData[msg.Target])

// 			Target := msg.Target
// 			clients[Target].WriteJSON(Message{
// 				Type: "callRejected",
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
// 			setUserInCall(peer, allClientsData[peer])
// 			setUserInCall(id, allClientsData[id])
// 			return peer, nil
// 		}
// 	}
// 	setUserWaiting(id, allClientsData[id])

// 	go func(uid int64) {
// 		time.Sleep(2 * time.Minute)
// 		setUserAvailable(uid, allClientsData[uid])
// 	}(id)

// 	return 0, errors.New("did not find any user")
// }

// func ShowRelatedUsersList(id int64) ([]types.User, error) {
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
// 			msg = append(msg, data)
// 			maxCounter += 1
// 		}
// 	}

// 	return msg, nil
// }
