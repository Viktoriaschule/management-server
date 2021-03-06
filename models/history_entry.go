package models

import "time"

type HistoryEntry struct {
	Id           string    `json:"id"`
	Level        int64     `json:"level"`
	Modified     time.Time `json:"modified"`
	Timestamp    time.Time `json:"timestamp"`
	LoggedinUser string    `json:"loggedin_user"`
	Status       string    `json:"status"`
}

func DeviceToHistoryEntry(device *GeneralDevice) *HistoryEntry {
	return &HistoryEntry{
		Id:           device.Id,
		Level:        device.BatteryLevel,
		Timestamp:    time.Now(),
		Modified:     device.LastModified,
		LoggedinUser: device.LoggedinUser,
		Status:       device.Status,
	}
}
