package types

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
	Id         int64    `json:"id"` // Set by DB
	FullName   string   `json:"full_name" binding:"required"`
	Username   string   `json:"username" binding:"required"`
	Email      string   `json:"email" binding:"required,email"`
	Password   string   `json:"password,omitempty"` // Only used in local auth
	Age        string   `json:"age" binding:"required"`
	Gender     string   `json:"gender" binding:"required"`
	Interests  []string `json:"interests" binding:"required"`
	ProfilePic string   `json:"profile_pic" binding:"required,url"`
	AuthType   string   `json:"auth_type" binding:"required"` // "google" or "email"
	CreatedAt  string   `json:"created_at,omitempty"`         // Set on backend
}

type AuthResponse struct {
	IsRegistered bool   `json:"isRegistered"`
	User         User   `json:"user"`
	Message      string `json:"message"`
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

type GoogleAccountCreate struct {
	IDToken        string   `json:"id_token" binding:"required"`
	Username       string   `json:"username" binding:"required"`
	Interests      []string `json:"interests" binding:"required"`
	Gender         string   `json:"gender" binding:"required"`
	Age            string   `json:"age" binding:"required"`
	NativeLanguage string   `json:"native_language" binding:"required"`
}
