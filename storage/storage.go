package storage

import "github/english-app/internal/types"

type Storage interface {
	CheckUserInDatabase(userID string) (bool, types.User, error)
	SaveUserInDatabase(user types.User) error
	StoreToken(user types.User, token string) error
}
