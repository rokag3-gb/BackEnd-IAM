package iamdb

import (
	"fmt"
	"iam/models"
)

func MetricCount(realms []string) (map[string]int, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}
	queryParams := []interface{}{}

	RealmParam := ""
	if len(realms) != 0 {
		RealmParam = "AND REALM_ID IN("
		for _ = range realms {
			RealmParam += "?,"
		}
		RealmParam = RealmParam[0 : len(RealmParam)-1]
		RealmParam += ")"
	}

	query := fmt.Sprintf(`select 
	(select count(*) from USER_ENTITY WHERE 1=1 %s AND SERVICE_ACCOUNT_CLIENT_LINK is NULL) AS users,
	(select count(*) from KEYCLOAK_GROUP where 1=1 %s) AS groups,
	(select count(*) from CLIENT where 1=1 %s AND (NAME IS NULL OR LEN(NAME) = 0)) AS applicastions,
	(select count(*) from roles where 1=1 %s) AS roles,
	(select count(*) from authority where 1=1 %s) AS authorities`, RealmParam, RealmParam, RealmParam, RealmParam, RealmParam)

	for i := 0; i < 5; i++ {
		for _, realm := range realms {
			queryParams = append(queryParams, realm)
		}
	}

	rows, err := db.Query(query, queryParams...)
	if err != nil {
		return nil, err
	}
	users := 0
	groups := 0
	applications := 0
	roles := 0
	authorities := 0

	rows.Next()
	err = rows.Scan(&users, &groups, &applications, &roles, &authorities)
	if err != nil {
		return nil, err
	}

	m := make(map[string]int)
	m["users"] = users
	m["groups"] = groups
	m["applications"] = applications
	m["roles"] = roles
	m["authorities"] = authorities

	return m, nil
}

func GetApplications(realm string) ([]string, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	query := `select CLIENT_ID from CLIENT where REALM_ID = ? AND (NAME IS NULL OR LEN(NAME) = 0)`

	rows, err := db.Query(query, realm)
	if err != nil {
		return nil, err
	}
	arr := make([]string, 0, 10)

	for rows.Next() {
		cid := ""
		err = rows.Scan(&cid)
		if err != nil {
			return nil, err
		}

		arr = append(arr, cid)
	}

	return arr, nil
}

func GetLoginApplication(date int, realms []string) ([]models.MetricItem, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}
	queryParams := []interface{}{}

	RealmParam := ""
	if len(realms) != 0 {
		RealmParam = "AND REALM_ID IN("
		for _ = range realms {
			RealmParam += "?,"
		}
		RealmParam = RealmParam[0 : len(RealmParam)-1]
		RealmParam += ")"
	}

	query := fmt.Sprintf(`select B.CLIENT_ID
	, B.REALM_ID
	, count(A.CLIENT_ID) as count
	from
	(select * FROM
	(SELECT
	CLIENT_ID, DATEADD(SECOND, EVENT_TIME/1000, '01/01/1970 09:00:00') as etime
	FROM EVENT_ENTITY
	where TYPE = 'LOGIN' %s
	) AA 
	where AA.etime > getdate()-?
	) A
	RIGHT OUTER JOIN
	(select CLIENT_ID, REALM_ID from CLIENT
	where (NAME IS NULL OR LEN(NAME) = 0) %s
	) B
	ON A.CLIENT_ID = B.CLIENT_ID
	group by B.client_id, B.REALM_ID`, RealmParam, RealmParam)

	for _, realm := range realms {
		queryParams = append(queryParams, realm)
	}

	queryParams = append(queryParams, date)

	for _, realm := range realms {
		queryParams = append(queryParams, realm)
	}

	rows, err := db.Query(query, queryParams...)

	if err != nil {
		return nil, err
	}
	arr := make([]models.MetricItem, 0)

	for rows.Next() {
		var m models.MetricItem
		err = rows.Scan(&m.Key, &m.Realm, &m.Value)
		if err != nil {
			return nil, err
		}

		arr = append(arr, m)
	}

	return arr, nil
}

func GetLoginApplicationDate(date int, realms []string) ([]map[string]interface{}, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}
	queryParams := []interface{}{}

	RealmParam := ""
	if len(realms) != 0 {
		RealmParam = "AND REALM_ID IN("
		for _ = range realms {
			RealmParam += "?,"
		}
		RealmParam = RealmParam[0 : len(RealmParam)-1]
		RealmParam += ")"
	}

	query := fmt.Sprintf(`select 
	E.CLIENT_ID,
	CONVERT(CHAR(10), E.SYSTEM_DATE, 23),
	ISNULL(B.COUNT, 0) as LOGIN_COUNT
	from
	(
	select 
	A.CLIENT_ID, 
	A.EVENT_DATE,
	count(CLIENT_ID) as COUNT
	from 
	(
	SELECT
	CLIENT_ID,
	CONVERT(DATE, DATEADD(SECOND, EVENT_TIME/1000, '01/01/1970 09:00:00')) as EVENT_DATE
	FROM EVENT_ENTITY
	where  TYPE = 'LOGIN'
	AND JSON_VALUE(DETAILS_JSON, '$.response_mode') IS NULL
	) A
	GROUP BY A.CLIENT_ID, A.EVENT_DATE
	) B
	RIGHT OUTER JOIN
	(
	select * from
	(select CLIENT_ID from CLIENT
	where (NAME IS NULL OR LEN(NAME) = 0) %s
	) C
	join
	(
	SELECT CONVERT(DATE, DATEADD(DAY, NUMBER, getdate()-?), 112) AS SYSTEM_DATE
	FROM MASTER..SPT_VALUES WITH(NOLOCK)
	WHERE TYPE = 'P'
	AND NUMBER <= DATEDIFF(DAY, getdate()-?, getdate())
	) D
	ON 1=1
	) E
	ON B.EVENT_DATE = E.SYSTEM_DATE
	AND B.CLIENT_ID = E.CLIENT_ID
	ORDER BY E.SYSTEM_DATE, E.CLIENT_ID`, RealmParam)

	for _, realm := range realms {
		queryParams = append(queryParams, realm)
	}

	queryParams = append(queryParams, date)
	queryParams = append(queryParams, date)

	rows, err := db.Query(query, queryParams...)

	if err != nil {
		return nil, err
	}

	arr := make([]map[string]interface{}, 0)
	tmpMap := make(map[string]map[string]interface{})

	for rows.Next() {
		var client_id string
		var event_date string
		var login_count int

		err = rows.Scan(&client_id, &event_date, &login_count)
		if err != nil {
			return nil, err
		}

		m, exist := tmpMap[event_date]
		if !exist {
			m = make(map[string]interface{})
			tmpMap[event_date] = m
			arr = append(arr, m)
		}

		m[client_id] = login_count
		m["date"] = event_date
	}

	return arr, nil
}

