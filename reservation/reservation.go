package reservation

import (
	"github.com/go-sql-driver/mysql"
	"github.com/viktoriaschule/management-server/database"
	"github.com/viktoriaschule/management-server/helper"
	"github.com/viktoriaschule/management-server/log"
	"github.com/viktoriaschule/management-server/models"
	"github.com/viktoriaschule/management-server/relution"
	"strconv"
	"strings"
	"time"
)

const (
	maxHistoryStoreDuration = time.Hour * 24 * 30
	maxCurrentStoreDuration = time.Hour
	Success                 = 200
	Request                 = 202
	NotAllowed              = 403
	Failed                  = 500
)

// Creates a new reservation in the reservations and history list
func createOrUpdateReservation(reservation *models.Reservation, database *database.Database) (status int, err error) {
	log.Debugf("Create or update reservation...")
	databases := []string{"reservations", "reservationHistory"}

	// If the reservation id is smaller zero, the reservation is new and need a new id
	if reservation.Id < 0 {
		// Get the current reservation id
		ids, _err := getCurrentIDs(database)
		if _err != nil {
			return Failed, _err
		}

		reservation.Id = ids.Reservations

		// Update the reservation ids count
		err = updateID(reservationIDs, database)
		if err != nil {
			return Failed, err
		}
	} else {
		// If the reservation was updated, delete the old iPad groups reservations
		err = deleteIPadGroupsOfReservation(reservation.Id, database)
		if err != nil {
			return Failed, err
		}
	}

	status, err = reserveIPadGroups(reservation, database)

	if err != nil || status != Success {
		return status, err
	}

	for _, dbName := range databases {
		_, err := database.DB.Exec("INSERT INTO "+dbName+" VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE groupID = ?, timetableID = ?, participant = ?, date = ?, priority_level = ?, priority_description = ?, modified = ?;",
			reservation.Id,
			reservation.GroupID,
			reservation.TimetableID,
			reservation.Participant,
			reservation.Date.UTC().Format(helper.SqlDateTimeFormat),
			reservation.Priority.Level,
			reservation.Priority.Description,
			reservation.Created.UTC().Format(helper.SqlDateTimeFormat),
			reservation.Modified.UTC().Format(helper.SqlDateTimeFormat),
			reservation.GroupID,
			reservation.TimetableID,
			reservation.Participant,
			reservation.Date.UTC().Format(helper.SqlDateTimeFormat),
			reservation.Priority.Level,
			reservation.Priority.Description,
			reservation.Modified.UTC().Format(helper.SqlDateTimeFormat),
		)

		if err != nil {
			log.Errorf("Database query failed: %v", err)
			_err := &helper.LoadError{Msg: "Database query failed "}
			return Failed, _err
		}
	}
	return Success, nil
}

