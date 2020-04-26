package reservation

import (
	"github.com/go-sql-driver/mysql"
	"github.com/viktoriaschule/management-server/database"
	"github.com/viktoriaschule/management-server/helper"
	"github.com/viktoriaschule/management-server/log"
	"github.com/viktoriaschule/management-server/models"
	"time"
)

const (
	maxHistoryStoreDuration = time.Hour * 24 * 30
	maxCurrentStoreDuration = time.Hour
)

// Creates a new reservation in the reservations and history list
func createOrUpdateReservation(reservation *models.Reservation, database *database.Database) (err error) {
	log.Debugf("Create or update reservation...")
	databases := []string{"reservations", "reservationHistory"}

	// If the reservation id is smaller zero, the reservation is new and need a new id
	if reservation.Id < 0 {
		// Get the current reservation id
		ids, _err := getCurrentIDs(database)
		if _err != nil {
			return _err
		}

		reservation.Id = ids.Reservations

		// Update the reservation ids count
		err = updateID(reservationIDs, database)
		if err != nil {
			return err
		}
	}

	for _, dbName := range databases {
		_, err := database.DB.Exec("INSERT INTO "+dbName+" VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE groupID = ?, timetableID = ?, participant = ?, date = ?, priority_level = ?, priority_description = ?, ipad_group = ?, modified = ?;",
			reservation.Id,
			reservation.GroupID,
			reservation.TimetableID,
			reservation.Participant,
			reservation.Date.UTC().Format(helper.SqlDateFormat),
			reservation.Priority.Level,
			reservation.Priority.Description,
			reservation.IPadGroup,
			reservation.Created.UTC().Format(helper.SqlDateFormat),
			reservation.Modified.UTC().Format(helper.SqlDateFormat),
			reservation.GroupID,
			reservation.TimetableID,
			reservation.Participant,
			reservation.Date.UTC().Format(helper.SqlDateFormat),
			reservation.Priority.Level,
			reservation.Priority.Description,
			reservation.IPadGroup,
			reservation.Modified.UTC().Format(helper.SqlDateFormat),
		)

		if err != nil {
			log.Errorf("Database query failed: %v", err)
			_err := &helper.LoadError{Msg: "Database query failed "}
			return _err
		}
	}
	return nil
}

func getReservations(filter string, database *database.Database, onlyFuture bool) (reservations []models.Reservation, err error) {
	dbName := "reservations"
	if !onlyFuture {
		dbName = "reservationHistory"
	}

	if len(filter) > 0 {
		filter = " WHERE " + filter
	}

	rows, _err := database.DB.Query("SELECT * FROM " + dbName + filter)
	if _err != nil {
		log.Errorf("Database query failed: %v", _err)
		err = &helper.LoadError{Msg: "Database query failed "}
		return nil, err
	}

	reservations = []models.Reservation{}

	// Init the reservation ids object
	reservation := models.Reservation{}
	reservation.Priority = models.Priority{}

	//noinspection GoUnhandledErrorResult
	defer rows.Close()
	for rows.Next() {
		var timestamp mysql.NullTime
		var created mysql.NullTime
		var modified mysql.NullTime

		err := rows.Scan(&reservation.Id, &reservation.GroupID, &reservation.TimetableID, &reservation.Participant, &timestamp, &reservation.Priority.Level, &reservation.Priority.Description, &reservation.IPadGroup, &created, &modified)
		if err != nil {
			log.Errorf("Database query failed: %v", err)
			err = &helper.LoadError{Msg: "Database query failed"}
			return nil, err
		}

		// Get the timestamps
		if timestamp.Valid {
			reservation.Date = timestamp.Time
		} else {
			log.Warnf("Cannot read timestamp of reservation entry")
		}
		if created.Valid {
			reservation.Created = created.Time
		} else {
			log.Warnf("Cannot read timestamp of reservation entry")
		}
		if modified.Valid {
			reservation.Modified = modified.Time
		} else {
			log.Warnf("Cannot read timestamp of reservation entry")
		}

		reservations = append(reservations, reservation)
	}
	err = rows.Err()
	if err != nil {
		log.Errorf("Database query failed: %v", err)
		err = &helper.LoadError{Msg: "Database query failed"}
		return nil, err
	}

	return reservations, err
}

func deleteReservation(reservation *models.Reservation, database *database.Database) error {
	log.Debugf("Delete reservation with id %d...", reservation.Id)
	databases := []string{"reservations", "reservationHistory"}

	for _, dbName := range databases {
		_, err := database.DB.Exec("DELETE FROM "+dbName+" WHERE id = ?;", reservation.Id)

		if err != nil {
			log.Errorf("Database query failed: %v", err)
			_err := &helper.LoadError{Msg: "Database query failed "}
			return _err
		}
	}
	return nil
}

func deleteOldReservations(database *database.Database) {
	databases := []string{"reservations", "reservationHistory"}
	durations := []time.Duration{maxCurrentStoreDuration, maxHistoryStoreDuration}

	for index, dbName := range databases {
		duration := durations[index]

		oldestDate := time.Now().Add(-duration).UTC().Format(helper.SqlDateFormat)
		log.Debugf("Remove reservations entries in %s older than %s...", dbName, oldestDate)
		_, err := database.DB.Exec("DELETE FROM "+dbName+" WHERE timestamp < ?", oldestDate)

		if err != nil {
			log.Warnf("Error deleting old reservation entries: %v", err)
		}
	}
	log.Debugf("Removed old reservations...")
}

// Adds one to the current id count
func updateID(name string, database *database.Database) error {
	_, err := database.DB.Exec("INSERT IGNORE INTO reservationIds VALUES (?, 0);", name)
	_, _err := database.DB.Exec("UPDATE reservationIds SET id=id+1 WHERE name=?;", name)

	if err != nil {
		log.Errorf("Database query failed: %v", err)
		_err := &helper.LoadError{Msg: "Database query failed "}
		return _err
	}
	if _err != nil {
		log.Errorf("Database query failed: %v", _err)
		_err := &helper.LoadError{Msg: "Database query failed "}
		return _err
	}
	return nil
}

// Returns the current id count for the reservations and groups
func getCurrentIDs(database *database.Database) (ids *models.ReservationIDs, err error) {
	rows, _err := database.DB.Query("SELECT * FROM reservationIds")
	if _err != nil {
		log.Errorf("Database query failed: %v", _err)
		err = &helper.LoadError{Msg: "Database query failed "}
		return nil, err
	}

	// Init the reservation ids object
	ids = &models.ReservationIDs{}
	ids.Reservations = 0
	ids.Groups = 0

	var name string
	var id int

	//noinspection GoUnhandledErrorResult
	defer rows.Close()
	for rows.Next() {

		err := rows.Scan(&name, &id)
		if err != nil {
			log.Errorf("Database query failed: %v", err)
			err = &helper.LoadError{Msg: "Database query failed"}
			return nil, err
		}

		switch name {
		case "reservations":
			ids.Reservations = id
			break
		case "groups":
			ids.Groups = id
			break
		}
	}
	err = rows.Err()
	if err != nil {
		log.Errorf("Database query failed: %v", err)
		err = &helper.LoadError{Msg: "Database query failed"}
		return nil, err
	}

	return ids, err
}
