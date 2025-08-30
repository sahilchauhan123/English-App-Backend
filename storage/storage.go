package storage

import "github/english-app/internal/types"

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
}
