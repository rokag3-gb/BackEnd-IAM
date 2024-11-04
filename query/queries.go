package query

import (
	"database/sql"
	"fmt"
)

func resultErrorCheck(rows *sql.Rows) error {
	if rows != nil {
		count := 0
		for rows.Next() {
			err := rows.Scan(&count)
			if err != nil {
				return err
			}
		}

		if count == 0 {
			return fmt.Errorf("0 rows affected")
		}
	}

	return nil
}

func ConnectionTest(db *sql.DB) {
	query := "select 1"

	rows, err := db.Query(query)

	if err != nil {
		panic(err)
	}
	defer rows.Close()

	if !rows.Next() {
		panic("DB Connection fail")
	}
}
