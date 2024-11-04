package clients

import (
	"database/sql"
	"iam/query"
)

var dbConfig DbConfig = DbConfig{}

type DbConfig struct {
	DriverName     string
	DataSourceName string
}

func InitDbClient(driverName string, dataSourceName string) error {
	if dbConfig == (DbConfig{}) {
		dbConfig = DbConfig{
			DriverName:     driverName,
			DataSourceName: dataSourceName,
		}
	}
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		panic(err)
	}

	query.ConnectionTest(db)

	return nil
}

func DBClient() (*sql.DB, error) {
	db, err := sql.Open(dbConfig.DriverName, dbConfig.DataSourceName)
	if err != nil {
		return nil, err
	}
	return db, nil
}
