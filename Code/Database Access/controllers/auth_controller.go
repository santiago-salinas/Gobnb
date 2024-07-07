package controllers

import (
	"net/http"
	"pocketbase_go/logger"
	"pocketbase_go/services/interfaces"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/core"
)

type AuthController struct {
	AuthService interfaces.IAuthService
}

func (controller *AuthController) InitAuthEndpoints(app core.App) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.POST("/login", func(c echo.Context) error {
			token := c.Request().Header.Get("auth")

			roles, _, err := controller.Login(token)
			if err != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			return c.JSON(http.StatusOK, roles)
		})

		e.Router.POST("/user", func(c echo.Context) error {
			token := c.Request().Header.Get("auth")

			err := controller.AddUser(token)
			if err != nil && err.Error() == "error: user already exists" {
				return c.JSON(http.StatusOK, map[string]string{"message": err.Error()})
			}

			if err != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			return c.JSON(http.StatusCreated, map[string]string{"message": "Success"})
		})

		return nil
	})
}

func (c *AuthController) Login(userToken string) ([]string, string, error) {
	logger.Info("Controller: Logging in user with token: ", userToken)
	roles, userId, err := c.AuthService.Login(userToken)
	if err != nil {
		return nil, "", err
	}

	logger.Info("Controller: User logged in successfully")
	return roles, userId, nil
}

func (c *AuthController) AddUser(token string) error {
	logger.Info("Controller: Adding user with token: ", token)
	err := c.AuthService.AddUser(token)
	if err != nil {
		return err
	}

	logger.Info("Controller: User added successfully")
	return nil
}
