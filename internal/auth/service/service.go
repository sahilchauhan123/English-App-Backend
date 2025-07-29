package authservice

import (
	"context"
	"encoding/json"
	"fmt"
	"github/english-app/internal/smtp"
	"github/english-app/internal/types"
	"github/english-app/storage"
	"github/english-app/storage/redis"
	"math/rand"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

type EmailPassWordReset struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	OTP      string `json:"otp" binding:"required"`
}
type OtpRedisResponse struct {
	Email string `json:"email" binding:"required"`
	OTP   string `json:"otp" binding:"required"`
}

func generateRandomNumber() string {
	// Generate a random number between 1000 and 9999 (inclusive)
	num := rand.Intn(9000) + 1000
	return strconv.Itoa(num)
}

func validateIdToken(IDToken string) (*idtoken.Payload, error) {
	payload, err := idtoken.Validate(context.Background(), IDToken, os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		fmt.Println("Error validating ID token:", err)
		return nil, err
	}
	fmt.Println("payload : ", payload)
	return payload, nil
}

func HandleGoogleLogin(IDToken string, db storage.Storage) (types.AuthResponse, error) {
	// This function should handle the Google login logic.
	// It should verify the ID token, fetch user information,
	var AuthResponse types.AuthResponse
	payload, err := validateIdToken(IDToken)
	if err != nil {
		return AuthResponse, nil
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		return AuthResponse, nil
	}
	// Check user in Database
	isAvailable, user, err := db.CheckUserInDatabase(email)
	if err != nil {
		return AuthResponse, nil
	}
	fmt.Println("user available in database:", isAvailable, "user:", user)

	// If user is not available, return isAvailable as false
	if !isAvailable {
		AuthResponse = types.AuthResponse{
			IsRegistered: false,
			User:         user,
			Message:      "User not found in database, please register",
		}
		return AuthResponse, nil
	}

	return types.AuthResponse{
		IsRegistered: isAvailable,
		User:         user,
		Message:      "User found in database, login successful",
	}, nil
}

func HandleGoogleUserCreation(body types.GoogleAccountCreate, db storage.Storage) (types.AuthResponse, error) {
	// This function should handle the Google user creation logic.
	var AuthResponse types.AuthResponse

	payload, err := validateIdToken(body.IDToken)
	if err != nil {
		return types.AuthResponse{}, err
	}

	//check user should not already exist in the database
	isAvailable, _, err := db.CheckUserInDatabase(payload.Claims["email"].(string))

	if err != nil {
		fmt.Println("runing errrror")
		// return AuthResponse, err
	}

	if isAvailable {
		fmt.Println("User already exists")
		return types.AuthResponse{
			IsRegistered: true,
			User:         types.User{},
			Message:      "User already exists, please login",
		}, fmt.Errorf("User already exists, please login")
	}
	user := types.User{
		FullName:            payload.Claims["name"].(string),
		Username:            body.Username,
		Gender:              body.Gender,
		ProfilePic:          payload.Claims["picture"].(string),
		Email:               payload.Claims["email"].(string),
		Age:                 body.Age,
		MainChallenge:       body.MainChallenge,
		CurrentEnglishLevel: body.CurrentEnglishLevel,
		NativeLanguage:      body.NativeLanguage,
		// CreatedAt:  time.Now(),
		AuthType: "google", // Assuming the auth type is google
		Password: "",       // Password is not required for Google auth
	}
	fmt.Println("user data above : ", user)
	// It should save the user information in the database.
	err = db.SaveUserInDatabase(&user)
	if err != nil {
		return AuthResponse, err
	}
	fmt.Println("user data down : ", user)
	AuthResponse = types.AuthResponse{
		IsRegistered: true,
		User:         user,
		Message:      "User created successfully",
	}
	return AuthResponse, nil
}

