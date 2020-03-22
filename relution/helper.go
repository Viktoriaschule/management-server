package relution

import "github.com/viktoriaschule/management-server/models"

type loadError struct {
	s string
}

func (e *loadError) Error() string {
	return e.s
}

type relutionDevicesResponse struct {
	Status  string
	Message string
	Errors  struct{}
	Total   int
	Results []models.RelutionDevice
}
