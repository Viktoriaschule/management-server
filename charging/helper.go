package charging

import "time"

type BatteriesRequest struct {
	Ids  []string  `json:"ids"`
	Date time.Time `json:"date"`
}
