package iamdb

import (
	"database/sql"
	"errors"
	"iam/models"
	"strings"
)

func GetUsers(params map[string][]string) ([]models.GetUserInfo, error) {
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
	, U.FIRST_NAME
	, U.LAST_NAME
	, U.EMAIL
	, U.REALM_ID
	, U.PhoneNumber
	, ISNULL(TRIM(', ' FROM A.Roles), '') as Roles 
	, ISNULL(TRIM(', ' FROM B.Groups), '') as Groups 
	, ISNULL(TRIM(', ' FROM D.Account), '') as Account 
	, ISNULL(TRIM(', ' FROM D.AccountId), '') as AccountId 
	, ISNULL(TRIM(', ' FROM C.openid), '') as openid 
	, FORMAT(U.createDate, 'yyyy-MM-dd HH:mm') as createDate
	, u1.USERNAME as Creator
	, FORMAT(U.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate
	, u2.USERNAME as Modifier
	from
	USER_ENTITY U
	left outer join 
	(select u.ID, 
	', '+string_agg(r.rName, ', ')+', ' as Roles
	from roles r 
	join UserRole ur 
	on r.rId = ur.RoleId
	join USER_ENTITY u
	on ur.userId = u.ID
	GROUP BY u.ID) A
	ON U.ID = A.ID
	left outer join 
	(select
	u.ID, 
	ISNULL(', '+string_agg(g.NAME, ', ')+', ', '') as Groups
	, ISNULL(', '+string_agg(gu.GROUP_ID, ', ')+', ', '') as GROUP_ID
	from USER_ENTITY u
	left outer join USER_GROUP_MEMBERSHIP gu
	on u.id = gu.USER_ID
	join KEYCLOAK_GROUP g
	on g.ID = gu.GROUP_ID
	GROUP BY u.ID) B
	On U.ID = B.ID
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
	WHERE U.SERVICE_ACCOUNT_CLIENT_LINK is NULL `

	queryParams := []interface{}{}

	for key, values := range params {
		query += " AND ("

		for i, q := range values {
			if i != 0 {
				query += " OR "
			}
			query += key
			//정확히 일치해야 검색이 되는 종류의 검색 파라미터
			if key == "D.AccountId" || key == "A.Roles" || key == "B.Groups" || key == "D.Account" || key == "C.openid" {
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

	var arr = make([]models.GetUserInfo, 0)

	for rows.Next() {
		var r models.GetUserInfo

		err := rows.Scan(&r.ID, &r.Enabled, &r.Username, &r.FirstName, &r.LastName, &r.Email, &r.Realm, &r.PhoneNumber, &r.Roles, &r.Groups, &r.Account, &r.AccountId, &r.OpenId, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}
	return arr, err
}

func GetUserDetail(userId, realm string) ([]models.GetUserInfo, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	query := `SELECT U.ID, U.ENABLED, U.USERNAME, U.FIRST_NAME, U.LAST_NAME, U.EMAIL, U.PhoneNumber, 
	(SELECT STRING_AGG(REQUIRED_ACTION, ',') FROM USER_REQUIRED_ACTION WHERE USER_ID=U.ID) as REQUIRED_ACTION,
	FORMAT(U.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(U.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
	FROM USER_ENTITY U
	LEFT OUTER JOIN USER_ENTITY u1
	on U.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2
	on U.modifyId = u2.ID
	WHERE U.ID = ? AND U.REALM_ID = ?`

	rows, err := db.Query(query, userId, realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.GetUserInfo, 0)
	blank := make([]string, 0)

	for rows.Next() {
		var r models.GetUserInfo

		RequiredActions := ""
		err := rows.Scan(&r.ID, &r.Enabled, &r.Username, &r.FirstName, &r.LastName, &r.Email, &r.PhoneNumber, &RequiredActions, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
		if err != nil {
			return nil, err
		}

		if RequiredActions == "" {
			r.RequiredActions = &blank
		} else {
			tmp := strings.Split(RequiredActions, ",")
			r.RequiredActions = &tmp
		}

		arr = append(arr, r)
	}
	return arr, err
}

func UsersCreate(userId, reqUserId string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `UPDATE USER_ENTITY SET 
	createId=?,
	createDate=GETDATE(),
	modifyId=?,
	modifyDate=GETDATE()
	FROM USER_ENTITY A
	where A.ID = ?
	SELECT @@ROWCOUNT`

	rows, err := db.Query(query, reqUserId, reqUserId, userId)
	err = resultErrorCheck(rows)
	return err
}

func UsersUpdate(userId, phoneNumber, reqUserId string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	if phoneNumber != "" {
		query := `UPDATE USER_ENTITY SET 
		PhoneNumber=?,
		modifyId=?,
		modifyDate=GETDATE()
		FROM USER_ENTITY A
		where A.ID = ?
		SELECT @@ROWCOUNT`

		rows, err := db.Query(query, phoneNumber, reqUserId)
		if err != nil {
			return err
		}
		err = resultErrorCheck(rows)
		return err
	} else {
		query := `UPDATE USER_ENTITY SET 
		modifyId=?,
		modifyDate=GETDATE()
		FROM USER_ENTITY A
		where A.ID = ?
		SELECT @@ROWCOUNT`

		rows, err := db.Query(query, reqUserId, userId)
		if err != nil {
			return err
		}
		err = resultErrorCheck(rows)
		return err
	}
}

func CreateUserAddRole(uid, reqUserId string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `INSERT INTO UserRole(userId, rId, createId, createDate, modifyId, modifyDate)
	(SELECT ? as userId, rId, ?, GETDATE(), ?, GETDATE() from roles A where defaultRole = 1)`

	_, err := db.Query(query, uid, reqUserId, reqUserId)
	if err != nil {
		return err
	}

	return nil
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
	(SELECT rId FROM IAM.dbo.UserRole
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

func SelectAccountList(user_id string) ([]int64, error) {
	var arr = make([]int64, 0)

	db, err := DBClient()
	defer db.Close()
	if err != nil {
		return arr, err
	}

	query := `SELECT AccountId FROM SALE.dbo.Account_User WHERE userId = ?`

	rows, err := db.Query(query, user_id)
	if err != nil {
		return arr, err
	}
	defer rows.Close()

	for rows.Next() {
		var aId int64

		err := rows.Scan(&aId)
		if err != nil {
			return arr, err
		}

		arr = append(arr, aId)
	}

	return arr, err
}

func GetAccountUserId(id string) ([]string, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	query := `SELECT Seq 
	FROM Sale.dbo.Account_User AU
	JOIN IAM.dbo.USER_ENTITY U ON AU.UserId = ?`

	rows, err := db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]string, 0)

	if rows.Next() {
		str := ""
		err := rows.Scan(&str)
		if err != nil {
			return nil, err
		}

		arr = append(arr, str)
	}

	return arr, nil
}

func GetUserRealmById(groupId string) (string, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return "", dbErr
	}

	query := `SELECT REALM_ID from USER_ENTITY WHERE ID = ?`

	rows, err := db.Query(query, groupId)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var realm string
		err := rows.Scan(&realm)
		if err != nil {
			return "", err
		}

		return realm, nil
	}
	return "", errors.New("Realm not found")
}
