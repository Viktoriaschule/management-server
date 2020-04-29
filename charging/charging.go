package charging

import (
	"time"

	"github.com/viktoriaschule/management-server/database"
	"github.com/viktoriaschule/management-server/log"
	"github.com/viktoriaschule/management-server/models"
)

// A loading duration without any updates is max 15 minutes long
const maxLoadingDuration = time.Minute * 15

// All the charging management
var currentLoadingEntries map[string][]models.HistoryEntry

// Prepares the device charging synchronization
func StartSync(database *database.Database, getHistoryEntriesInDuration func(*database.Database, time.Duration) (map[string][]models.HistoryEntry, error)) {

	var err error
	currentLoadingEntries, err = getHistoryEntriesInDuration(database, maxLoadingDuration)

	if err != nil {
		log.Errorf("Cannot load current loading entries: %v", err)
		return
	}
}

// Updates the device charging state and store changed battery levels in the database
//
// Before this function, the start sync and must be called.
//
// And after syncing all devices the end sync function must be called
func SyncDevice(device *models.GeneralDevice, oldDevice *models.GeneralDevice, isNew bool) {
	if !isNew {
		updateChargingState(oldDevice, device, currentLoadingEntries[device.Id])
	}
}

// Updates the charging state of a device
func updateChargingState(oldDevice *models.GeneralDevice, newDevice *models.GeneralDevice, batteryEntries []models.HistoryEntry) {
	if newDevice.BatteryLevel > oldDevice.BatteryLevel {
		// If the new battery level is higher, the device is currently charging
		newDevice.IsCharging = true
	} else if newDevice.BatteryLevel < oldDevice.BatteryLevel {
		// if the new battery level is smaller, the device is definitely not charging
		newDevice.IsCharging = false
	} else if oldDevice.IsCharging {
		// If the the battery level did not changed, but is currently marked as charged,
		// check if until the max loading duration are values higher or lower than the current

		// The entries are already reduced to the max duration and sorted by date
		newDevice.IsCharging = false
		for _, entry := range batteryEntries {
			if entry.Level > newDevice.BatteryLevel {
				// If there was a higher level, the dive is not charging anymore
				newDevice.IsCharging = false
				break
			} else if entry.Level < newDevice.BatteryLevel {
				// If there was a lower level, the device still charging
				newDevice.IsCharging = true
				break
			} else {
				// If the value is the same, the device should have a charging state
				// until max loading duration is over or one of the conditions above is not true
				// but first check all other values in the duration
				newDevice.IsCharging = true
			}
		}
	}
}
