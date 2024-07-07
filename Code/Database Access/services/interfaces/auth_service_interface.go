package interfaces

import (
	"pocketbase_go/my_models"
)

type IAuthService interface {
	Login(userToken string) (roles []string, id string, err error)
	AddUser(token string) error
	GetUserById(id string) (my_models.User, error)
}
