package charging

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/viktoriaschule/management-server/database"
	"github.com/viktoriaschule/management-server/helper"
	"github.com/viktoriaschule/management-server/log"
	"github.com/viktoriaschule/management-server/models"
	"strings"
	"time"
)

// A loading duration without any updates is max 15 minutes long
const maxLoadingDuration = time.Minute * time.Duration(15)

// The max duration to store device battery levels (7d)
const maxStoreDuration = time.Hour * time.Duration(24) * 7

// All the charging management
var changedSqlLoadingEntries []string
var currentLoadingEntries map[string][]models.BatteryLevelEntry

// Prepares the device charging synchronization
func StartSync(database *database.Database) {
	changedSqlLoadingEntries = []string{}

	var err error
	currentLoadingEntries, err = getBatteryEntriesInLoadingDuration(database)

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
	if isNew || device.BatteryLevel != oldDevice.BatteryLevel {
		changedSqlLoadingEntries = append(changedSqlLoadingEntries, getSqlBatteryEntry(device))
	}

	if !isNew {
		updateChargingState(oldDevice, device, currentLoadingEntries[device.Id])
	}
}

// Synchronizes all previous synced devices to the database
// and removes all the too old values
func EndSync(database *database.Database) {
	if len(changedSqlLoadingEntries) > 0 {
		addBatteryEntries(database, &changedSqlLoadingEntries)
	}
	removeOldBatteryEntries(database)
}

// Updates the charging state of a device
func updateChargingState(oldDevice *models.GeneralDevice, newDevice *models.GeneralDevice, batteryEntries []models.BatteryLevelEntry) {
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

// Returns an sql batter entry value
func getSqlBatteryEntry(device *models.GeneralDevice) string {
	entry := models.DeviceToBatteryLevelEntry(device)
	return fmt.Sprintf(`("%s", "%d", "%s")`, entry.Id, entry.Level, entry.Timestamp.Format("2006-01-02 15:04:05"))
}

// Adds the given sql battery level values to the database
func addBatteryEntries(database *database.Database, entries *[]string) {
	log.Infof("Add %d battery entries...", len(*entries))
	_, err := database.DB.Exec("INSERT INTO battery VALUES " + strings.Join(*entries, ", "))

	if err != nil {
		log.Warnf("Error during adding a new battery level entry: %v", err)
	}

	log.Infof("Added battery entries...")
}

// Removes all battery entries older than the max store duration
func removeOldBatteryEntries(database *database.Database) {
	oldestDate := time.Now().Add(-maxStoreDuration).Format("2006-01-02 15:04:05")
	log.Infof("Remove battery entries older than %s...", oldestDate)
	_, err := database.DB.Exec("DELETE FROM battery WHERE timestamp < ?", oldestDate)

	if err != nil {
		log.Warnf("Error deleting old battery level entries: %v", err)
	}

	log.Infof("Removed old devices...")
}

// Returns all battery entries in the last max loading duration sorted by the date
func getBatteryEntriesInLoadingDuration(database *database.Database) (entries map[string][]models.BatteryLevelEntry, err error) {
	oldestDate := time.Now().Add(-maxLoadingDuration).Format("2006-01-02 15:04:05")
	return getBatteryEntriesForDevicesAndTime(database, nil, &oldestDate)
}

// Returns all battery entries for the given devices
func GetBatteryEntriesForDevices(database *database.Database, ids []string) (entries map[string][]models.BatteryLevelEntry, err error) {
	return getBatteryEntriesForDevicesAndTime(database, &ids, nil)
}

// Returns all battery entries in the last max loading duration and with the given ids sorted by the date
func getBatteryEntriesForDevicesAndTime(database *database.Database, ids *[]string, oldestDate *string) (entries map[string][]models.BatteryLevelEntry, err error) {
	// Only entries newer than oldest date, if set
	timeFilter := ""
	if oldestDate != nil {
		timeFilter = "timestamp >= \"" + *oldestDate + "\""
	}

	// Filter for all given ids, or when no given, return all
	idFilter := ""
	if ids != nil && len(*ids) > 0 {
		idsCount := len(*ids)
		idFilter += "("
		for i, id := range *ids {
			idFilter += "id = \"" + id + "\""
			if i != idsCount-1 {
				idFilter += "OR "
			}
		}
		idFilter += ")"
	}

	filter := ""
	if len(timeFilter) > 0 || len(idFilter) > 0 {
		filter += "WHERE "

		if len(timeFilter) > 0 {
			filter += timeFilter

			if len(idFilter) > 0 {
				filter += " AND "
			}
		}

		if len(idFilter) > 0 {
			filter += idFilter
		}
		filter += " "
	}

	rows, _err := database.DB.Query("SELECT * FROM battery " + filter + "ORDER BY timestamp DESC")
	if _err != nil {
		log.Errorf("Database query failed: ", _err)
		err = &helper.LoadError{Msg: fmt.Sprintf("Database query failed")}
		return nil, err
	}

	entries = map[string][]models.BatteryLevelEntry{}
	entry := &models.BatteryLevelEntry{}
	count := 0
	//noinspection GoUnhandledErrorResult
	defer rows.Close()
	for rows.Next() {
		count++
		var timestamp mysql.NullTime
		err := rows.Scan(&entry.Id, &entry.Level, &timestamp)
		if err != nil {
			log.Errorf("Database query failed: ", err)
			err = &helper.LoadError{Msg: "Database query failed"}
			return nil, err
		}
		if timestamp.Valid {
			entry.Timestamp = timestamp.Time
		} else {
			log.Warnf("Cannot read timestamp of battery entry")
		}

		_, exists := entries[entry.Id]

		if !exists {
			entries[entry.Id] = []models.BatteryLevelEntry{*entry}
		} else {
			entries[entry.Id] = append(entries[entry.Id], *entry)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Errorf("Database query failed: ", err)
		err = &helper.LoadError{Msg: "Database query failed"}
		return nil, err
	}
	log.Infof("Found %d battery entries for the max duration", count)
	return entries, err
}
