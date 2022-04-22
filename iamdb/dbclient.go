package iamdb

import (
	"database/sql"
)

var db *sql.DB = nil

func InitDbClient(driverName string, dataSourceName string) error {
	if db == nil {
		var err error
		db, err = sql.Open(driverName, dataSourceName)
		if err != nil {
			panic(err)
		}
	}

	rows, err := ConnectionTest()

	if err != nil {
		panic(err)
	}
	defer rows.Close()

	if !rows.Next() {
		panic("DB Connection fail")
	}

	return nil
}

func DBClient() *sql.DB {
	return db
}
