package signalling

import (
	"fmt"
	"github/english-app/storage"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for WebSocket connections
	},
}

func HandleWebSocket(storage *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("getting there ")
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		go handleClient(conn, storage)
	}
}
