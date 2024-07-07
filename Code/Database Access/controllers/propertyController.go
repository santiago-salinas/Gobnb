package controllers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"pocketbase_go/logger"
	"pocketbase_go/my_models"
	"pocketbase_go/services/interfaces"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type PropertyController struct {
	Service        interfaces.IPropertyService
	PaymentService interfaces.IPaymentService
	AuthService    interfaces.IAuthService
}

func (controller *PropertyController) InitPropertyEndpoints(app core.App) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.POST("/property/img/:id", func(c echo.Context) error {
			id := c.PathParam("id")
			token := c.Request().Header.Get("auth")

			file, err := c.FormFile("file")
			if err != nil {
				logger.Error("Failed to read file", err)
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "Failed to read file"})
			}
			//Check file format .jpg .jpeg .png
			//Check file size < 500kb
			if file.Size > 500000 {
				logger.Error("File size is too big")
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "File size is too big"})
			}
			fileExtension := file.Filename[strings.LastIndex(file.Filename, "."):]
			if fileExtension != ".jpg" && fileExtension != ".jpeg" && fileExtension != ".png" {
				logger.Error("File format is not supported")
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "File format is not supported"})
			}

			fileData, err := file.Open()
			if err != nil {
				logger.Error("Failed to open file", err)
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "Failed to open file"})
			}

			err = controller.PostImage(id, fileData, fileExtension, token)
			if err != nil {
				logger.Error("Failed to post image", err)
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "Failed to post image"})
			}
			return c.JSON(http.StatusCreated, map[string]string{"message": "Success"})
		})

		e.Router.POST("/property/:id/addUnavailableDate", func(c echo.Context) error {
			id := c.PathParam("id")
			token := c.Request().Header.Get("auth")

			var req []my_models.DateRange

			if err := c.Bind(&req); err != nil {
				logger.Error("Failed to read request data", err)
				return apis.NewBadRequestError("Failed to read request data", err)
			}
			logger.Info("Request: ", req)
			response := controller.AddUnavailableDates(id, req, token)
			if response != nil {
				logger.Error(response.Error())
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": response.Error()})
			}
			logger.Info("Success adding unavailable date", req)
			return c.JSON(http.StatusCreated, map[string]string{"message": "Success"})
		})

		e.Router.DELETE("/property/:id/removeUnavailableDate", func(c echo.Context) error {
			id := c.PathParam("id")
			token := c.Request().Header.Get("auth")

			var req my_models.DateRange

			if err := c.Bind(&req); err != nil {
				logger.Error("Failed to read request data", err)
				return apis.NewBadRequestError("Failed to read request data", err)
			}
			logger.Info("Request: ", req)
			response := controller.Service.RemoveUnavailableDates(id, req, token)
			if response != nil {
				logger.Error(response.Error())
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": response.Error()})
			}
			logger.Info("Success removing unavailable date", req)
			return c.JSON(http.StatusCreated, map[string]string{"message": "Success"})
		})

		e.Router.POST("/property", func(c echo.Context) error {
			token := c.Request().Header.Get("auth")

			var req my_models.Property

			if err := c.Bind(&req); err != nil {
				logger.Error("Failed to read request data", err)
				return apis.NewBadRequestError("Failed to read request data", err)
			}
			logger.Info("Request: %v\n", req)
			propertyId, err := controller.PostProperty(req, token)
			if err != nil {
				logger.Error(err.Error())
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			logger.Info("Success creating property", req)
			return c.JSON(http.StatusCreated, map[string]string{"message": "Success with id: " + propertyId})
		})

		e.Router.GET("/property", func(c echo.Context) error {
			var req my_models.PropertyFilter
			token := c.Request().Header.Get("auth")

			logger.Info("Received request to get inmueble with query parameters:", c.QueryParams())

			if pageQuery := c.QueryParam("page"); pageQuery != "" {
				page, err := strconv.Atoi(pageQuery)
				if err != nil || page < 1 {
					logger.Error("Page is not valid, pages start from 1")
					return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "Page is not valid, pages start from 1"})
				}
				req.Page = &page
			} else {
				page := 1
				req.Page = &page
			}

			if sizeQuery := c.QueryParam("size"); sizeQuery != "" {
				size, err := strconv.Atoi(sizeQuery)
				if err != nil || size < 1 {
					logger.Error("Size is not valid, sizes start from 1")
					return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "Size is not valid, sizes start from 1"})
				}
				req.Size = &size
			} else {
				size := 5
				req.Size = &size
			}

			// Helper function to parse and validate number parameters
			parseNumber := func(param string) (*int, error) {
				num, err := strconv.Atoi(param)
				if err != nil || num < 0 || num > 99 {
					logger.Error("Value must be between 0 and 99")
					return nil, fmt.Errorf("value must be between 0 and 99")
				}
				return &num, nil
			}

			if val := c.QueryParam("adultQuantityMax"); val != "" {
				if num, err := parseNumber(val); err != nil {
					logger.Error("AdultQuantityMax", err)
					return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "AdultQuantityMax " + err.Error()})
				} else {
					req.AdultQuantityMax = num
				}
			}
			if val := c.QueryParam("adultQuantityMin"); val != "" {
				if num, err := parseNumber(val); err != nil {
					logger.Error("AdultQuantityMin", err)
					return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "AdultQuantityMin " + err.Error()})
				} else {
					req.AdultQuantityMin = num
				}
			}
			if val := c.QueryParam("kingSizedBedsMax"); val != "" {
				if num, err := parseNumber(val); err != nil {
					logger.Error("KingSizedBedsMax", err)
					return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "KingSizedBedsMax " + err.Error()})
				} else {
					req.KingSizedBedsMax = num
				}
			}
			if val := c.QueryParam("kingSizedBedsMin"); val != "" {
				if num, err := parseNumber(val); err != nil {
					logger.Error("KingSizedBedsMin", err)
					return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "KingSizedBedsMin " + err.Error()})
				} else {
					req.KingSizedBedsMin = num
				}
			}
			if val := c.QueryParam("singleBedsMax"); val != "" {
				if num, err := parseNumber(val); err != nil {
					logger.Error("SingleBedsMax", err)
					return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "SingleBedsMax " + err.Error()})
				} else {
					req.SingleBedsMax = num
				}
			}
			if val := c.QueryParam("singleBedsMin"); val != "" {
				if num, err := parseNumber(val); err != nil {
					logger.Error("SingleBedsMin", err)
					return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "SingleBedsMin " + err.Error()})
				} else {
					req.SingleBedsMin = num
				}
			}
			if val := c.QueryParam("hasAC"); val != "" {
				hasAC, err := strconv.ParseBool(val)
				if err != nil {
					logger.Error("HasAC", err)
					return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "Invalid value for hasAC"})
				}
				req.HasAC = &hasAC
			}
			if val := c.QueryParam("hasWIFI"); val != "" {
				hasWIFI, err := strconv.ParseBool(val)
				if err != nil {
					logger.Error("HasWIFI", err)
					return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "Invalid value for hasWIFI"})
				}
				req.HasWIFI = &hasWIFI
			}
			if val := c.QueryParam("hasGarage"); val != "" {
				hasGarage, err := strconv.ParseBool(val)
				if err != nil {
					logger.Error("HasGarage", err)
					return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "Invalid value for hasGarage"})
				}
				req.HasGarage = &hasGarage
			}
			if val := c.QueryParam("type"); val != "" {
				typea, err := strconv.Atoi(val)
				if err != nil || (typea != 1 && typea != 2) {
					logger.Error("Type must be either 1 or 2")
					return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "Type must be either 1 or 2"})
				}
				req.Type = &typea
			}
			if val := c.QueryParam("beachDistanceMax"); val != "" {
				if num, err := parseNumber(val); err != nil {
					logger.Error("BeachDistanceMax", err)
					return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "BeachDistanceMax " + err.Error()})
				} else {
					req.BeachDistanceMax = num
				}
			}
			if val := c.QueryParam("beachDistanceMin"); val != "" {
				if num, err := parseNumber(val); err != nil {
					logger.Error("BeachDistanceMin", err)
					return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "BeachDistanceMin " + err.Error()})
				} else {
					req.BeachDistanceMin = num
				}
			}
			if state := c.QueryParam("state"); state != "" {
				req.State = &state
			}
			if resort := c.QueryParam("resort"); resort != "" {
				req.Resort = &resort
			}
			if neighborhood := c.QueryParam("neighborhood"); neighborhood != "" {
				req.Neighborhood = &neighborhood
			}
			if dateFrom := c.QueryParam("dateFrom"); dateFrom != "" {
				if dateTo := c.QueryParam("dateTo"); dateTo != "" {
					// Parse the dates
					startDate, err := time.Parse("2006-01-02", dateFrom)
					if err != nil {
						logger.Error("Invalid dateFrom format, expected YYYY-MM-DD")
						return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "Invalid dateFrom format, expected YYYY-MM-DD"})
					}
					endDate, err := time.Parse("2006-01-02", dateTo)
					if err != nil {
						logger.Error("Invalid dateTo format, expected YYYY-MM-DD")
						return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "Invalid dateTo format, expected YYYY-MM-DD"})
					}

					// Check if dateFrom is later than dateTo
					if startDate.After(endDate) {
						logger.Error("dateFrom cannot be later than dateTo")
						return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "dateFrom cannot be later than dateTo"})
					}

					req.DateFrom = &dateFrom
					req.DateTo = &dateTo
				} else {
					logger.Error("dateTo is required when dateFrom is provided")
					return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "dateTo is required when dateFrom is provided"})
				}
			}

			response, err := controller.GetFilteredProperties(token, req)
			if err != nil {
				logger.Error(err.Error())
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			return c.JSON(http.StatusOK, response)
		})

		e.Router.POST("/property/pay", func(c echo.Context) error {
			type PayBody struct {
				PropertyId string `json:"propertyId"`
				CardInfo   my_models.CardInformation
			}

			var req PayBody

			if err := c.Bind(&req); err != nil {
				return apis.NewBadRequestError("Failed to read request data", err)
			}
			response := controller.PayProperty(req.PropertyId, req.CardInfo)
			if response != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": response.Error()})
			}
			return c.JSON(http.StatusCreated, map[string]string{"message": "Success"})
		})

		return nil
	})
}

