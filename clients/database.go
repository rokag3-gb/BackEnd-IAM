package clients

import (
	"database/sql"
)

var db *sql.DB = nil

func InitDbClient(driverName string, dataSourceName string) error {
	if db == nil {
		var err error
		db, err = sql.Open(driverName, dataSourceName)
		if err != nil {
			return err
		}
	}
	return nil
}

func DBClient() *sql.DB {
	return db
}
