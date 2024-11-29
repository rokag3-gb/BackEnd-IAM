package iamdb

import (
	"database/sql"
)

func SelectUserByEmail(db *sql.DB, email string) (string, error) {
	userID := ""
	query := `SELECT ID FROM USER_ENTITY WHERE EMAIL = ?`

	rows, err := db.Query(query, email)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&userID)
		if err != nil {
			return "", err
		}
	}

	return userID, err
}

func SelectEmailByUser(db *sql.DB, userID string) (string, error) {
	email := ""
	query := `SELECT EMAIL FROM USER_ENTITY WHERE ID = ?`

	rows, err := db.Query(query, userID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&email)
		if err != nil {
			return "", err
		}
	}

	return email, err
}

// 현재 AccountKey를 사용하는 Salse API가 존재하지 않아 이렇게 할 수 밖에 없습니다...
func SelectAccountUserByEmail(db *sql.DB, email, accountKey string) (bool, error) {
	query := `SELECT 1 FROM [Sale].[dbo].[AccountKey] ak
JOIN [Sale].[dbo].[Account_User] au ON ak.AccountId = au.AccountId
JOIN [IAM].[dbo].[USER_ENTITY] u ON au.UserId = u.ID
WHERE ak.AccountKey = ?
AND u.EMAIL = ?`

	rows, err := db.Query(query, accountKey, email)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		return true, err
	}

	return false, err
}

func SelectAccountIDByAccountKey(db *sql.DB, accountKey string) (int64, error) {
	var result int64
	query := `SELECT AccountId FROM [Sale].[dbo].[AccountKey]
WHERE AccountKey = ?`

	rows, err := db.Query(query, accountKey)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&result)
		if err != nil {
			return result, err
		}
	}

	return result, err
}