func (c *PropertyController) GetFilteredProperties(token string, filter my_models.PropertyFilter) ([]my_models.Property, error) {
	logger.Info("Controller: Getting filtered properties")
	roles, _, err := c.AuthService.Login(token)
	if err != nil {
		logger.Error("Controller: Error in GetFilteredProperties: ", err)
		return nil, err
	}

	for _, role := range roles {
		if role == "Tenant" {
			properties, err := c.Service.GetFilteredProperties(filter)
			if err != nil {
				logger.Error("Controller: Error in GetFilteredProperties: ", err)
				return nil, err
			}
			logger.Info("Controller: Got filtered properties succesfully")
			return properties, nil
		}
	}

	err = fmt.Errorf("Controller: User is not an admin")
	logger.Error("Controller: Error in GetFilteredProperties: User is not a Tenant")
	return nil, err

}

func (c *PropertyController) PostImage(id string, image multipart.File, fileExtension string, userToken string) error {
	logger.Info("Controller: Adding image to property with id: ", id)
	err := c.Service.AddPropertyImage(id, image, fileExtension, userToken)
	if err != nil {
		return err
	}

	logger.Info("Controller: Image added to property with id: ", id)
	return nil
}

func (c *PropertyController) PostProperty(property my_models.Property, userToken string) (string, error) {
	logger.Info("Controller: Adding property")
	propertyId, err := c.Service.AddProperty(property, userToken)
	if err != nil {
		return "", err
	}

	logger.Info("Controller: Property added")
	return propertyId, nil
}

func (c *PropertyController) AddUnavailableDates(propertyId string, dates []my_models.DateRange, userToken string) error {
	logger.Info("Controller: Adding unavailable dates to property with id: %d", propertyId)
	err := c.Service.AddUnavailableDates(propertyId, dates, userToken)
	if err != nil {
		return err
	}
	return nil
}

func (c *PropertyController) PayProperty(propertyId string, cardInformation my_models.CardInformation) error {
	logger.Info("Controller: Paying for property with id: ", propertyId)
	err := c.PaymentService.PayProperty(propertyId, cardInformation)
	if err != nil {
		return err
	}

	logger.Info("Controller: Payment successful")
	return nil
}
