package iamdb

import (
	"database/sql"
	"iam/models"
)

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
