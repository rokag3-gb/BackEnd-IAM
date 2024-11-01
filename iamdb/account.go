package iamdb

import (
	"database/sql"
	"strings"
)

func SelectAccount(db *sql.DB, email, user_id string) (bool, error) {
	ret := false

	query := `DECLARE @AccId BIGINT
	DECLARE @UserId varchar(36)
	
	SET @UserId = ?
	SET @AccId = (SELECT AccountId FROM 
		  SALE.dbo.Account
		  WHERE ? LIKE '%' + EmailDomain)
	
	IF @AccId IS NULL
		BEGIN
			-- 계정에 대응되는 Account가 없다는 것을 알려줌
			SELECT 0
		END
	ELSE
		BEGIN
			IF NOT EXISTS (SELECT au.seq FROM SALE.dbo.Account_User au
					WHERE au.AccountId = @AccId
					AND au.UserId = @UserId)
				BEGIN
					INSERT INTO SALE.dbo.Account_User (AccountId, UserId) VALUES (@AccId, @UserId)
				END
			-- 계정에 대응되는 Account가 존재한다는 것을 알려줌
			SELECT 1
		END`

	rows, err := db.Query(query, user_id, email)
	if err != nil {
		return ret, err
	}
	defer rows.Close()

	for rows.Next() {
		var result int

		err := rows.Scan(&result)
		if err != nil {
			return ret, err
		}

		if result == 1 {
			ret = true
		}
	}

	return ret, err
}

func SelectAccountId(db *sql.DB, userId string) ([]int64, error) {
	ret := make([]int64, 0)

	query := `SELECT [AccountId]
FROM [Sale].[dbo].[Account_User]
WHERE [UserId] = ?`

	rows, err := db.Query(query, userId)
	if err != nil {
		return ret, err
	}
	defer rows.Close()

	for rows.Next() {
		var result int64

		err := rows.Scan(&result)
		if err != nil {
			return ret, err
		}

		ret = append(ret, result)
	}

	return ret, err
}

func SelectDefaultAccount(db *sql.DB, email, userId string) ([]int64, error) {
	ret := make([]int64, 0)

	if !strings.Contains(email, "@") {
		return ret, nil
	}

	emailString := email[strings.Index(email, "@"):]

	query := `SELECT AccountId 
FROM [Sale].[dbo].[Account]
WHERE EmailDomain = ?
AND AccountId NOT IN (SELECT AccountId
FROM [Sale].[dbo].[Account_User]
WHERE UserId = ?)`

	rows, err := db.Query(query, emailString, userId)
	if err != nil {
		return ret, err
	}
	defer rows.Close()

	for rows.Next() {
		var result int64

		err := rows.Scan(&result)
		if err != nil {
			return ret, err
		}

		ret = append(ret, result)
	}

	return ret, err
}

func CheckAccountUser(db *sql.DB, accountId, userId, realm string) (bool, error) {
	query := `SELECT AU.AccountId
	FROM [Sale].[dbo].[Account_User] AU
	JOIN [IAM].[dbo].[USER_ENTITY] U
	ON AU.UserId = U.ID
	WHERE ((AccountId = ? AND UserId = ?)
	OR (AccountId = '1' AND UserId = ?))
	AND REALM_ID = ?`

	rows, err := db.Query(query, accountId, userId, userId, realm)

	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		return true, nil
	}

	return false, nil
}

func InsertAccountUser(db *sql.DB, accountId, userId, saveId string) error {
	query := `INSERT INTO Sale.dbo.Account_User(AccountId, UserId, SaveId) VALUES(?, ?, ?)`

	_, err := db.Query(query, accountId, userId, saveId)
	if err != nil {
		return err
	}

	return nil
}
