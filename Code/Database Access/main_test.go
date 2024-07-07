package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pocketbase_go/controllers"
	"pocketbase_go/my_models"
	repositories "pocketbase_go/repos/implementations"
	"pocketbase_go/services"
	"pocketbase_go/services/mocks"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/tests"
)

const testDataDir = "./test_pb_data"
const adminToken = "valid_token"

var testSensor = my_models.Sensor{
	Id:                  "zxnnlmaps6adf0v",
	Description:         "Test description",
	SerialNumber:        "Lorem ipsum dolor sit amet, consectetur.12345",
	Address:             "Test address",
	Brand:               "Test brand",
	LastMaintenanceDate: "2020-10-31",
	ServiceType:         "Testing",
	AssignedTo:          "676",
	ReportStructure: my_models.ReportStructure{
		SensorId: "zxnnlmaps6adf0v",
		MeasureStructures: []my_models.Item{
			{Type: "Fire", Value: "^(Alert)$", Unit: ""},
		}}}

func TestPostSensor(t *testing.T) {
	initLogger()

	sensor := my_models.Sensor{
		Id:                  "2",
		Description:         "test sensor",
		Brand:               "test brand",
		Address:             "test address",
		LastMaintenanceDate: "2021-01-01",
		ServiceType:         "test service type",
		AssignedTo:          "676",
		ReportStructure:     my_models.ReportStructure{SensorId: "2", MeasureStructures: []my_models.Item{{Type: "Fire", Value: "^(Alert)$", Unit: ""}}},
	}
	sensorJSON, err := json.Marshal(sensor)
	if err != nil {
		t.Fatal(err)
	}

	setupTestApp := func(t *testing.T) *tests.TestApp {
		testApp, err := tests.NewTestApp(testDataDir)
		if err != nil {
			t.Fatal(err)
		}

		sensorRepo := repositories.PocketSensorRepo{Db: testApp, Cache: nil}

		sensorService := services.SensorService{Repo: &sensorRepo}
		authService := mocks.MockAuthService{LoginFunc: func(token string) ([]string, string, error) {
			if token == adminToken {
				return []string{"Admin"}, "admin", nil
			}
			return nil, "", nil
		}}

		sensorController := controllers.SensorController{Service: &sensorService, AuthService: authService}
		sensorController.InitSensorEndpoints(testApp)

		return testApp
	}

	scenarios := []tests.ApiScenario{
		{
			Name:            "try with different http method, eg. GET",
			Method:          http.MethodGet,
			Url:             "/sensor",
			ExpectedStatus:  405,
			ExpectedContent: []string{"\"data\":{}"},
			Body:            strings.NewReader(string(sensorJSON)),
			TestAppFactory:  setupTestApp,
		},
		{
			Name:            "try as guest (aka. no Authorization header)",
			Method:          http.MethodPost,
			Url:             "/sensor",
			ExpectedStatus:  406,
			ExpectedContent: []string{"Controller: User is not an admin"},
			Body:            strings.NewReader(string(sensorJSON)),
			TestAppFactory:  setupTestApp,
			ExpectedEvents: map[string]int{
				"OnBeforeApiError": 0,
				"OnAfterApiError":  0,
			},
		},
		{
			Name:   "try as authenticated app user",
			Method: http.MethodPost,
			Url:    "/sensor",
			RequestHeaders: map[string]string{
				"auth":
				// Fill in token
				"",
			},
			ExpectedStatus:  406,
			ExpectedContent: []string{"Controller: User is not an admin"},
			Body:            strings.NewReader(string(sensorJSON)),
			TestAppFactory:  setupTestApp,
			ExpectedEvents: map[string]int{
				"OnBeforeApiError": 0,
				"OnAfterApiError":  0,
			},
		},
		{
			Name:   "try as authenticated admin",
			Method: http.MethodPost,
			Url:    "/sensor",
			RequestHeaders: map[string]string{
				"auth": adminToken,
			},
			ExpectedStatus:  201,
			ExpectedContent: []string{"Success"},
			Body:            strings.NewReader(string(sensorJSON)),
			TestAppFactory:  setupTestApp,
			ExpectedEvents: map[string]int{
				"OnModelAfterCreate":  1,
				"OnModelBeforeCreate": 1,
				"OnModelAfterUpdate":  1,
				"OnModelBeforeUpdate": 1,
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestGetSensor(t *testing.T) {
	initLogger()

	sensorJSON, err := json.Marshal(testSensor)
	if err != nil {
		t.Fatal(err)
	}

	setupTestApp := func(t *testing.T) *tests.TestApp {
		testApp, err := tests.NewTestApp(testDataDir)
		if err != nil {
			t.Fatal(err)
		}

		sensorRepo := repositories.PocketSensorRepo{Db: testApp, Cache: nil}
		sensorService := services.SensorService{Repo: &sensorRepo}
		authService := mocks.MockAuthService{LoginFunc: func(token string) ([]string, string, error) {
			if token == adminToken {
				return []string{"Admin"}, "admin", nil
			}
			return nil, "", nil
		}}

		sensorController := controllers.SensorController{Service: &sensorService, AuthService: authService}
		sensorController.InitSensorEndpoints(testApp)

		return testApp
	}

	scenarios := []tests.ApiScenario{
		{
			Name:            "successful retrieval of sensor",
			Method:          http.MethodGet,
			Url:             fmt.Sprint("/sensor/", testSensor.Id),
			ExpectedStatus:  http.StatusCreated,
			ExpectedContent: []string{string(sensorJSON)},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:            "sensor not found",
			Method:          http.MethodGet,
			Url:             "/sensor/unknown",
			ExpectedStatus:  http.StatusNotAcceptable,
			ExpectedContent: []string{"\"message\":\"Repo: Sensor not found\""},
			TestAppFactory:  setupTestApp,
			ExpectedEvents: map[string]int{
				"OnBeforeApiError": 0,
				"OnAfterApiError":  0,
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestAssignSensor(t *testing.T) {
	initLogger()

	setupTestApp := func(t *testing.T) *tests.TestApp {
		testApp, err := tests.NewTestApp(testDataDir)
		if err != nil {
			t.Fatal(err)
		}

		sensorRepo := repositories.PocketSensorRepo{Db: testApp, Cache: nil}
		sensorService := services.SensorService{Repo: &sensorRepo}
		authService := mocks.MockAuthService{LoginFunc: func(token string) ([]string, string, error) {
			if token == adminToken {
				return []string{"Admin"}, "admin", nil
			}
			return nil, "", nil
		}}

		sensorController := controllers.SensorController{Service: &sensorService, AuthService: authService}
		sensorController.InitSensorEndpoints(testApp)

		return testApp
	}

	scenarios := []tests.ApiScenario{
		{
			Name:            "try with different http method, eg. GET",
			Method:          http.MethodGet,
			Url:             "/sensor/assign",
			ExpectedStatus:  406,
			ExpectedContent: []string{"\"message\":\"Repo: Sensor not found\""},
			TestAppFactory:  setupTestApp,
			ExpectedEvents: map[string]int{
				"OnBeforeApiError": 0,
				"OnAfterApiError":  0,
			},
		},
		{
			Name:            "try as guest (aka. no Authorization header)",
			Method:          http.MethodPost,
			Url:             "/sensor/assign",
			ExpectedStatus:  406,
			ExpectedContent: []string{"\"message\":\"Controller: User is not an admin\""},
			TestAppFactory:  setupTestApp,
			ExpectedEvents: map[string]int{
				"OnBeforeApiError": 0,
				"OnAfterApiError":  0,
			},
		},
		{
			Name:   "try as authenticated app user",
			Method: http.MethodPost,
			Url:    "/sensor/assign",
			RequestHeaders: map[string]string{
				"auth":
				// Fill in token
				"",
			},
			ExpectedStatus:  406,
			ExpectedContent: []string{"Controller: User is not an admin"},
			TestAppFactory:  setupTestApp,
			ExpectedEvents: map[string]int{
				"OnBeforeApiError": 0,
				"OnAfterApiError":  0,
			},
		},
		{
			Name:   "try as authenticated admin",
			Method: http.MethodPost,
			Url:    "/sensor/assign",
			RequestHeaders: map[string]string{
				"auth": adminToken,
			},
			ExpectedStatus:  201,
			ExpectedContent: []string{"Success"},
			TestAppFactory:  setupTestApp,
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}
