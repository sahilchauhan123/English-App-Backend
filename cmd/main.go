package main

import (
	"fmt"
	authhandler "github/english-app/internal/auth/handler"
	"github/english-app/internal/auth/middleware"
	"github/english-app/internal/auth/token"
	"github/english-app/internal/signalling"
	"github/english-app/storage/postgresql"
	"github/english-app/storage/redis"
	"github/english-app/storage/s3"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("server running...")
	_ = godotenv.Load()

	jwtMaker := token.NewJWTMaker(os.Getenv("JWTTOKEN_SECRET"))
	if jwtMaker == nil {
		fmt.Println("Error creating JWT maker")
		return
	}
	_ = s3.NewS3Client(os.Getenv("R2_ACCOUNT_ID"), os.Getenv("R2_BUCKET_ACCESS_KEY"), os.Getenv("R2_BUCKET_SECRET_ACCESS_KEY"), os.Getenv("R2_REGION"))
	// url, err := s3Client.PutObject(os.Getenv("R2_BUCKET_NAME"), "testing", 1000)
	// if err != nil {
	// 	fmt.Println("err in s3", err)
	// }
	// fmt.Println("url", url)
	// check user every 15 sec
	signalling.StartUserChecker()

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

	// UNPROTECTED ROUTES
	r := gin.Default()
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/google/login", authhandler.GoogleLoginHandler(storage, jwtMaker, redisClient))
		authGroup.POST("/google/signup", authhandler.GoogleCreateHandler(storage, jwtMaker, redisClient)) // Assuming this is the same handler for creating a user
		authGroup.POST("/email/login", authhandler.EmailLoginHandler(storage, jwtMaker, redisClient))
		authGroup.POST("/email/signup", authhandler.EmailCreateHandler(storage, jwtMaker, redisClient))
		authGroup.POST("/email/generateloginotp", authhandler.GenerateEmailLoginOtp(storage, jwtMaker, redisClient))
		authGroup.POST("/email/generatesignupotp", authhandler.GenerateEmailSignUpOtp(storage, jwtMaker, redisClient))
		authGroup.GET("/checkusername", authhandler.CheckUsernameIsAvailable(storage, redisClient))
		// authGroup.POST("/email/forgetPassword", authhandler.ForgetPasswordHandler(storage, redisClient))
		// authGroup.POST("/email/resetPassword", authhandler.ResetPasswordHandler(storage, redisClient))
		authGroup.POST("/refreshToken", authhandler.UpdateTokenHandler(storage, redisClient, *jwtMaker))
	}

	// WEBSOCKET
	r.GET("/ws", signalling.HandleWebSocket(storage, *jwtMaker))

	// PROTECTED ROUTES
	userGroup := r.Group("/api/user")
	userGroup.Use(middleware.AuthMiddleware(jwtMaker))
	{

	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local dev
	}

	r.Run(":" + port) // Gin binds to specified port
}
