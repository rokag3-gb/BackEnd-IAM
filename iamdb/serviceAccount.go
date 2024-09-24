package iamdb

import (
	"database/sql"
	"iam/models"
)

func SelectServiceAccounts(db *sql.DB) ([]models.GetServiceAccount, []string, error) {
	var arr = make([]models.GetServiceAccount, 0)
	var idArr = make([]string, 0)

	var rows *sql.Rows

	query := `SELECT 
	U.ID
	, U.USERNAME
	, E.ENABLED
	, E.CLIENT_ID
	, E.REALM_ID
	, ISNULL(TRIM(', ' FROM A.Roles), '') as Roles 
	, ISNULL(TRIM(', ' FROM D.Account), '') as Account 
	, ISNULL(TRIM(', ' FROM D.AccountId), '') as AccountId 
FROM USER_ENTITY U
JOIN CLIENT E ON U.SERVICE_ACCOUNT_CLIENT_LINK = E.ID
LEFT OUTER JOIN (select u.ID, 
		', '+string_agg(r.rName, ', ')+', ' as Roles
	FROM roles r 
	JOIN UserRole ur on r.rId = ur.RoleId
	JOIN USER_ENTITY u on ur.userId = u.ID
	GROUP BY u.ID) A ON U.ID = A.ID
LEFT OUTER JOIN (select USER_ID, ISNULL(', '+string_agg(IDENTITY_PROVIDER, ', ')+', ', '') as openid
	FROM FEDERATED_IDENTITY
	GROUP BY USER_ID) C	ON U.ID = C.USER_ID
LEFT OUTER JOIN 
	(SELECT	ISNULL(', '+string_agg(AC.AccountName, ', ')+', ', '') as Account
		, AU.UserId
		, ISNULL(', '+string_agg(AC.AccountId, ', ')+', ', '') as AccountId
	FROM [Sale].[dbo].[Account_User] AU
	JOIN [Sale].[dbo].[Account] AC ON AU.AccountId = AC.AccountId
	WHERE AU.IsUse = 1
	GROUP BY AU.UserId) D ON U.ID = D.UserId
WHERE U.SERVICE_ACCOUNT_CLIENT_LINK IS NOT NULL 
ORDER BY U.USERNAME`

	rows, err := db.Query(query)

	if err != nil {
		return arr, idArr, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.GetServiceAccount

		err := rows.Scan(&r.ID, &r.Username, &r.Enabled, &r.ClientId, &r.RealmId, &r.Roles, &r.Account, &r.AccountId)
		if err != nil {
			return arr, idArr, err
		}

		arr = append(arr, r)
		idArr = append(idArr, *r.ClientId)
	}

	return arr, idArr, err
}

func InsertUpdateClientAttribute(db *sql.DB, IdOfClient, key, value string) error {
	query := `MERGE INTO CLIENT_ATTRIBUTES A
	USING (SELECT 1 AS dual) AS B
	ON A.CLIENT_ID = ? AND A.NAME = ?
	WHEN MATCHED THEN
		UPDATE SET VALUE = ?
	WHEN NOT MATCHED THEN
		INSERT (CLIENT_ID, VALUE, NAME)
	VALUES(?, ?, ?);`

	_, err := db.Query(query, IdOfClient, key, value, IdOfClient, value, key)
	return err
}

func DeleteClientAttribute(db *sql.DB, IdOfClient string) error {
	query := `DELETE FROM CLIENT_ATTRIBUTES WHERE CLIENT_ID = ?`

	_, err := db.Query(query, IdOfClient)
	return err
}

