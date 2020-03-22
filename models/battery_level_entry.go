package models

import (
	"time"
)

type BatteryLevelEntry struct {
	Id        string    `json:"id"`
	Level     int64     `json:"level"`
	Timestamp time.Time `json:"timestamp"`
}

func DeviceToBatteryLevelEntry(device *GeneralDevice) *BatteryLevelEntry {
	return &BatteryLevelEntry{
		Id:        device.Id,
		Level:     device.BatteryLevel,
		Timestamp: time.Now(),
	}
}
