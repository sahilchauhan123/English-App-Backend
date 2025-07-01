package main

import (
	"fmt"
	authhandler "github/english-app/internal/auth/handler"
	"github/english-app/internal/auth/token"
	"github/english-app/storage/postgresql"
	"github/english-app/storage/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("server running...")
	jwtMaker := token.NewJWTMaker("3232jfh793232sds")

	storage, err := postgresql.New()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}
	redisClient, err := redis.New()
	if err != nil {
		fmt.Println("Error connecting to Redis:", err)
		return
	}
	r := gin.Default()
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/logingoogle", authhandler.GoogleLoginHandler(storage, jwtMaker, redisClient))
		authGroup.POST("/creategoogleuser", authhandler.GoogleCreateHandler(storage, jwtMaker, redisClient)) // Assuming this is the same handler for creating a user
		// authGroup.POST("/emaillogin", authhandler.EmailLoginHandler(storage, jwtMaker, redisClient))
		authGroup.POST("/createemailuser", authhandler.EmailCreateHandler(storage, jwtMaker, redisClient))

	}
	r.Run(":8082") // listen and serve on
}