func HandleEmailUserCreation(user *types.User, db storage.Storage, redis *redis.RedisClient) (types.AuthResponse, error) {
	var AuthResponse types.AuthResponse
	var OtpRedis OtpRedisResponse

	//check already in db
	isUserNameAvaibale := db.CheckUsernameIsAvailable(user.Username)
	if !isUserNameAvaibale {
		return AuthResponse, fmt.Errorf("username already exists")
	}

	isAlreadyAvailable, _, _ := db.CheckUserInDatabase(user.Email)
	if isAlreadyAvailable {
		return AuthResponse, fmt.Errorf("user already exists")
	}

	// check otp is matching or not
	key := fmt.Sprintf("otp_signup:%s", user.Email)
	data, err := redis.Client.Get(context.Background(), key).Result()
	if err != nil {
		fmt.Println("Error fetching OTP from Redis:", err)
		return AuthResponse, err
	}
	err = json.Unmarshal([]byte(data), &OtpRedis)
	if err != nil {
		AuthResponse.Message = "error unmarshalling OTP data"
		return AuthResponse, fmt.Errorf("error unmarshalling OTP data: %v", err)
	}
	fmt.Println("OtpRedis data : ", OtpRedis, "user otp : ", user.Otp)
	if user.Otp != OtpRedis.OTP {
		AuthResponse.Message = "otp is invalid"
		return AuthResponse, fmt.Errorf("otp is invalid")
	}
	//store in db
	err = db.SaveUserInDatabase(user)
	if err != nil {
		return AuthResponse, fmt.Errorf("error saving user in database")
	}
	return AuthResponse, nil
}

func HandleEmailSignUpOtp(email string, db storage.Storage, redisclient *redis.RedisClient) (bool, error) {

	exists, _, err := db.CheckUserInDatabase(email)
	if err != nil {
		fmt.Println("Error checking user in database:", err)
		return false, err
	}
	if exists {
		return false, nil
	}

	key := fmt.Sprintf("otp_signup:%s", email)
	otp := generateRandomNumber()
	value := map[string]any{
		"email": email,
		"otp":   otp,
	}
	valueJSON, err := json.Marshal(value)
	if err != nil {
		fmt.Println("Error marshalling OTP data:", err)
		return false, err
	}
	_, err = redisclient.Client.Set(context.Background(), key, valueJSON, time.Minute*5).Result()

	if err != nil {
		fmt.Println("Error setting OTP in Redis:", err)
		return false, err
	}
	err = smtp.SendEmailOTP(email, otp)
	if err != nil {
		return false, err
	}
	return false, nil
}

func HandleEmailLogin(email string, otp string, db storage.Storage, redisclient *redis.RedisClient) (types.AuthResponse, error) {

	var AuthResponse types.AuthResponse
	var OtpRedis OtpRedisResponse
	//check user in db
	isAvailable, user, err := db.CheckUserInDatabase(email)
	fmt.Println("userdata : ", user)

	if err != nil {
		AuthResponse.Message = "Error checking user in database"
		return AuthResponse, fmt.Errorf("error checking user in database: %v", err)
	}
	if !isAvailable {
		AuthResponse.Message = "User not found in database, please register"
		return AuthResponse, fmt.Errorf("user not found in database")
	}
	//check AuthType
	if user.AuthType != "email" {
		AuthResponse.Message = "User is not registered with email authentication"
		return AuthResponse, fmt.Errorf("user is not registered with email authentication")
	}

	// fetch otp from redis if not found
	key := fmt.Sprintf("otp_login:%s", email)

	data, err := redisclient.Client.Get(context.Background(), key).Result()
	if err != nil {
		return AuthResponse, err
	}

	err = json.Unmarshal([]byte(data), &OtpRedis)
	if err != nil {
		AuthResponse.Message = "error unmarshalling OTP data"
		return AuthResponse, fmt.Errorf("error unmarshalling OTP data: %v", err)
	}

	if OtpRedis.OTP != otp {
		AuthResponse.Message = "Invalid Otp"
		return AuthResponse, fmt.Errorf("invalid Otp")
	}

	// delete the Refresh token
	_ = db.DeleteToken(user.Id)
	// delete the Access token
	redisclient.Client.Del(context.Background(), fmt.Sprintf("access_token:%d", user.Id))

	AuthResponse = types.AuthResponse{
		IsRegistered: true,
		User:         user,
		Message:      "Login successful",
	}
	return AuthResponse, nil

}

