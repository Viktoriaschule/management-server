package history

import "time"

type Request struct {
	Ids  []string  `json:"ids"`
	Date time.Time `json:"date"`
}