func GetLoginApplicationLog(date string, realms []string) ([]models.MetricLogItem, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}
	queryParams := []interface{}{}

	RealmParam := ""
	if len(realms) != 0 {
		RealmParam = "AND REALM_ID IN("
		for _ = range realms {
			RealmParam += "?,"
		}
		RealmParam = RealmParam[0 : len(RealmParam)-1]
		RealmParam += ")"
	}

	query := fmt.Sprintf(`SELECT 
	E.CLIENT_ID,
	U.USERNAME,
	CONVERT(NVARCHAR, DATEADD(SECOND, EVENT_TIME/1000, '01/01/1970 09:00:00'), 23) as EVENT_DATE
	FROM EVENT_ENTITY E
	JOIN USER_ENTITY U
	ON E.USER_ID = U.ID
	WHERE E.REALM_ID = %s
	AND TYPE = 'LOGIN'
	AND JSON_VALUE(DETAILS_JSON, '$.response_mode') IS NULL
	AND EVENT_TIME > CAST(DATEDIFF(SECOND,{d '1970-01-01'}, ?) AS BIGINT) * 1000
	ORDER BY EVENT_TIME`, RealmParam)

	for _, realm := range realms {
		queryParams = append(queryParams, realm)
	}

	rows, err := db.Query(query, queryParams...)

	if err != nil {
		return nil, err
	}

	arr := make([]models.MetricLogItem, 0)

	for rows.Next() {
		tmp := models.MetricLogItem{}

		err = rows.Scan(&tmp.ClientId, &tmp.Username, &tmp.EventDate)
		if err != nil {
			return nil, err
		}

		arr = append(arr, tmp)
	}

	return arr, nil
}

func GetLoginError(date int, realms []string) ([]models.MetricItem, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	queryParams := []interface{}{}

	RealmParam := ""
	if len(realms) != 0 {
		RealmParam = "AND REALM_ID IN("
		for _ = range realms {
			RealmParam += "?,"
		}
		RealmParam = RealmParam[0 : len(RealmParam)-1]
		RealmParam += ")"
	}

	query := fmt.Sprintf(`declare @values table
	(
		error varchar(64),
		errorMessage varchar(64)
	)
	insert into @values values ('different_user_authenticated', 'Different user authenticated')
	insert into @values values ('invalid_user_credentials', 'Invalid user credentials')
	insert into @values values ('rejected_by_user', 'Rejected by user')
	insert into @values values ('user_disabled', 'User disabled')
	insert into @values values ('user_not_found', 'User not found')
	
	SELECT 
	B.errorMessage,
	count(a.ERROR)
	FROM 
	(SELECT * from
	(SELECT
	ERROR, DATEADD(SECOND, EVENT_TIME/1000, '01/01/1970 09:00:00') as etime
	FROM EVENT_ENTITY
	where TYPE = 'LOGIN_ERROR' %s) AA
	where AA.etime > GETDATE() - ?) A
	RIGHT OUTER JOIN @values B
	ON A.ERROR = B.error
	GROUP BY B.errorMessage`, RealmParam)

	for _, realm := range realms {
		queryParams = append(queryParams, realm)
	}
	queryParams = append(queryParams, date)

	rows, err := db.Query(query, queryParams...)
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

func GetIdpCount(realms []string) ([]models.MetricItem, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	queryParams := []interface{}{}

	RealmParam := ""
	if len(realms) != 0 {
		RealmParam = "WHERE REALM_ID IN("
		for _ = range realms {
			RealmParam += "?,"
		}
		RealmParam = RealmParam[0 : len(RealmParam)-1]
		RealmParam += ")"
	}

	query := fmt.Sprintf(`select 
	A.PROVIDER_ALIAS, 
	count(B.IDENTITY_PROVIDER) as count
	from IDENTITY_PROVIDER A
	LEFT OUTER JOIN FEDERATED_IDENTITY B
	ON B.IDENTITY_PROVIDER = A.PROVIDER_ALIAS %s
	GROUP BY A.PROVIDER_ALIAS`, RealmParam)

	for _, realm := range realms {
		queryParams = append(queryParams, realm)
	}

	rows, err := db.Query(query, queryParams...)
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
