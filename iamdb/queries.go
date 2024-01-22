package iamdb

import (
	"database/sql"
	"fmt"
	"iam/models"
	"strings"
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

func CreateUserAddRole(uid, username, realm string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `DECLARE @C_USERNAME nvarchar(255);
	SET @C_USERNAME = (SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?)
	
	INSERT INTO user_roles_mapping(userId, rId, createId, createDate, modifyId, modifyDate)
	(SELECT ? as userId, rId,
		@C_USERNAME as createId,
		GETDATE() as createDate,
		@C_USERNAME as modifyId,
		GETDATE() as modifyDate
	from roles A
	where defaultRole = 1)`

	_, err := db.Query(query, username, realm, uid)
	if err != nil {
		return err
	}

	return nil
}

func CreateAccountUser(accountId string, userId string, saveId string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `INSERT INTO Sale.dbo.Account_User(AccountId, UserId, SaveId)
	VALUES(?, ?, ?)`

	_, err := db.Query(query, accountId, userId, saveId)
	if err != nil {
		return err
	}

	return nil
}

func GetIdpCount(realm string) ([]models.MetricItem, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	query := `select 
	A.PROVIDER_ALIAS, 
	count(B.IDENTITY_PROVIDER) as count
	from IDENTITY_PROVIDER A
	LEFT OUTER JOIN FEDERATED_IDENTITY B
	ON B.IDENTITY_PROVIDER = A.PROVIDER_ALIAS
	WHERE A.REALM_ID = ?
	GROUP BY A.PROVIDER_ALIAS`

	rows, err := db.Query(query, realm)
	if err != nil {
		return nil, err
	}

	arr := make([]models.MetricItem, 0)

	for rows.Next() {
		var m models.MetricItem
		err = rows.Scan(&m.Key, &m.Value)
		if err != nil {
			return nil, err
		}

		arr = append(arr, m)
	}

	return arr, nil
}

func GetAllSecret(data []models.SecretGroupItem, realm string) ([]models.SecretItem, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	var groups []string

	args := []interface{}{}
	for _, group := range data {
		groups = append(groups, group.Name)
		args = append(args, group.Name)
	}
	args = append(args, realm)

	query := `SELECT SG.vSecretGroupPath, 
	S.vSecretPath, 
	S.url, 
	FORMAT(S.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(S.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
FROM vSecret S
JOIN vSecretGroup SG
	ON S.vSecretGroupId = SG.vSecretGroupId
LEFT OUTER JOIN USER_ENTITY u1
	on S.createId = u1.ID
LEFT OUTER JOIN USER_ENTITY u2
	on S.modifyId = u2.ID
WHERE SG.vSecretGroupPath IN (?` + strings.Repeat(",?", len(groups)-1) + `)
AND SG.REALM_ID = ?`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	arr := make([]models.SecretItem, 0)

	for rows.Next() {
		var r models.SecretItem

		err := rows.Scan(&r.SecretGroup, &r.Name, &r.Url, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}

	return arr, err
}

func CheckAccountUser(accountId, userId, realm string) (bool, error) {
	db, err := DBClient()
	defer db.Close()
	if err != nil {
		return false, err
	}

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

func SelectNotExsistRole(client_id, user_id, realm string) ([]string, error) {
	var arr = make([]string, 0)

	db, err := DBClient()
	defer db.Close()
	if err != nil {
		return arr, err
	}

	query := `SELECT roleId FROM IAM.dbo.clientDefaultRole
	WHERE clientId = ?
	AND isEnable = 1
	AND roleId NOT IN
	(SELECT rId FROM IAM.dbo.user_roles_mapping
	WHERE userId = ?)`

	rows, err := db.Query(query, client_id, user_id)
	if err != nil {
		return arr, err
	}
	defer rows.Close()

	for rows.Next() {
		var rId string

		err := rows.Scan(&rId)
		if err != nil {
			return arr, err
		}

		arr = append(arr, rId)
	}

	return arr, err
}

func SelectServiceAccount(params map[string][]string, realm string) ([]models.GetServiceAccountInfo, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	var rows *sql.Rows
	var err error

	query := `select 
	U.ID
	, U.ENABLED
	, U.USERNAME
	, E.CLIENT_ID
	, ISNULL(TRIM(', ' FROM A.Roles), '') as Roles 
	, ISNULL(TRIM(', ' FROM D.Account), '') as Account 
	, ISNULL(TRIM(', ' FROM D.AccountId), '') as AccountId 
	, FORMAT(U.createDate, 'yyyy-MM-dd HH:mm') as createDate
	from
	USER_ENTITY U
    JOIN CLIENT E
    ON U.SERVICE_ACCOUNT_CLIENT_LINK = E.ID
	left outer join 
	(select u.ID, 
	', '+string_agg(r.rName, ', ')+', ' as Roles
	from roles r 
	join user_roles_mapping ur 
	on r.rId = ur.rId
	join USER_ENTITY u
	on ur.userId = u.ID
	GROUP BY u.ID) A
	ON U.ID = A.ID
	left outer join 
	(select 
	USER_ID, ISNULL(', '+string_agg(IDENTITY_PROVIDER, ', ')+', ', '') as openid
	from
	FEDERATED_IDENTITY
	group by USER_ID
	) C
	ON U.ID = C.USER_ID
	LEFT OUTER JOIN USER_ENTITY u1
	on U.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2
	on U.modifyId = u2.ID
	left outer join 
	(SELECT
	ISNULL(', '+string_agg(AC.AccountName, ', ')+', ', '') as Account, AU.UserId,
	ISNULL(', '+string_agg(AC.AccountId, ', ')+', ', '') as AccountId
	FROM [Sale].[dbo].[Account_User] AU
	JOIN [Sale].[dbo].[Account] AC
	ON AU.AccountId = AC.AccountId
	WHERE AU.IsUse = 1
	group by AU.UserId
	) D
	ON U.ID = D.UserId
	WHERE
	U.REALM_ID = ?
	AND U.SERVICE_ACCOUNT_CLIENT_LINK IS NOT NULL `

	queryParams := []interface{}{realm}

	for key, values := range params {
		query += " AND ("

		for i, q := range values {
			if i != 0 {
				query += " OR "
			}
			query += key
			//정확히 일치해야 검색이 되는 종류의 검색 파라미터
			if key == "D.AccountId" || key == "A.Roles" || key == "D.Account" || key == "C.openid" {
				query += " LIKE (?) "
				queryParams = append(queryParams, "%, "+q+",%")
				//정확히 일치하지 않아도 검색이 되는 종류의 검색 파라미터
			} else {
				query += " LIKE (?) "
				queryParams = append(queryParams, "%"+q+"%")
			}
		}

		query += ")"
	}

	query += " ORDER BY U.USERNAME"

	rows, err = db.Query(query, queryParams...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.GetServiceAccountInfo, 0)

	for rows.Next() {
		var r models.GetServiceAccountInfo

		err := rows.Scan(&r.ID, &r.Enabled, &r.Username, &r.ClientId, &r.Roles, &r.Account, &r.AccountId, &r.CreateDate)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}
	return arr, err
}
