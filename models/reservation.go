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
	IPadGroups  []int     `json:"iPadGroups"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
	// Whether it is important that it is exactly this iPadGroup or not
	IsPinned bool `json:"isPinned"`
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

type ReservationUnit struct {
	Date      time.Time `json:"date"`
	Unit      int       `json:"unit"`
	IPadGroup int       `json:"iPadGroup"`
	// Whether it is important that it is exactly this iPadGroup or not
	IsPinned      bool `json:"isPinned"`
	ReservationID int  `json:"reservationID"`
}

type FreeIPadGroups struct {
	MaxCount       int   `json:"maxCount"`
	FreeGroups     []int `json:"freeGroups"`
	UnpinnedGroups []int `json:"unpinnedGroups"`
	ReservedGroups []int `json:"reservedGroups"`
}
