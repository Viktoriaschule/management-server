package database

import (
	"github.com/go-sql-driver/mysql"
	"github.com/viktoriaschule/management-server/helper"
	"github.com/viktoriaschule/management-server/log"
	"time"
)

func ParseRows(database *Database, sqlQuery string, parser func(), parserArgs ...interface{}) (err error) {

	sqlRows, _err := database.DB.Query(sqlQuery)
	if _err != nil {
		log.Errorf("Database query failed: %v", _err)
		err = &helper.LoadError{Msg: "Database query failed "}
		return err
	}

	//noinspection GoUnhandledErrorResult
	defer sqlRows.Close()
	for sqlRows.Next() {

		err := sqlRows.Scan(parserArgs...)
		if err != nil {
			log.Errorf("Database query failed: %v", err)
			err = &helper.LoadError{Msg: "Database query failed"}
			return err
		}

		// Parse the row
		parser()
	}

	err = sqlRows.Err()
	if err != nil {
		log.Errorf("Database query failed: %v", err)
		err = &helper.LoadError{Msg: "Database query failed"}
		return err
	}

	return err
}

func GetDatetime(nullTime mysql.NullTime) time.Time {
	if nullTime.Valid {
		return nullTime.Time
	} else {
		log.Warnf("Cannot parse mysql date")
		return time.Unix(0, 0)
	}
}
