package authservice

import (
	"context"
	"fmt"
	"github/english-app/internal/types"
	"github/english-app/storage"
	"os"
	"time"

	"google.golang.org/api/idtoken"
)

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
			IsRegistered: isAvailable,
			User:         types.User{},
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
		return AuthResponse, err
	}
	if isAvailable {
		return types.AuthResponse{
			IsRegistered: true,
			User:         types.User{},
			Message:      "User already exists, please login",
		}, nil
	}
	user := types.User{
		FullName:   payload.Claims["name"].(string),
		Username:   body.Username,
		Gender:     body.Gender,
		ProfilePic: payload.Claims["picture"].(string),
		Email:      payload.Claims["email"].(string),
		Interests:  body.Interests,
		Age:        body.Age,
		CreatedAt:  time.Now(),
		AuthType:   "google", // Assuming the auth type is google
		Password:   "",       // Password is not required for Google auth
	}
	// It should save the user information in the database.
	err = db.SaveUserInDatabase(user)
	if err != nil {
		return AuthResponse, err
	}

	AuthResponse = types.AuthResponse{
		IsRegistered: true,
		User:         user,
		Message:      "User created successfully",
	}
	return AuthResponse, nil
}

func HandleEmailUserCreation(user types.User, db storage.Storage) (types.AuthResponse, error) {
	var AuthResponse types.AuthResponse
	//check already in db
	isAlreadyAvailable, _, _ := db.CheckUserInDatabase(user.Email)
	if isAlreadyAvailable {
		return AuthResponse, fmt.Errorf("user already exists")
	}
	//store in db

	err := db.SaveUserInDatabase(user)
	if err != nil {
		return AuthResponse, fmt.Errorf("error saving user in database")
	}
	return AuthResponse, nil
}
