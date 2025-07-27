package types

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// type GoogleUser struct {
// 	FullName   string    `json:"full_name" binding:"required"`
// 	Username   string    `json:"username" binding:"required"`
// 	Id         int64     `json:"id"` // usually not required from client
// 	Email      string    `json:"email" binding:"required,email"`
// 	Age        string    `json:"age" binding:"required"`
// 	Gender     string    `json:"gender" binding:"required"`
// 	Interests  []string  `json:"interests" binding:"required"`
// 	CreatedAt  time.Time `json:"created_at"`                         // usually generated server-side
// 	ProfilePic string    `json:"profile_pic" binding:"required,url"` // URL to the profile picture
// 	AuthType   string    `json:"auth_type" binding:"required"`       // "google" or "email"
// 	Password   string    `json:"password,omitempty"`                 // Optional, used for local auth
// }

type User struct {
	Id       int64  `json:"id"` // Set by DB
	FullName string `json:"full_name" binding:"required"`
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password,omitempty"` // Only used in local auth
	Age      int    `json:"age" binding:"required"`
	Gender   string `json:"gender" binding:"required"`
	// Interests      []string `json:"interests" binding:"required"`
	ProfilePic          string `json:"profile_pic" binding:"required,url"`
	AuthType            string `json:"auth_type" binding:"required"` // "google" or "email"
	MainChallenge       string `json:"mainChallenge" binding:"required"`
	NativeLanguage      string `json:"nativeLanguage" binding:"required"`
	CurrentEnglishLevel string `json:"currentEnglishLevel" binding:"required"`
	// CreatedAt  string   `json:"created_at,omitempty"`         // Set on backend
	CreatedAt pgtype.Timestamptz `json:"created_at,omitempty"` // <-- CHANGE THIS LINE
	Otp       string             `json:"opt,omitempty"`
}

// gender: '',   				DONE
// nativeLanguage:'',			DONE
// currentEnglishLevel:'',  	DONE
// age:'',						DONE
// mainChallenge:'',			DONE
// username:''					DONE

type AuthResponse struct {
	IsRegistered bool   `json:"isRegistered"`
	User         User   `json:"user"`
	Message      string `json:"message"`
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

type GoogleAccountCreate struct {
	IDToken             string `json:"id_token" binding:"required"`
	Username            string `json:"username" binding:"required"`
	Gender              string `json:"gender" binding:"required"`
	Age                 int    `json:"age" binding:"required"`
	NativeLanguage      string `json:"nativeLanguage" binding:"required"`
	CurrentEnglishLevel string `json:"currentEnglishLevel" binding:"required"`
	MainChallenge       string `json:"mainChallenge" binding:"required"`
}
