package services

import (
	"pocketbase_go/logger"
	"pocketbase_go/my_models"
	interfaces "pocketbase_go/repos/interfaces"
)

type AuthService struct {
	Repo interfaces.IUserRepo
}

func (s AuthService) Login(userToken string) (roles []string, id string, err error) {
	roles, userId, err := s.Repo.Login(userToken)
	if err != nil {
		return nil, "", err
	}

	logger.Info("Service: User logged in successfully")
	return roles, userId, nil
}

func (s AuthService) AddUser(token string) error {
	err := s.Repo.AddUser(token)
	if err != nil {
		return err
	}

	logger.Info("Service: User added successfully")
	return nil
}

func (s AuthService) GetUserById(id string) (my_models.User, error) {
	logger.Info("Service: Getting user by id: ", id)
	return s.Repo.GetUserById(id)
}
