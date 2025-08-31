package signalling

import (
	"fmt"
	"github/english-app/internal/auth/token"
	"github/english-app/storage"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for WebSocket connections
	},
}

// func HandleWebSocket(storage storage.Storage, jwt token.JWTMaker) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// 1. Extract token from header (or query param if you’re sending it that way)
// 		tokenString := c.GetHeader("Authorization")
// 		if tokenString == "" {
// 			// maybe from query param
// 			tokenString = c.Query("token")
// 		}

// 		if tokenString == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
// 			return
// 		}

// 		// if you’re sending `Authorization: Bearer <token>`, strip it
// 		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
// 			tokenString = tokenString[7:]
// 		}

// 		// 2. Verify JWT
// 		payload, err := jwt.VerifyToken(tokenString)
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
// 			return
// 		}

// 		fmt.Println("✅ Authenticated user:", payload)
// 		fmt.Println("getting there ")
// 		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
// 		if err != nil {
// 			return
// 		}

// 		go handleClient(conn, storage)
// 	}
// }

func HandleWebSocket(storage storage.Storage, jwtMaker token.JWTMaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("[WS] Incoming request...")

		// 1. Extract token
		tokenString := c.Query("token")
		log.Println("[WS] Token from query:", tokenString)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			return
		}

		// 2. Verify JWT
		payload, err := jwtMaker.VerifyToken(tokenString)
		if err != nil {
			fmt.Println("[WS] ❌ Token verification failed:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		fmt.Printf("[WS] ✅ Authenticated user: %+v\n", payload)

		// 3. Upgrade to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("[WS] ❌ Upgrade failed:", err)
			return
		}
		fmt.Println("[WS] 🔗 WebSocket upgrade success")

		go handleClient(conn, storage)
	}
}
