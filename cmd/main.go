package main

import (
	"fmt"
	authhandler "github/english-app/internal/auth/handler"
	"github/english-app/internal/auth/middleware"
	"github/english-app/internal/auth/token"
	"github/english-app/internal/notifications"
	service "github/english-app/internal/rateLimiting"
	"github/english-app/internal/signalling"
	userhandler "github/english-app/internal/user/handler"
	"github/english-app/storage/postgresql"
	"github/english-app/storage/redis"
	"github/english-app/storage/s3"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("server running...")

	// ENV LOADING
	_ = godotenv.Load()

	// FIREBASE initialization
	filePath := "strango.json"
	notifications.EnsureFirebaseKeyFile(filePath)

	// JWT MAKER
	jwtMaker := token.NewJWTMaker(os.Getenv("JWTTOKEN_SECRET"))
	if jwtMaker == nil {
		fmt.Println("Error creating JWT maker")
		return
	}

	// CHECK WEBSOCKET USER CONNECTION / PRESENCE
	signalling.StartUserChecker()

	// S3 OBJECT STORAGE
	R2_ACCOUNT_ID := os.Getenv("R2_ACCOUNT_ID")
	R2_BUCKET_ACCESS_KEY := os.Getenv("R2_BUCKET_ACCESS_KEY")
	R2_BUCKET_SECRET_ACCESS_KEY := os.Getenv("R2_BUCKET_SECRET_ACCESS_KEY")
	R2_REGION := os.Getenv("R2_REGION")
	s3Client := s3.NewS3Client(R2_ACCOUNT_ID, R2_BUCKET_ACCESS_KEY, R2_BUCKET_SECRET_ACCESS_KEY, R2_REGION)

	// SQL INSTANCE
	storage, err := postgresql.New()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}

	// REDIS INSTANCE
	redisClient, err := redis.New()
	if err != nil {
		fmt.Println("Error connecting to Redis:", err)
		return
	}

	// ROUTER
	r := gin.Default()
	// RATE LIMITER
	r.Use(service.RateLimiter())
	go service.CleanupVisitors() // launch the cleanup goroutine

	r.MaxMultipartMemory = 8 << 20 // 8 MiB
	// UNPROTECTED ROUTES

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
		userGroup.POST("/upload/image", userhandler.UploadImageHandler(storage, s3Client))
		userGroup.GET("/profile", userhandler.GetProfileHandler(storage))
		userGroup.GET("/userprofile/:id", userhandler.GetOtherUserProfileHandler(storage))
		userGroup.GET("/call/history", userhandler.GetCallHistoryHandler(storage))
		userGroup.GET("/account/delete", userhandler.DeleteAccountHandler(storage))
		userGroup.GET("/picture/delete", userhandler.DeletePictureHandler(storage, s3Client))
		userGroup.GET("/block/:id", userhandler.BlockUserHandler(storage))
		userGroup.GET("/leaderboard", userhandler.LeaderboardHandler(storage))
		userGroup.GET("/aicharacters", userhandler.AiCharactersHandler())
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local dev
	}

	r.Run(":" + port) // Gin binds to specified port
}
