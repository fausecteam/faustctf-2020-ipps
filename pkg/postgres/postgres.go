// Package postgres implements interfaces for retrieving data from a
// PostgreSQL database.
package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Config struct {
	Host string `toml:"hostname"`
	Port uint16
	// Name is the database name.
	Name string
	// User is the username.
	User string `toml:"username"`
	// Password is the database user's password
	Password string
}

const connFmt = "host=%s port=%d user=%s password=%s dbname=%s connect_timeout=5"

// Connect connects to the Postgres database using the connection settings specified
// in conf.
func Connect(conf *Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(connFmt, conf.Host, conf.Port, conf.User, conf.Password, conf.Name)
	return sql.Open("postgres", connStr)
}

// InstallTables installs all tables necessary to run the website, using
// PostgreSQL as the underlying storage.
func InstallTables(db *sql.DB) error {
	_, err := db.Exec(installUserTable)
	if err != nil {
		return err
	}
	_, err = db.Exec(installCardTable)
	if err != nil {
		return err
	}
	_, err = db.Exec(installFeedbackTable)
	if err != nil {
		return err
	}
	_, err = db.Exec(installAddressTable)
	if err != nil {
		return err
	}
	_, err = db.Exec(installParcelTable)
	if err != nil {
		return err
	}
	_, err = db.Exec(installParcelEventTable)

	return err
}
