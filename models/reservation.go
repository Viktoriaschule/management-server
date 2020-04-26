package models

import (
	"time"
)

type Reservation struct {
	Id          int       `json:"id"`
	GroupID     int       `json:"groupID"`
	TimetableID string    `json:"timetableID"`
	Participant string    `json:"participant"`
	Date        time.Time `json:"date"`
	Priority    Priority  `json:"priority"`
	IPadGroup   int       `json:"iPadGroup"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

type Group struct {
	Id           int    `json:"id"`
	Reservations []int  `json:"reservations"`
	Name         string `json:"name"`
	Participant  string `json:"participant"`
}

type Priority struct {
	Level       int    `json:"level"`
	Description string `json:"description"`
}

type ReservationIDs struct {
	Reservations int `json:"reservations"`
	Groups       int `json:"groups"`
}