func SelectClientAttribute(db *sql.DB, ClientId string) (map[string]string, error) {
	query := `SELECT
      CA.NAME
      , CA.VALUE
FROM [CLIENT_ATTRIBUTES] CA
JOIN CLIENT C ON CA.CLIENT_ID = C.ID 
WHERE C.CLIENT_ID = ?`

	rows, err := db.Query(query, ClientId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result = make(map[string]string)

	for rows.Next() {
		var key, value string

		err := rows.Scan(&key, &value)
		if err != nil {
			return nil, err
		}

		result[key] = value
	}

	return result, err
}

func SelectClientsAttribute(db *sql.DB, ClientIds []string) (map[string]string, error) {
	queryParams := []interface{}{}
	query := `SELECT C.CLIENT_ID
      ,CA.NAME
      ,CA.VALUE
FROM [CLIENT_ATTRIBUTES] CA
JOIN CLIENT C ON CA.CLIENT_ID = C.ID 
WHERE C.CLIENT_ID IN (`

	for _, id := range ClientIds {
		query += "?, "
		queryParams = append(queryParams, id)
	}
	query = query[:len(query)-2] + ")"

	rows, err := db.Query(query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result = make(map[string]string)

	for rows.Next() {
		var id, key, value string

		err := rows.Scan(&id, &key, &value)
		if err != nil {
			return nil, err
		}

		result[id+"_"+key] = value
	}

	return result, err
}

func SelectIdFromClientId(db *sql.DB, ClientId string) (string, error) {
	var id string

	query := `SELECT ID FROM CLIENT WHERE CLIENT_ID = ?`

	err := db.QueryRow(query, ClientId).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, err
}

func SelectClientIdFromUserId(db *sql.DB, UserId string) (string, error) {
	var id string

	query := `SELECT C.CLIENT_ID 
  FROM USER_ENTITY U
  	JOIN CLIENT C ON U.SERVICE_ACCOUNT_CLIENT_LINK = C.ID
  WHERE U.ID = ?`

	err := db.QueryRow(query, UserId).Scan(&id)

	return id, err
}

func SelectServiceAccount(db *sql.DB, userId string) (models.GetServiceAccount, error) {
	var r models.GetServiceAccount
	var rows *sql.Rows

	query := `SELECT 
	U.ID
	, U.USERNAME
	, E.ENABLED
	, E.CLIENT_ID
	, E.SECRET
	, E.REALM_ID
	, ISNULL(TRIM(', ' FROM A.Roles), '') as Roles 
	, ISNULL(TRIM(', ' FROM D.Account), '') as Account 
	, ISNULL(TRIM(', ' FROM D.AccountId), '') as AccountId 
FROM USER_ENTITY U
JOIN CLIENT E ON U.SERVICE_ACCOUNT_CLIENT_LINK = E.ID
LEFT OUTER JOIN (select u.ID, 
		', '+string_agg(r.rName, ', ')+', ' as Roles
	FROM roles r 
	JOIN UserRole ur on r.rId = ur.RoleId
	JOIN USER_ENTITY u on ur.userId = u.ID
	GROUP BY u.ID) A ON U.ID = A.ID
LEFT OUTER JOIN (select USER_ID, ISNULL(', '+string_agg(IDENTITY_PROVIDER, ', ')+', ', '') as openid
	FROM FEDERATED_IDENTITY
	GROUP BY USER_ID) C	ON U.ID = C.USER_ID
LEFT OUTER JOIN 
	(SELECT	ISNULL(', '+string_agg(AC.AccountName, ', ')+', ', '') as Account
		, AU.UserId
		, ISNULL(', '+string_agg(AC.AccountId, ', ')+', ', '') as AccountId
	FROM [Sale].[dbo].[Account_User] AU
	JOIN [Sale].[dbo].[Account] AC ON AU.AccountId = AC.AccountId
	WHERE AU.IsUse = 1
	GROUP BY AU.UserId) D ON U.ID = D.UserId
WHERE U.SERVICE_ACCOUNT_CLIENT_LINK IS NOT NULL 
AND U.ID = ?`

	rows, err := db.Query(query, userId)

	if err != nil {
		return r, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&r.ID, &r.Username, &r.Enabled, &r.ClientId, &r.Secret, &r.RealmId, &r.Roles, &r.Account, &r.AccountId)
		if err != nil {
			return r, err
		}
	}

	return r, err
}
