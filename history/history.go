package history

import (
	"fmt"
	"github.com/viktoriaschule/management-server/charging"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/viktoriaschule/management-server/database"
	"github.com/viktoriaschule/management-server/helper"
	"github.com/viktoriaschule/management-server/log"
	"github.com/viktoriaschule/management-server/models"
)

// The max duration to store device changes (30d)
const maxStoreDuration = time.Hour * 24 * 30

// All the charging management
var changedSqlHistoryEntries []string

// All the charging management
var oldHistoryEntries map[string][]models.HistoryEntry

// Prepares the device history synchronization
func StartSync(database *database.Database) {
	changedSqlHistoryEntries = []string{}

	// Load all old entries
	var err error
	oldHistoryEntries, err = getHistoryEntriesForDevicesAndTime(database, nil, nil)

	if err != nil {
		log.Warnf("Error during fetching old history entries: %v", err)
	}

	charging.StartSync(database, getHistoryEntriesInDuration)
}

// Adds a device state to the history
func SyncDevice(device *models.GeneralDevice, oldDevice *models.GeneralDevice, isNew bool) {
	oldEntries, isNotNew := oldHistoryEntries[device.Id]

	charging.SyncDevice(device, oldDevice, isNew)

	if isNotNew {
		for _, entry := range oldEntries {
			if models.CompareTimes(entry.Modified, device.LastModified) {
				if models.HasObjectChanged(entry, *models.DeviceToHistoryEntry(device)) {
					log.Warnf("Value changed, but modified date is the same: %s %s", device.Id, device.LastModified)
				}
				return
			}
		}
	}

	if isNew || !isNotNew {
		changedSqlHistoryEntries = append(changedSqlHistoryEntries, getSqlHistoryEntry(device))
	}
}

// Synchronizes all previous synced devices to the database
// and removes all the too old values
func EndSync(database *database.Database) {
	addHistoryEntries(database, &changedSqlHistoryEntries)
	removeOldHistoryEntries(database)
}

// Returns an sql batter entry value
func getSqlHistoryEntry(device *models.GeneralDevice) string {
	entry := models.DeviceToHistoryEntry(device)

	return fmt.Sprintf(`("%s", "%d", "%s", "%s", "%s", "%s", "%s")`,
		entry.Id,
		entry.Level,
		entry.LoggedinUser,
		entry.Status,
		entry.LastConnection.Format(helper.SqlDateFormat),
		entry.Modified.Format(helper.SqlDateFormat),
		entry.Timestamp.Format(helper.SqlDateFormat),
	)
}

// Adds the given sql history values to the database
func addHistoryEntries(database *database.Database, entries *[]string) {
	log.Infof("Add %d history entries...", len(*entries))

	if len(changedSqlHistoryEntries) > 0 {
		_, err := database.DB.Exec("INSERT INTO history VALUES " + strings.Join(*entries, ", "))

		if err != nil {
			log.Warnf("Error during adding a new history entry: %v", err)
		}

		log.Infof("Added history entries...")
	}
}

// Removes all history entries older than the max store duration
func removeOldHistoryEntries(database *database.Database) {
	oldestDate := time.Now().Add(-maxStoreDuration).Format(helper.SqlDateFormat)
	log.Infof("Remove history entries older than %s...", oldestDate)
	_, err := database.DB.Exec("DELETE FROM history WHERE timestamp < ?", oldestDate)

	if err != nil {
		log.Warnf("Error deleting old battery level entries: %v", err)
	}

	log.Infof("Removed old devices...")
}

// Returns all battery entries in the last max loading duration sorted by the date
func getHistoryEntriesInDuration(database *database.Database, duration time.Duration) (entries map[string][]models.HistoryEntry, err error) {
	oldestDate := time.Now().Add(duration).Format(helper.SqlDateFormat)
	return getHistoryEntriesForDevicesAndTime(database, nil, &oldestDate)
}

// Returns all battery entries for the given devices
func GetHistoryEntriesForDevices(database *database.Database, ids []string, date time.Time) (entries map[string][]models.HistoryEntry, err error) {
	oldestDate := date.Format(helper.SqlDateFormat)
	return getHistoryEntriesForDevicesAndTime(database, &ids, &oldestDate)
}

// Returns all battery entries in the last max loading duration and with the given ids sorted by the date
func getHistoryEntriesForDevicesAndTime(database *database.Database, ids *[]string, oldestDate *string) (entries map[string][]models.HistoryEntry, err error) {
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

	rows, _err := database.DB.Query("SELECT * FROM history " + filter + "ORDER BY timestamp DESC")
	if _err != nil {
		log.Errorf("Database query failed: ", _err)
		err = &helper.LoadError{Msg: fmt.Sprintf("Database query failed")}
		return nil, err
	}

	entries = map[string][]models.HistoryEntry{}
	entry := &models.HistoryEntry{}
	count := 0
	//noinspection GoUnhandledErrorResult
	defer rows.Close()
	for rows.Next() {
		count++
		var timestamp mysql.NullTime
		var modified mysql.NullTime
		var connection mysql.NullTime
		err := rows.Scan(&entry.Id, &entry.Level, &entry.LoggedinUser, &entry.Status, &modified, &connection, &timestamp)
		if err != nil {
			log.Errorf("Database query failed: ", err)
			err = &helper.LoadError{Msg: "Database query failed"}
			return nil, err
		}

		// Get all dates
		if timestamp.Valid {
			entry.Timestamp = timestamp.Time
		} else {
			log.Warnf("Cannot read timestamp of history entry")
		}
		if modified.Valid {
			entry.Modified = modified.Time
		} else {
			log.Warnf("Cannot read last modified of history entry")
		}
		if connection.Valid {
			entry.LastConnection = connection.Time
		} else {
			log.Warnf("Cannot read last connection of history entry")
		}

		_, exists := entries[entry.Id]

		if !exists {
			entries[entry.Id] = []models.HistoryEntry{*entry}
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
	log.Infof("Loaded %d history entries", count)
	return entries, err
}
