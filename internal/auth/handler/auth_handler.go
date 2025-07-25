package authhandler

import (
	"context"
	"fmt"
	authservice "github/english-app/internal/auth/service"
	"github/english-app/internal/auth/token"
	"github/english-app/storage/redis"
	"time"

	// "github/english-app/internal/auth/token"
	"github/english-app/internal/response"
	"github/english-app/internal/types"
	"github/english-app/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

var ctx = context.Background()

type GoogleLoginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

type EmailLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Res struct {
	IsRegistered bool   `json:"is_registered"`
	Message      string `json:"message"`
}

type EmailPassWordForget struct {
	Email           string `json:"email" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

type EmailPassWordReset struct {
	Email string `json:"email" binding:"required"`
	Otp   string `json:"otp" binding:"required"`
}

type Token struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func GoogleLoginHandler(db storage.Storage, jwtMaker *token.JWTMaker, redisClient *redis.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req GoogleLoginRequest

		// validate the request body
		err := c.ShouldBindJSON(&req)
		if err != nil {
			response.Failed(c, http.StatusBadRequest, "Invalid Request")
			return
		}

		// Handle Google login
		userData, err := authservice.HandleGoogleLogin(req.IDToken, db)
		if err != nil {
			fmt.Println("Error handling Google login:", err)
			response.Failed(c, http.StatusInternalServerError, "Failed to login with Google")
			return
		}

		if !userData.IsRegistered {
			fmt.Println("User not found in database")
			response.Success(c, map[string]any{
				"isRegistered": false,
				"message":      "Please Create Your Account First",
			})
			return
		}

		// TOKENS CREATION
		// ACCESS TOKEN
		accessToken, err := jwtMaker.CreateToken(userData.User.Id, time.Hour*24*3)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, err.Error())
			return
		}
		// STORE TO REDIS
		err = redisClient.Client.Set(ctx, fmt.Sprintf("access_token:%d", userData.User.Id), accessToken, time.Hour*24*3).Err()
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, err.Error())
			return
		}
		fmt.Println(accessToken)

		// REFRESH TOKEN
		refreshToken, err := jwtMaker.CreateToken(userData.User.Id, time.Hour*24*90)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, err.Error())
			return
		}
		// STORE TO POSTGRESQL
		err = db.StoreToken(userData.User, refreshToken)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, err.Error())
			return
		}
		fmt.Println(refreshToken)

		userData.AccessToken = accessToken
		userData.RefreshToken = refreshToken
		userData.User.Password = "" // Clear password for security

		response.Success(c, userData)
	}
}

func GoogleCreateHandler(db storage.Storage, jwtMaker *token.JWTMaker, redisClient *redis.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req types.GoogleAccountCreate
		var userData types.AuthResponse
		err := c.ShouldBindJSON(&req)
		if err != nil {
			fmt.Println("failed in validation", err)
			response.Failed(c, http.StatusBadRequest, "Invalid Request")
			return
		}
		// Handle Google user creation
		userData, err = authservice.HandleGoogleUserCreation(req, db)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, userData.Message)
			return
		}

		// TOKENS CREATION
		// ACCESS TOKEN
		accessToken, err := jwtMaker.CreateToken(userData.User.Id, time.Hour*24*3)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, err.Error())
			return
		}
		// STORE TO REDIS
		err = redisClient.Client.Set(ctx, fmt.Sprintf("access_token:%d", userData.User.Id), accessToken, time.Hour*24*3).Err()
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, err.Error())
			return
		}
		fmt.Println(accessToken)

		// REFRESH TOKEN
		refreshToken, err := jwtMaker.CreateToken(userData.User.Id, time.Hour*24*90)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, err.Error())
			return
		}
		// STORE TO POSTGRESQL
		err = db.StoreToken(userData.User, refreshToken)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, err.Error())
			return
		}
		fmt.Println(refreshToken)

		userData.AccessToken = accessToken
		userData.RefreshToken = refreshToken
		userData.User.Password = "" // Clear password for security

		response.Success(c, userData)
	}
}

func EmailCreateHandler(db storage.Storage, jwtMaker *token.JWTMaker, redisClient *redis.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {

		var user types.User
		var authResponse types.AuthResponse
		err := c.ShouldBindJSON(&user)
		if err != nil {
			response.Failed(c, http.StatusBadRequest, "Invalid Request")
			return
		}
		if user.AuthType != "email" {
			response.Failed(c, http.StatusBadRequest, "Invalid Auth Type")
			return
		}

		authResponse, err = authservice.HandleEmailUserCreation(&user, db)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, fmt.Errorf("user creation failed %v", err).Error())
			return
		}
		// TOKENS CREATIONS
		//ACCESS TOKEN
		token, err := jwtMaker.CreateToken(user.Id, time.Hour*24*3)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, "Failed to create access token")
			return
		}
		// STORE TO REDIS
		err = redisClient.Client.Set(ctx, fmt.Sprintf("access_token:%d", user.Id), token, time.Hour*24*3).Err()
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, fmt.Errorf("failed to store access token in redis: %v", err).Error())
			return
		}

		//REFRESH TOKEN
		refreshToken, err := jwtMaker.CreateToken(user.Id, time.Hour*24*90)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, fmt.Errorf("Failed to create refresh token: %v", err).Error())
			return
		}
		// STORE TO POSTGRESQL
		err = db.StoreToken(user, refreshToken)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, fmt.Errorf("Failed to store refresh token in PostgreSQL: %v", err).Error())
			return
		}
		authResponse.IsRegistered = true
		authResponse.User = user
		authResponse.AccessToken = token
		authResponse.RefreshToken = refreshToken
		authResponse.Message = "User created successfully"

		response.Success(c, authResponse)
		// return
	}
}

func CheckUsernameIsAvailable(db storage.Storage, redis *redis.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {

		username := c.Query("username")

		if username == "" {
			response.Failed(c, http.StatusBadRequest, "Invalid Request")
			return
		}

		isAvailable := db.CheckUsernameIsAvailable(username)
		if !isAvailable {
			res := Res{
				IsRegistered: true,
				Message:      "Username is already taken",
			}
			response.Success(c, res)
			return
		}
		res := Res{
			IsRegistered: false,
			Message:      "Username is available",
		}

		response.Success(c, res)
	}
}

func EmailLoginHandler(db storage.Storage, jwtMaker *token.JWTMaker, redisClient *redis.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req EmailLoginRequest

		//validate the request body
		err := c.ShouldBindJSON(&req)
		if err != nil {
			response.Failed(c, http.StatusBadRequest, "Invalid Request")
			return
		}

		// Handle Email login
		authResponse, err := authservice.HandleEmailLogin(req.Email, req.Password, db, redisClient)
		if err != nil {
			response.Failed(c, http.StatusBadRequest, fmt.Errorf("error handling email login: %v", err).Error())
			return
		}

		// TOKENS CREATIONS
		//ACCESS TOKEN
		token, err := jwtMaker.CreateToken(authResponse.User.Id, time.Hour*24*3)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, "Failed to create access token")
			return
		}
		// STORE TO REDIS
		err = redisClient.Client.Set(ctx, fmt.Sprintf("access_token:%d", authResponse.User.Id), token, time.Hour*24*3).Err()
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, fmt.Errorf("failed to store access token in redis: %v", err).Error())
			return
		}

		//REFRESH TOKEN
		refreshToken, err := jwtMaker.CreateToken(authResponse.User.Id, time.Hour*24*90)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, fmt.Errorf("Failed to create refresh token: %v", err).Error())
			return
		}
		// STORE TO POSTGRESQL
		err = db.StoreToken(authResponse.User, refreshToken)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, fmt.Errorf("Failed to store refresh token in PostgreSQL: %v", err).Error())
			return
		}

		authResponse.RefreshToken = refreshToken
		authResponse.AccessToken = token

		response.Success(c, authResponse)
	}
}

func ForgetPasswordHandler(db storage.Storage, redis *redis.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var passwordReset EmailPassWordForget
		err := c.ShouldBindJSON(&passwordReset)
		if err != nil {
			response.Failed(c, http.StatusBadRequest, "Invalid Request")
			return
		}
		if passwordReset.Password != passwordReset.ConfirmPassword {
			response.Failed(c, http.StatusBadRequest, "Passwords do not match")
			return
		}
		err = authservice.HandlePasswordForget(passwordReset.Email, passwordReset.Password, db, redis)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, fmt.Errorf("error handling password reset: %v", err).Error())
			return
		}
		response.Success(c, "Password reset successful, please check your email for the OTP")
	}
}

func ResetPasswordHandler(db storage.Storage, redis *redis.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var passwordReset EmailPassWordReset
		err := c.ShouldBindJSON(&passwordReset)
		if err != nil {
			response.Failed(c, http.StatusBadRequest, "Invalid Request")
			return
		}

		err = authservice.HandlePasswordReset(passwordReset.Email, passwordReset.Otp, db, *redis)
		if err != nil {
			response.Failed(c, http.StatusUnauthorized, err.Error())
			return
		}
		response.Success(c, "Password Reset Successfull Please login now")
	}
}

func UpdateTokenHandler(db storage.Storage, redis *redis.RedisClient, jwtMaker token.JWTMaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token Token
		err := c.ShouldBindJSON(&token)
		if err != nil {
			fmt.Println("error in  UpdateTokenHandler : ", err)
			response.Failed(c, http.StatusBadRequest, "invalid request")
		}
		id, err := authservice.HandleUpdateAccessToken(token.RefreshToken, db, *redis)
		if err != nil {
			response.Failed(c, http.StatusBadRequest, err.Error())
		}

		accessToken, err := jwtMaker.CreateToken(id, time.Hour*24*3)
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, "Failed to create access token")
			return
		}
		// STORE TO REDIS
		err = redis.Client.Set(ctx, fmt.Sprintf("access_token:%d", id), accessToken, time.Hour*24*3).Err()
		if err != nil {
			response.Failed(c, http.StatusInternalServerError, fmt.Errorf("failed to store access token in redis: %v", err).Error())
			return
		}
		response.Success(c, map[string]any{
			"id":          id,
			"accessToken": accessToken,
		})
		return
	}
}
