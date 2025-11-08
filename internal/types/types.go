package types

import (
	"github.com/jackc/pgx/v5/pgtype"
)

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
	Otp       string             `json:"otp,omitempty"`
	Pictures  []string           `json:"pictures,omitempty"`
	Is_active bool               `json:"is_active,omitempty"`
}

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

type CallHistory struct {
	CallId        string `json:"call_id"`
	PeerID1       int64  `json:"peer1_id"`
	PeerID2       int64  `json:"peer2_id"`
	PeerName1     string `json:"peer1_name"`
	PeerName2     string `json:"peer2_name"`
	PeerPic1      string `json:"peer1_pic"`
	PeerPic2      string `json:"peer2_pic"`
	CallStart     string `json:"started_at"`
	CallEnd       string `json:"ended_at"`
	Status        string `json:"status"`
	DurationInMin string `json:"duration_in_min"`
}

type CallRecord struct {
	CallId        string `json:"call_id"`
	PeerId        int64  `json:"peer_id"`
	PeerName      string `json:"peer_name"`
	PeerPic       string `json:"peer_pic"`
	CallStart     string `json:"started_at"`
	CallEnd       string `json:"ended_at"`
	Status        string `json:"status"`
	DurationInMin string `json:"duration_in_min"`
}

type LeaderboardEntry struct {
	UserData      User    `json:"user_data"`
	PeriodType    string  `json:"period_type"`
	PeriodStart   string  `json:"period_start"`
	TotalDuration float64 `json:"total_duration"`
	UpdatedAt     string  `json:"updated_at"`
	Rank          int     `json:"rank"`
}

type AiCharacters struct {
	Id              int8   `json:"id"`
	Name            string `json:"name"`
	PicUrl          string `json:"picUrl"`
	Description     string `json:"description"`
	BackgroundColor string `json:"backgroundColor"`
}