func reserveIPadGroups(reservation *models.Reservation, database *database.Database) (status int, err error) {

	date := reservation.Date.UTC().Format(helper.SqlDateFormat)
	unit, _ := strconv.Atoi(strings.Split(reservation.TimetableID, "-")[3])

	freeIPadGroups, _err := GetFreeIPadGroups(reservation.Date, unit, database)

	if _err != nil {
		return Failed, _err
	}

	// Check if the reserved groups count is not to big
	if len(reservation.IPadGroups) > freeIPadGroups.MaxCount {
		return NotAllowed, nil
	}

	// Check if all groups are free to reserve
	for index, iPadGroup := range reservation.IPadGroups {
		if helper.ContainsInt(freeIPadGroups.UnpinnedGroups, iPadGroup) {
			if reservation.IsPinned {
				//TODO: Switch groups
				return Request, nil
			} else if len(freeIPadGroups.FreeGroups) > 0 {
				reservation.IPadGroups[index] = freeIPadGroups.FreeGroups[0]
				freeIPadGroups.FreeGroups = freeIPadGroups.FreeGroups[1:]
				freeIPadGroups.UnpinnedGroups = append(freeIPadGroups.UnpinnedGroups, reservation.IPadGroups[index])
			}
		} else if helper.ContainsInt(freeIPadGroups.ReservedGroups, iPadGroup) {
			if reservation.IsPinned {
				//TODO: Ask for permission
				return Request, nil
			} else if len(freeIPadGroups.FreeGroups) > 0 {
				reservation.IPadGroups[index] = freeIPadGroups.FreeGroups[0]
				freeIPadGroups.FreeGroups = freeIPadGroups.FreeGroups[1:]
				freeIPadGroups.UnpinnedGroups = append(freeIPadGroups.UnpinnedGroups, iPadGroup)
			}
		} else if !helper.ContainsInt(freeIPadGroups.FreeGroups, iPadGroup) {
			// The iPad group to reserve does not exists
			return NotAllowed, nil
		}
	}

	for _, iPadGroup := range reservation.IPadGroups {

		// Reserve the iPad groups for the correct unit
		_, err = database.DB.Exec("INSERT INTO reservationUnits VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE is_pinned = ?, reservation_id = ?;",
			date,
			unit,
			iPadGroup,
			reservation.IsPinned,
			reservation.Id,
			reservation.IsPinned,
			reservation.Id,
		)

		if err != nil {
			log.Errorf("Database query failed: %v", err)
			_err := &helper.LoadError{Msg: "Database query failed "}
			return Failed, _err
		}
	}

	return Success, nil
}

func deleteIPadGroupsOfReservation(reservationID int, database *database.Database) error {
	_, err := database.DB.Exec("DELETE FROM reservationUnits WHERE reservation_id = ?;", reservationID)

	if err != nil {
		log.Errorf("Database query failed: %v", err)
		_err := &helper.LoadError{Msg: "Database query failed "}
		return _err
	}
	return nil
}

func GetFreeIPadGroups(date time.Time, unit int, db *database.Database) (freeIPadGroups *models.FreeIPadGroups, err error) {
	_reservedGroups, err := getReservedIPadGroups(date, unit, db)
	reservedGroups := make(map[int]models.ReservationUnit)

	if err != nil {
		return nil, err
	}

	for _, group := range _reservedGroups {
		reservedGroups[group.IPadGroup] = group
	}

	freeIPadGroups = &models.FreeIPadGroups{}
	freeIPadGroups.FreeGroups = []int{}
	freeIPadGroups.UnpinnedGroups = []int{}
	freeIPadGroups.MaxCount = 0

	allGroups, err := relution.GetAllIPadGroups(db)

	if err != nil {
		return nil, err
	}

	for _, group := range *allGroups {
		// If the group is already reserved, check if it important which group is reserved
		if reservation, reserved := reservedGroups[int(group)]; reserved {
			if reservation.IsPinned {
				freeIPadGroups.ReservedGroups = append(freeIPadGroups.ReservedGroups, int(group))
			} else {
				freeIPadGroups.UnpinnedGroups = append(freeIPadGroups.UnpinnedGroups, int(group))
			}
		} else {
			freeIPadGroups.FreeGroups = append(freeIPadGroups.FreeGroups, int(group))
			freeIPadGroups.MaxCount++
		}
	}

	return freeIPadGroups, nil
}

func getReservedIPadGroups(date time.Time, unit int, db *database.Database) (iPadGroups []models.ReservationUnit, err error) {

	parsedDate := date.UTC().Format(helper.SqlDateFormat)

	iPadGroups = []models.ReservationUnit{}
	iPadGroup := models.ReservationUnit{}
	var _date mysql.NullTime

	err = database.ParseRows(
		db,
		"SELECT * FROM reservationUnits "+`WHERE reservation_date = "`+parsedDate+`" AND unit = `+strconv.Itoa(unit)+";",
		func() {
			iPadGroup.Date = database.GetDatetime(_date)
			iPadGroups = append(iPadGroups, iPadGroup)
		},
		&_date, &iPadGroup.Unit, &iPadGroup.IPadGroup, &iPadGroup.IsPinned, &iPadGroup.ReservationID,
	)

	return
}

