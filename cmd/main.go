package main

import (
	"fmt"
	authhandler "github/english-app/internal/auth/handler"
	"github/english-app/internal/auth/token"
	"github/english-app/storage/postgresql"
	"github/english-app/storage/redis"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("server running...")
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}
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
		authGroup.POST("/logingoogle", authhandler.GoogleLoginHandler(storage, jwtMaker, redisClient))
		authGroup.POST("/creategoogleuser", authhandler.GoogleCreateHandler(storage, jwtMaker, redisClient)) // Assuming this is the same handler for creating a user
		authGroup.POST("/emaillogin", authhandler.EmailLoginHandler(storage, jwtMaker, redisClient))
		authGroup.POST("/createemailuser", authhandler.EmailCreateHandler(storage, jwtMaker, redisClient))
		authGroup.GET("/checkuser", authhandler.CheckUsernameIsAvailable(storage, redisClient))
		authGroup.POST("/forgetPassword", authhandler.ForgetPasswordHandler(storage, redisClient))
		authGroup.POST("/resetPassword", authhandler.ResetPasswordHandler(storage, redisClient))
	}
	r.Run(":8082") // listen and serve on
}
