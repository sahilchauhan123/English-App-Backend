package main

import (
	"fmt"
	authhandler "github/english-app/internal/auth/handler"
	"github/english-app/internal/auth/token"
	"github/english-app/internal/signalling"
	"github/english-app/storage/postgresql"
	"github/english-app/storage/redis"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("server running...")
	_ = godotenv.Load()
	// if err != nil {
	// 	fmt.Println("Error loading .env file")
	// 	return
	// }
	jwtMaker := token.NewJWTMaker(os.Getenv("JWTTOKEN_SECRET"))
	if jwtMaker == nil {
		fmt.Println("Error creating JWT maker")
		return
	}

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
		authGroup.POST("/google/login", authhandler.GoogleLoginHandler(storage, jwtMaker, redisClient))
		authGroup.POST("/google/signup", authhandler.GoogleCreateHandler(storage, jwtMaker, redisClient)) // Assuming this is the same handler for creating a user
		authGroup.POST("/email/login", authhandler.EmailLoginHandler(storage, jwtMaker, redisClient))
		authGroup.POST("/email/signup", authhandler.EmailCreateHandler(storage, jwtMaker, redisClient))
		authGroup.GET("/checkusername", authhandler.CheckUsernameIsAvailable(storage, redisClient))
		authGroup.POST("email/forgetPassword", authhandler.ForgetPasswordHandler(storage, redisClient))
		authGroup.POST("email/resetPassword", authhandler.ResetPasswordHandler(storage, redisClient))
		authGroup.POST("/refershToken", authhandler.UpdateTokenHandler(storage, redisClient, *jwtMaker))
	}
	r.GET("/ws", signalling.HandleWebSocket(storage))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local dev
	}
	r.Run(":" + port) // Gin binds to specified port
}
