package models

type IPadGroup struct {
	Id        int `json:"id"`
	Color     int `json:"color"`
	StationId int `json:"stationID"`
}

type Station struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
