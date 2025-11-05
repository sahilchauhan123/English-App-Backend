package storage

import (
	"github/english-app/internal/types"
	"time"
)

type Storage interface {
	CheckUserInDatabase(email string) (bool, types.User, error)
	SaveUserInDatabase(user *types.User) error
	StoreToken(user types.User, token string) error
	CheckUsernameIsAvailable(username string) bool
	DeleteToken(userId int64) error
	ChangePassword(email string, newPassword string) error
	StartCall(peer1, peer2 int64) (string, error)
	CheckToken(token string) (bool, int64)
	EndCall(id string) error
	InsertPicture(id int64, imageUrl string) error
	CheckPictureLength(id int64) (int, error)
	GetProfile(id int64) (types.User, error)
	GetCallHistory(id int64, timestamp time.Time) ([]types.CallHistory, error)
	DeleteAccount(id int64) error
	DeletePicture(userId int64, imageUrl string) error
	BlockUser(userId int64, blockUserId int64) error
	GetLeaderboard(periodType string) ([]types.LeaderboardEntry, error)
	UpdateLeaderboard(userID int64, duration float64) error
}