func getReservations(filter string, db *database.Database, onlyFuture bool) (reservations []models.Reservation, err error) {
	dbName := "reservations"
	if !onlyFuture {
		dbName = "reservationHistory"
	}

	if len(filter) > 0 {
		filter = " WHERE " + filter
	}

	reservations = []models.Reservation{}
	reservation := models.Reservation{}
	reservation.Priority = models.Priority{}
	var date mysql.NullTime
	var created mysql.NullTime
	var modified mysql.NullTime

	err = database.ParseRows(
		db,
		"SELECT * FROM "+dbName+filter,
		func() {
			reservation.Date = database.GetDatetime(date)
			reservation.Created = database.GetDatetime(created)
			reservation.Modified = database.GetDatetime(modified)
			reservations = append(reservations, reservation)
		},
		&reservation.Id, &reservation.GroupID, &reservation.TimetableID, &reservation.Participant, &date, &reservation.Priority.Level, &reservation.Priority.Description, &created, &modified,
	)

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

	return deleteIPadGroupsOfReservation(reservation.Id, database)
}

//TODO: Delete all old values in all tables
func deleteOldReservations(database *database.Database) {
	databases := []string{"reservations", "reservationHistory"}
	durations := []time.Duration{maxCurrentStoreDuration, maxHistoryStoreDuration}

	for index, dbName := range databases {
		duration := durations[index]

		oldestDate := time.Now().Add(-duration).UTC().Format(helper.SqlDateTimeFormat)
		log.Debugf("Remove reservations entries in %s older than %s...", dbName, oldestDate)
		_, err := database.DB.Exec("DELETE FROM "+dbName+" WHERE timestamp < ?", oldestDate)

		if err != nil {
			log.Warnf("Error deleting old reservation entries: %v", err)
		}
	}
	log.Debugf("Removed old reservations...")
}

// Creates a new reservation in the reservations and history list
func createOrUpdateGroup(group *models.Group, database *database.Database) (err error) {
	log.Debugf("Create or update reservation group...")

	// If the reservation id is smaller zero, the reservation is new and need a new id
	if group.Id < 0 {
		// Get the current reservation id
		ids, _err := getCurrentIDs(database)
		if _err != nil {
			return _err
		}

		group.Id = ids.Groups

		// Update the reservation ids count
		err = updateID(groupsIDs, database)
		if err != nil {
			return err
		}
	}

	_, err = database.DB.Exec("INSERT INTO reservationGroups VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE name = ?, participant = ?;",
		&group.Id,
		&group.Name,
		&group.Participant,
		&group.Name,
		&group.Participant,
	)

	if err != nil {
		log.Errorf("Database query failed: %v", err)
		_err := &helper.LoadError{Msg: "Database query failed "}
		return _err
	}
	return nil
}

func getGroups(filter string, db *database.Database) (groups []models.Group, err error) {

	if len(filter) > 0 {
		filter = " WHERE " + filter
	}

	groups = []models.Group{}
	group := models.Group{}

	err = database.ParseRows(
		db,
		"SELECT * FROM reservationGroups"+filter,
		func() {
			groups = append(groups, group)
		},
		&group.Id, &group.Name, &group.Participant,
	)

	return groups, err
}

func deleteGroup(group *models.Group, database *database.Database) error {
	log.Debugf("Delete group with id %d...", group.Id)

	_, err := database.DB.Exec("DELETE FROM reservationGroups WHERE id = ?;", group.Id)

	if err != nil {
		log.Errorf("Database query failed: %v", err)
		_err := &helper.LoadError{Msg: "Database query failed "}
		return _err
	}
	return nil
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
func getCurrentIDs(db *database.Database) (ids *models.ReservationIDs, err error) {

	// Init the reservation ids object
	ids = &models.ReservationIDs{}
	ids.Reservations = 0
	ids.Groups = 0

	var name string
	var id int

	err = database.ParseRows(db,
		"SELECT * FROM reservationIds",
		func() {
			switch name {
			case "reservations":
				ids.Reservations = id
				break
			case "groups":
				ids.Groups = id
				break
			}
		},
		&name, &id,
	)

	return ids, err
}
