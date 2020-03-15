package relution

import "github.com/viktoriaschule/management-server/models"

type relutionDevicesResponse struct {
	Status  string
	Message string
	Errors  struct{}
	Total   int
	Results []models.RelutionDevice
}
