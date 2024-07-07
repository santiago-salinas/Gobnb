package mocks

import (
	"pocketbase_go/my_models"
)

type MockAuthService struct {
	LoginFunc       func(userToken string) (roles []string, id string, err error)
	AddUserFunc     func(token string) error
	GetUserByIdFunc func(id string) (my_models.User, error)
}

func (m MockAuthService) Login(userToken string) (roles []string, id string, err error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(userToken)
	}
	return nil, "", nil
}

func (m MockAuthService) AddUser(token string) error {
	if m.AddUserFunc != nil {
		return m.AddUserFunc(token)
	}
	return nil
}

func (m MockAuthService) GetUserById(id string) (my_models.User, error) {
	if m.GetUserByIdFunc != nil {
		return m.GetUserByIdFunc(id)
	}
	return my_models.User{}, nil
}