// func matchHashedPasswords(hashedPassWord, password string) bool {
// 	err := bcrypt.CompareHashAndPassword([]byte(hashedPassWord), []byte(password))
// 	if err != nil {
// 		return false // Passwords do not match
// 	}

// 	return true
// }

func HandleEmailLoginOTPGenerate(db storage.Storage, email string, redis redis.RedisClient) (bool, error) {
	exist, userData, err := db.CheckUserInDatabase(email)

	if err != nil {
		return false, err
	}
	if !exist {
		return false, nil
	}
	if userData.AuthType == "google" {
		return false, fmt.Errorf("Please Login via Google.")
	}

	otp := generateRandomNumber()
	key := fmt.Sprintf("otp_login:%s", email)

	// storing in redis
	redis.Client.Set(context.Background(), key, map[string]any{
		"email": email,
		"opt":   otp,
	}, time.Minute*5)

	//send otp to user email
	err = smtp.SendEmailOTP(email, otp)
	if err != nil {
		return false, err
	}

	return true, nil

}
func HandlePasswordForget(email string, password string, db storage.Storage, redis *redis.RedisClient) error {
	// check user exists or not
	// if exists then check auth type should be email
	isAvailable, userData, err := db.CheckUserInDatabase(email)

	if !isAvailable {
		return fmt.Errorf("user not found in database")
	}
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	if userData.AuthType != "email" {
		return fmt.Errorf("user is not registered with email authentication")
	}
	handleOTPStore(email, password, redis)

	return nil
}

func handleOTPStore(email string, password string, redis *redis.RedisClient) error {
	// This function should handle the OTP storage logic.
	// It should store the OTP in the database or any other storage.
	// For now, we will just return nil to indicate success.
	// In a real application, you would implement the actual logic here.

	var EmailPassWordReset EmailPassWordReset
	EmailPassWordReset.Email = email
	EmailPassWordReset.Password = password
	EmailPassWordReset.OTP = generateRandomNumber() // Generate a random 6-digit OTP
	data, err := json.Marshal(EmailPassWordReset)
	if err != nil {
		return err
	}
	redis.Client.Set(context.Background(), fmt.Sprintf("reset_password:%s", email), data, time.Minute*5)
	err = smtp.SendEmailOTP(EmailPassWordReset.Email, EmailPassWordReset.OTP)

	if err != nil {
		fmt.Println("Error sending email OTP:", err)
		return err
	}

	return nil

}

func HandlePasswordReset(email string, otp string, db storage.Storage, redis redis.RedisClient) error {
	// check email in redis if available fetch its data from redis
	// check Otp should be same
	var otpData EmailPassWordReset
	val, err := redis.Client.Get(context.Background(), fmt.Sprintf("reset_password:%s", email)).Result()
	fmt.Println("value coming from redis : ", val)
	if err != nil {
		return fmt.Errorf("OTP not found or expired %e", err)
	}
	err = json.Unmarshal([]byte(val), &otpData)
	if err != nil {
		return fmt.Errorf("error unmarshalling OTP data: %v", err)
	}

	if otpData.OTP != otp {
		return fmt.Errorf("invalid OTP")
	}

	err = db.ChangePassword(email, otpData.Password)
	if err != nil {
		return fmt.Errorf("error changing password: %v", err)
	}
	return nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func HandleUpdateAccessToken(token string, db storage.Storage, redis redis.RedisClient) (int64, error) {
	//check refresh token in sqldb
	isPresent, id := db.CheckToken(token)
	if !isPresent {
		//Refresh Token expired
		return 0, fmt.Errorf("Refresh Token expired")
	}
	//create t

	return id, nil
}
