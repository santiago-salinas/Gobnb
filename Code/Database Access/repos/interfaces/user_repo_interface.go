package repointerfaces

import (
	"pocketbase_go/my_models"
)

type IUserRepo interface {
	AddUser(token string) error
	Login(token string) ([]string, string, error)
	GetUsersByRole(role string) ([]string, error)
	GetPropertyOwner(propertyId string) (string, error)
	GetUserById(userId string) (my_models.User, error)
}
