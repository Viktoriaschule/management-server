package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/viktoriaschule/management-server/config"
	"github.com/viktoriaschule/management-server/log"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase(config *config.Config) *Database {
	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			config.Mysql.User,
			config.Mysql.Password,
			config.Mysql.Host,
			config.Mysql.Port,
			config.Mysql.Name,
		))
	if err != nil {
		log.Errorf("Error connecting to database: %v", err)
		os.Exit(1)
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	return &Database{
		DB: db,
	}
}

func (d Database) CreateTables() {
	statements := []string{
		"CREATE TABLE IF NOT EXISTS devices (id VARCHAR(12) NOT NULL, name TEXT NOT NULL, loggedin_user TEXT NOT NULL, device_type BOOLEAN NOT NULL, battery_level FLOAT NOT NULL, is_charging BOOLEAN, device_group INT NOT NULL, device_group_index VARCHAR(1) NOT NULL, last_modified DATETIME, last_connection DATETIME NOT NULL, status TEXT NOT NULL, PRIMARY KEY (id))",
		"CREATE TABLE IF NOT EXISTS history (id VARCHAR(12) NOT NULL, level FLOAT NOT NULL, loggedin_user TEXT NOT NULL, status TEXT NOT NULL, last_connection DATETIME NOT NULL, modified DATETIME NOT NULL, timestamp DATETIME NOT NULL, PRIMARY KEY (id, modified))",
	}
	for _, statement := range statements {
		_, err := d.DB.Exec(statement)
		if err != nil {
			log.Errorf("Error executing statement: %v", err)
			os.Exit(1)
		}
	}
}
