package controllers

import (
	"mongo-server/controllers/repointerfaces"
	"mongo-server/mongo_models"
)

type ReportsController struct {
	repo repointerfaces.ReportsRepo
}

// Esto es un Factory Method porque no hay constructores
// En este caso lo usamos para iniciarlizar la variable privada
func NewReportsController(repo repointerfaces.ReportsRepo) ReportsController {
	// No se usa un puntero porque a las interfaces no se las trata con punteros
	// cuando manejamos su implementacion
	return ReportsController{repo: repo}
}

func (controller *ReportsController) AddAppReport(report mongo_models.AppReport) error {
	return controller.repo.AddAppReport(report)
}

func (controller *ReportsController) AddSensorReport(report mongo_models.SensorReport) error {
	return controller.repo.AddSensorReport(report)
}
