package iamdb

import (
	"database/sql"
	"errors"
	"iam/config"
	"iam/models"
	"strings"
)

func ConnectionTest() {
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

func GetUserAuthoritiesForEndpoint(username string, realm string, method string, endpoint string) (*sql.Rows, error) {
	query := `select
	1
	from
	USER_ENTITY u join
	user_roles_mapping ur
	on
	u.id = ur.userId
	join
	roles_authority_mapping ra
	on
	ur.rId = ra.rId
	join
	authority a
	on
	ra.aId = a.aId
	where
	u.USERNAME = ? AND
	u.REALM_ID = ? AND
	(a.method = ? OR a.method = 'ALL') AND
	(
		PATINDEX(REPLACE(a.url,'*','%%'), ?) = 1
		OR
		PATINDEX(REPLACE(a.url,'*','%%'), ?) = 1
	)
	`

	rows, err := db.Query(query, username, realm, method, endpoint, endpoint+"/")
	return rows, err
}

func GetRoles() ([]models.RolesInfo, error) {
	query := `select r.rId, r.rName, r.defaultRole,
	FORMAT(r.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(r.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
	from roles r
	LEFT OUTER JOIN USER_ENTITY u1
	on r.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2
	on r.modifyId = u2.ID
	where r.REALM_ID = ?`

	rows, err := db.Query(query, config.GetConfig().Keycloak_realm)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.RolesInfo, 0)

	for rows.Next() {
		var r models.RolesInfo

		err := rows.Scan(&r.ID, &r.Name, &r.DefaultRole, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}

	return arr, nil
}

func GetRoleIdByName(rolename string) (string, error) {
	query := `select rId from roles
	where rName = ?
	AND REALM_ID = ?`

	rows, err := db.Query(query, rolename, config.GetConfig().Keycloak_realm)

	if err != nil {
		return "", err
	}
	defer rows.Close()

	var id string

	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return "", err
		}
	}

	return id, nil
}

func GetAuthIdByName(authname string) (string, error) {
	query := `select aId from authority
	where aName = ?
	AND REALM_ID = ?`

	rows, err := db.Query(query, authname, config.GetConfig().Keycloak_realm)

	if err != nil {
		return "", err
	}
	defer rows.Close()

	var id string

	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return "", err
		}
	}

	return id, nil
}

func CreateRoles(role *models.RolesInfo, username string) error {
	query := `INSERT INTO roles(rName, defaultRole, REALM_ID, createId, modifyId) 
	select ?, ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	var DefaultRole bool = false
	if role.DefaultRole != nil && *role.DefaultRole == true {
		DefaultRole = true
	}

	_, err := db.Query(query, role.Name, DefaultRole, config.GetConfig().Keycloak_realm, username, config.GetConfig().Keycloak_realm)
	return err
}

func CreateRolesIdTx(tx *sql.Tx, id string, name string, username string) error {
	query := `INSERT INTO roles(rId, rName, REALM_ID, createId, modifyId) 
	select ?, ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := tx.Query(query, id, name, config.GetConfig().Keycloak_realm, username, config.GetConfig().Keycloak_realm)
	return err
}

func DeleteRolesTx(tx *sql.Tx, id string) error {
	query := `DELETE roles where rId = ? AND REALM_ID = ?`

	_, err := tx.Query(query, id, config.GetConfig().Keycloak_realm)
	return err
}

func DeleteRolesNameTx(tx *sql.Tx, name string) error {
	query := `DELETE roles where rName = ? AND REALM_ID = ?`

	_, err := tx.Query(query, name, config.GetConfig().Keycloak_realm)
	return err
}

func UpdateRolesTx(tx *sql.Tx, role *models.RolesInfo, username string) error {
	var err error

	//버그가 있는듯... db.Query에 nil 을 넣었을 때 IsNull 의 동작이 이상하다...
	//어쩔 수 없이 쿼리 2개로 나눠놓음
	if role.Name != nil {
		query := `UPDATE roles SET 
		rName = ?, 
		defaultRole = IsNull(?, defaultRole), 
		modifyDate=GETDATE(), 
		modifyId=(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) 
		where rId = ? AND REALM_ID = ?`
		_, err = tx.Query(query, role.Name, role.DefaultRole, username, config.GetConfig().Keycloak_realm, role.ID, config.GetConfig().Keycloak_realm)
	} else {
		query := `UPDATE roles SET 
		defaultRole = ?, 
		modifyDate=GETDATE(), 
		modifyId=(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) 
		where rId = ? AND REALM_ID = ?`
		_, err = tx.Query(query, role.DefaultRole, username, config.GetConfig().Keycloak_realm, role.ID, config.GetConfig().Keycloak_realm)
	}

	return err
}

func GetRolseAuth(id string) ([]models.RolesInfo, error) {
	query := `select
	a.aId, a.aName, ra.useYn, 
	FORMAT(ra.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(ra.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
	from 
	roles_authority_mapping ra 
	join 
	authority a 
	on 
	ra.aId = a.aId 
	LEFT OUTER JOIN USER_ENTITY u1
	on ra.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2
	on ra.modifyId = u2.ID
	where 
	ra.rId = ?`

	rows, err := db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.RolesInfo, 0)

	for rows.Next() {
		var r models.RolesInfo

		err := rows.Scan(&r.ID, &r.Name, &r.Use, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}

	return arr, err
}

func AssignRoleAuth(roleID string, authID string, username string) error {
	query := `INSERT INTO roles_authority_mapping(rId, aId, createId, modifyId)
	select ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := db.Query(query, roleID, authID, username, config.GetConfig().Keycloak_realm)
	return err
}

func AssignRoleAuthTx(tx *sql.Tx, roleID string, authID string, username string) error {
	query := `INSERT INTO roles_authority_mapping(rId, aId, createId, modifyId)
	select ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := tx.Query(query, roleID, authID, username, config.GetConfig().Keycloak_realm)
	return err
}

func DismissRoleAuth(roleID string, authID string) error {
	query := `DELETE FROM roles_authority_mapping where rId = ? AND aId = ?`

	_, err := db.Query(query, roleID, authID)
	return err
}

func UpdateRoleAuth(roleID string, authID string, use bool, username string) error {
	query := `UPDATE roles_authority_mapping SET 
	useYn = ?, 
	modifyDate=GETDATE(), 
	modifyId=(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?)
	where rId = ? 
	AND aId = ?`

	_, err := db.Query(query, use, username, config.GetConfig().Keycloak_realm, roleID, authID)
	return err
}

func GetUserRole(userID string) ([]models.RolesInfo, error) {
	query := `select r.rId, r.rName, r.defaultRole, ur.useYn,
	FORMAT(ur.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(ur.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
	from 
	roles r join
	user_roles_mapping ur 
	on r.rId = ur.rId
	LEFT OUTER JOIN USER_ENTITY u1
	on ur.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2
	on ur.modifyId = u2.ID
	where
	ur.userId = ?`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.RolesInfo, 0)

	for rows.Next() {
		var r models.RolesInfo

		err := rows.Scan(&r.ID, &r.Name, &r.DefaultRole, &r.Use, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}
	return arr, err
}

func AssignUserRole(userID string, roleID string, username string) error {
	query := `INSERT INTO user_roles_mapping(userId, rId, createId, modifyId)
	select ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := db.Query(query, userID, roleID, username, config.GetConfig().Keycloak_realm)
	return err
}

func AssignUserRoleTx(tx *sql.Tx, userID string, roleID string, username string) error {
	query := `INSERT INTO user_roles_mapping(userId, rId, createId, modifyId)
	select ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := tx.Query(query, userID, roleID, username, config.GetConfig().Keycloak_realm)
	return err
}

func DismissUserRole(userID string, roleID string) error {
	query := `DELETE FROM user_roles_mapping where userId = ? AND rId = ?`

	_, err := db.Query(query, userID, roleID)
	return err
}

func DeleteUserRoleByRoleNameTx(tx *sql.Tx, roleName string) error {
	query := `DELETE FROM user_roles_mapping where 
	rId = (select rId from roles where rName = ? AND REALM_ID = ?)`

	_, err := tx.Query(query, roleName, config.GetConfig().Keycloak_realm)
	return err
}

func DeleteUserRoleByRoleIdTx(tx *sql.Tx, roleName string) error {
	query := `DELETE FROM user_roles_mapping where rId = ?`

	_, err := tx.Query(query, roleName)
	return err
}

func UpdateUserRole(userID string, roleID string, use bool, username string) error {
	query := `UPDATE user_roles_mapping SET 
	useYn = ?, 
	modifyDate=GETDATE(), 
	modifyId=(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) 
	where userId = ? AND rId = ?`

	_, err := db.Query(query, use, username, config.GetConfig().Keycloak_realm, userID, roleID)
	return err
}

func GetUserAuth(userID string) ([]models.AutuhorityInfo, error) {
	query := `select a.aId, a.aName, a.url, a.method, 
	FORMAT(a.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(a.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
	from user_roles_mapping ur 
	join roles_authority_mapping ra 
	on ur.rId = ra.rId
	join authority a 
	on ra.aId = a.aId
	LEFT OUTER JOIN USER_ENTITY u1
	on a.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2
	on a.modifyId = u2.ID
	where userId = ?
	and	ur.useYn = 'true'
	and	ra.useYn = 'true'
	AND a.REALM_ID = ?`

	rows, err := db.Query(query, userID, config.GetConfig().Keycloak_realm)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.AutuhorityInfo, 0)

	for rows.Next() {
		var r models.AutuhorityInfo

		err := rows.Scan(&r.ID, &r.Name, &r.URL, &r.Method, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}

	return arr, nil
}

func GetUserAuthActive(userName string, authName string) (map[string]interface{}, error) {
	query := `select 1
	from USER_ENTITY u
	join user_roles_mapping ur 
	on u.ID = ur.userId
	join roles_authority_mapping ra 
	on ur.rId = ra.rId
	join authority a 
	on ra.aId = a.aId
	where u.USERNAME = ?
	AND a.aName = ?
	and	ur.useYn = 'true'
	and	ra.useYn = 'true'
	AND u.REALM_ID = ?`

	rows, err := db.Query(query, userName, authName, config.GetConfig().Keycloak_realm)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]interface{})
	if rows.Next() {
		m["active"] = true
	} else {
		m["active"] = false
	}

	return m, nil
}

func GetAuth() ([]models.AutuhorityInfo, error) {
	query := `select a.aId, a.aName, a.url, a.method, 
	FORMAT(a.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(a.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
	from authority a
	LEFT OUTER JOIN USER_ENTITY u1
	on a.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2
	on a.modifyId = u2.ID
	where a.REALM_ID = ?`

	rows, err := db.Query(query, config.GetConfig().Keycloak_realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.AutuhorityInfo, 0)

	for rows.Next() {
		var r models.AutuhorityInfo

		err := rows.Scan(&r.ID, &r.Name, &r.URL, &r.Method, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}
	return arr, err
}

func CreateAuth(auth *models.AutuhorityInfo, username string) error {
	query := `INSERT INTO authority(aId, aName, url, method, REALM_ID, createId, modifyId)
	select ?, ?, ?, ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := db.Query(query, auth.ID, auth.Name, auth.URL, auth.Method, config.GetConfig().Keycloak_realm, username, config.GetConfig().Keycloak_realm)
	return err
}

func CreateAuthIdTx(tx *sql.Tx, id string, name string, url string, method string, username string) error {
	query := `INSERT INTO authority(aId, aName, url, method, REALM_ID, createId, modifyId)
	select ?, ?, ?, ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := tx.Query(query, id, name, url, method, config.GetConfig().Keycloak_realm, username, config.GetConfig().Keycloak_realm)
	return err
}

func DeleteAuth(id string, tx *sql.Tx) error {
	query := `DELETE authority where aId = ? AND REALM_ID = ?`

	_, err := tx.Query(query, id, config.GetConfig().Keycloak_realm)
	return err
}

func DeleteAuthNameTx(tx *sql.Tx, name string) error {
	query := `DELETE authority where aName = ? AND REALM_ID = ?`

	_, err := tx.Query(query, name, config.GetConfig().Keycloak_realm)
	return err
}

func UpdateAuth(auth *models.AutuhorityInfo, username string) error {
	query := `UPDATE authority SET 
	aName = ?, 
	url = ?, 
	method = ?, 
	modifyDate=GETDATE(), 
	modifyId=(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) 
	where aId = ?`

	_, err := db.Query(query, auth.Name, auth.URL, auth.Method, username, config.GetConfig().Keycloak_realm, auth.ID)
	return err
}

func GetAuthInfo(authID string) (*models.AutuhorityInfo, error) {
	query := `select a.aId, a.aName, a.url, a.method, 
	FORMAT(a.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(a.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
	from authority a
	LEFT OUTER JOIN USER_ENTITY u1
	on a.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2
	on a.modifyId = u2.ID
	where a.aId = ? AND a.REALM_ID = ?`

	rows, err := db.Query(query, authID, config.GetConfig().Keycloak_realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var r *models.AutuhorityInfo = new(models.AutuhorityInfo)

	if rows.Next() {
		err := rows.Scan(&r.ID, &r.Name, &r.URL, &r.Method, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
		if err != nil {
			return nil, err
		}
	}
	return r, err
}

func DeleteRolesAuthByRoleIdTx(tx *sql.Tx, id string) error {
	query := `DELETE roles_authority_mapping where rId = ?`

	_, err := tx.Query(query, id)
	return err
}

func DeleteRolesAuthByAuthIdTx(tx *sql.Tx, id string) error {
	query := `DELETE roles_authority_mapping where aId = ?`

	_, err := tx.Query(query, id)
	return err
}

func DeleteRolesAuthByAuthNameTx(tx *sql.Tx, roleName string) error {
	query := `DELETE roles_authority_mapping where aId =
	(select aId from authority where aName = ? AND REALM_ID = ?)`

	_, err := tx.Query(query, roleName, config.GetConfig().Keycloak_realm)
	return err
}

func DeleteUserRoleByUserId(id string) error {
	query := `DELETE user_roles_mapping where userId = ?`

	_, err := db.Query(query, id)
	return err
}

func CheckRoleAuthID(roleID string, authID string) error {
	query := `select count(*) as result
	from
	(
	select rid as id from roles where rId = ? AND REALM_ID = ?
	union 
	select aid as id from authority where aId = ? AND REALM_ID = ?
	) a`

	rows, err := db.Query(query, roleID, config.GetConfig().Keycloak_realm, authID, config.GetConfig().Keycloak_realm)
	if err != nil {
		return err
	}

	var r int
	if rows.Next() {
		err := rows.Scan(&r)
		if err != nil {
			return err
		}
	}

	if r != 2 {
		return errors.New("Bad Request")
	}

	return nil
}

func CheckUserRoleID(userID string, roleID string) error {
	query := `select count(*) as result
	from
	(
	select ID as id from USER_ENTITY where ID = ? AND REALM_ID = ?
	union 
	select rId as id from roles where rId = ? AND REALM_ID = ?
	) a`

	rows, err := db.Query(query, userID, config.GetConfig().Keycloak_realm, roleID, config.GetConfig().Keycloak_realm)
	if err != nil {
		return err
	}

	var r int
	if rows.Next() {
		err := rows.Scan(&r)
		if err != nil {
			return err
		}
	}

	if r != 2 {
		return errors.New("Bad Request")
	}

	return nil
}

func GetGroup() ([]models.GroupItem, error) {
	query := `SELECT g.ID, NAME, 
	ISNULL((select count(USER_ID) from USER_GROUP_MEMBERSHIP where GROUP_ID = g.ID AND g.REALM_ID = ? group by GROUP_ID), 0) as countMembers,
	FORMAT(g.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(g.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
	from KEYCLOAK_GROUP g
	LEFT OUTER JOIN USER_ENTITY u1
	on g.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2
	on g.modifyId = u2.ID
	where
	g.REALM_ID = ?`

	rows, err := db.Query(query, config.GetConfig().Keycloak_realm, config.GetConfig().Keycloak_realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.GroupItem, 0)

	for rows.Next() {
		var r models.GroupItem

		err := rows.Scan(&r.ID, &r.Name, &r.CountMembers, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}
	return arr, err
}

func GroupCreate(groupId string, username string) error {
	query := `UPDATE KEYCLOAK_GROUP SET 
	createId=B.ID,
	createDate=GETDATE(),
	modifyId=B.ID,
	modifyDate=GETDATE()
	FROM KEYCLOAK_GROUP A,
	(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) B
	where A.ID = ?`

	_, err := db.Query(query, username, config.GetConfig().Keycloak_realm, groupId)
	return err
}

func GroupUpdate(groupId string, username string) error {
	query := `UPDATE KEYCLOAK_GROUP SET 
	modifyId=B.ID,
	modifyDate=GETDATE()
	FROM KEYCLOAK_GROUP A,
	(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) B
	where A.ID = ?`

	_, err := db.Query(query, username, config.GetConfig().Keycloak_realm, groupId)
	return err
}

func GetUsers(search string, groupid string) ([]models.GetUserInfo, error) {
	var rows *sql.Rows
	var err error

	query := `select 
	  U.ID
	, U.ENABLED
	, U.USERNAME
	, U.FIRST_NAME
	, U.LAST_NAME
	, U.EMAIL
	, ISNULL(A.Roles, '') as Roles 
	, ISNULL(B.Groups, '') as Groups 
	, ISNULL(C.openid, '') as openid 
	, FORMAT(U.createDate, 'yyyy-MM-dd HH:mm') as createDate
	, u1.USERNAME as Creator
	, FORMAT(U.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate
	, u2.USERNAME as Modifier
	from
	USER_ENTITY U
	left outer join 
	(select u.ID, 
	string_agg(r.rName, ', ') as Roles
	from roles r 
	join user_roles_mapping ur 
	on r.rId = ur.rId
	join USER_ENTITY u
	on ur.userId = u.ID
	GROUP BY u.ID) A
	ON U.ID = A.ID
	left outer join 
	(select
	u.ID, 
	ISNULL(string_agg(g.NAME, ', '), '') as Groups
	from USER_ENTITY u
	left outer join USER_GROUP_MEMBERSHIP gu
	on u.id = gu.USER_ID
	join KEYCLOAK_GROUP g
	on g.ID = gu.GROUP_ID
	GROUP BY u.ID) B
	On U.ID = B.ID
	left outer join 
	(select 
	USER_ID, ISNULL(string_agg(IDENTITY_PROVIDER, ', '), '') as openid
	from
	FEDERATED_IDENTITY
	group by USER_ID
	) C
	ON U.ID = C.USER_ID
	LEFT OUTER JOIN USER_ENTITY u1
	on U.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2
	on U.modifyId = u2.ID`

	if groupid != "" {
		query += ` join USER_GROUP_MEMBERSHIP UG
		ON U.ID = UG.USER_ID`
	}

	query += ` WHERE
	U.REALM_ID = ?
	AND U.SERVICE_ACCOUNT_CLIENT_LINK is NULL `

	if search != "" {
		query += " AND U.USERNAME LIKE ?"
	}

	if groupid != "" {
		query += ` AND UG.GROUP_ID = ?`
	}

	//나중에 방법을 찾아서 정리하는걸로...
	if search != "" && groupid != "" {
		rows, err = db.Query(query, config.GetConfig().Keycloak_realm, "%"+search+"%", groupid)
	} else if search != "" {
		rows, err = db.Query(query, config.GetConfig().Keycloak_realm, "%"+search+"%")
	} else if groupid != "" {
		rows, err = db.Query(query, config.GetConfig().Keycloak_realm, groupid)
	} else {
		rows, err = db.Query(query, config.GetConfig().Keycloak_realm)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.GetUserInfo, 0)

	for rows.Next() {
		var r models.GetUserInfo

		err := rows.Scan(&r.ID, &r.Enabled, &r.Username, &r.FirstName, &r.LastName, &r.Email, &r.Roles, &r.Groups, &r.OpenId, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}
	return arr, err
}

func GetUserDetail(userId string) ([]models.GetUserInfo, error) {
	query := `SELECT U.ID, U.ENABLED, U.USERNAME, U.FIRST_NAME, U.LAST_NAME, U.EMAIL, 
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

	rows, err := db.Query(query, userId, config.GetConfig().Keycloak_realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.GetUserInfo, 0)
	blank := make([]string, 0)

	for rows.Next() {
		var r models.GetUserInfo

		RequiredActions := ""
		err := rows.Scan(&r.ID, &r.Enabled, &r.Username, &r.FirstName, &r.LastName, &r.Email, &RequiredActions, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
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

func UsersCreate(userId string, username string) error {
	query := `UPDATE USER_ENTITY SET 
	createId=B.ID,
	createDate=GETDATE(),
	modifyId=B.ID,
	modifyDate=GETDATE()
	FROM USER_ENTITY A,
	(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) B
	where A.ID = ?`

	_, err := db.Query(query, username, config.GetConfig().Keycloak_realm, userId)
	return err
}

func UsersUpdate(userId string, username string) error {
	query := `UPDATE USER_ENTITY SET 
	modifyId=B.ID,
	modifyDate=GETDATE()
	FROM USER_ENTITY A,
	(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) B
	where A.ID = ?`

	_, err := db.Query(query, username, config.GetConfig().Keycloak_realm, userId)
	return err
}

func CreateSecretGroupTx(tx *sql.Tx, secretGroupPath string, username string) error {
	query := `INSERT INTO vSecretGroup(vSecretGroupPath, REALM_ID, createId, modifyId)
	select ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := tx.Query(query, secretGroupPath, config.GetConfig().Keycloak_realm, username, config.GetConfig().Keycloak_realm)
	return err
}

func DeleteSecretGroupTx(tx *sql.Tx, secretGroupPath string) error {
	query := `DELETE FROM vSecretGroup WHERE vSecretGroupPath = ? AND REALM_ID = ?`

	_, err := tx.Query(query, secretGroupPath, config.GetConfig().Keycloak_realm)
	return err
}

func MergeSecret(secretPath string, secretGroupPath string, username string) error {
	query := `MERGE INTO vSecret A
	USING (SELECT 
	? as spath, 
	(select vSecretGroupId from vSecretGroup where vSecretGroupPath = ? AND REALM_ID = ?) as sgid,
	(select ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) as userid
	) B
	ON A.vSecretPath = B.spath
	AND A.vSecretGroupId = B.sgid
	WHEN MATCHED THEN
		UPDATE SET 
		modifyDate = GETDATE(),
		modifyId = B.userid
	WHEN NOT MATCHED THEN
		INSERT (vSecretPath, vSecretGroupId, createId, modifyId)
		VALUES(B.spath, B.sgid, B.userid, B.userid);`

	_, err := db.Query(query, secretPath, secretGroupPath, config.GetConfig().Keycloak_realm, username, config.GetConfig().Keycloak_realm)

	return err
}

func DeleteSecret(secretPath string, secretGroupPath string) error {
	query := `DELETE FROM vSecret WHERE vSecretPath = ?
	AND vSecretGroupId = (select vSecretGroupId from vSecretGroup where vSecretGroupPath = ?)`

	_, err := db.Query(query, secretPath, secretGroupPath)
	return err
}

func GetSecretGroup(data []models.SecretGroupItem, username string) ([]models.SecretGroupItem, error) {
	query := `declare @values table
	(
		sg varchar(310)
	)`
	for _, d := range data {
		query += "insert into @values values ('/secret/" + d.Name + "/')"
	}
	query += `select C.secretGroup, 
	FORMAT(D.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(D.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
	from (select REPLACE(REPLACE(B.sg,'/secret/',''),'/','') as secretGroup 
	from (select
	REPLACE(url,'*','%%') as auth_url
	from USER_ENTITY u
	join user_roles_mapping ur on u.ID = ur.userId
	join roles_authority_mapping ra on ur.rId = ra.rId
	join authority a on ra.aId = a.aId
	where ur.useYn = 'true'
	and ra.useYn = 'true'
	and u.USERNAME = ?
	and u.REALM_ID = ?
	) A
	join @values B
	ON PATINDEX(A.auth_url, B.sg) = 1
	group by B.sg) C
	left outer join
	vSecretGroup D
	LEFT OUTER JOIN USER_ENTITY u1
	on D.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2
	on D.modifyId = u2.ID
	ON C.secretGroup = D.vSecretGroupPath`

	rows, err := db.Query(query, username, config.GetConfig().Keycloak_realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	arr := make([]models.SecretGroupItem, 0)

	for rows.Next() {
		var r models.SecretGroupItem

		err := rows.Scan(&r.Name, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
		if err != nil {
			return nil, err
		}

		for _, d := range data { //이게 최선인듯...
			if d.Name == r.Name {
				r.Description = d.Description
				break
			}
		}

		arr = append(arr, r)
	}
	return arr, err
}

func GetSecret(groupName string) (map[string]models.SecretItem, error) {
	query := `SELECT s.vSecretPath, 
	FORMAT(s.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(s.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
	FROM vSecret s
	LEFT OUTER JOIN USER_ENTITY u1
	on s.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2
	on s.modifyId = u2.ID
	WHERE s.vSecretGroupId = (SELECT vSecretGroupId FROM vSecretGroup WHERE vSecretGroupPath = ? AND REALM_ID = ?)`

	rows, err := db.Query(query, groupName, config.GetConfig().Keycloak_realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var m = make(map[string]models.SecretItem)

	for rows.Next() {
		var r models.SecretItem

		err := rows.Scan(&r.Name, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
		if err != nil {
			return nil, err
		}

		m[r.Name] = r
	}
	return m, err
}

func GetSecretGroupMetadata(groupName string) (models.SecretGroupResponse, error) {
	var result models.SecretGroupResponse

	result.Roles = make([]models.IdItem, 0)
	result.Users = make([]models.IdItem, 0)

	query := `SELECT 
	r.rId, r.rName
	FROM roles r
	JOIN roles_authority_mapping ra
	on r.rId = ra.rId
	join authority a
	on ra.aId = a.aId
	where a.aName = ?
	AND r.REALM_ID = ?`

	rows, err := db.Query(query, groupName+"_MANAGER", config.GetConfig().Keycloak_realm)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.IdItem

		err := rows.Scan(&r.Id, &r.Name)
		if err != nil {
			return result, err
		}

		result.Roles = append(result.Roles, r)
	}

	query = `SELECT 
	u.ID, u.USERNAME
	FROM USER_ENTITY u
	JOIN user_roles_mapping ur
	on u.ID = ur.userId
	JOIN roles r
	ON ur.rId = r.rId
	where r.rName = ?
	AND r.REALM_ID = ?`

	rows, err = db.Query(query, groupName+"_Manager", config.GetConfig().Keycloak_realm)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.IdItem

		err := rows.Scan(&r.Id, &r.Name)
		if err != nil {
			return result, err
		}

		result.Users = append(result.Users, r)
	}

	query = `SELECT 
	FORMAT(g.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(g.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
	FROM vSecretGroup g
	LEFT OUTER JOIN USER_ENTITY u1
	on g.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2
	on g.modifyId = u2.ID
	where vSecretGroupPath = ?
	AND g.REALM_ID = ?`

	rows, err = db.Query(query, groupName, config.GetConfig().Keycloak_realm)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&result.CreateDate, &result.Creator, &result.ModifyDate, &result.Modifier)
		if err != nil {
			return result, err
		}
	}

	return result, err
}

func GetSecretByName(groupName string, secretName string) (*models.SecretItem, error) {
	query := `SELECT s.vSecretPath, 
	FORMAT(s.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(s.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
	FROM vSecret s
	LEFT OUTER JOIN USER_ENTITY u1
	on s.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2
	on s.modifyId = u2.ID
	WHERE s.vSecretGroupId = (SELECT vSecretGroupId FROM vSecretGroup WHERE vSecretGroupPath = ? AND REALM_ID = ?) AND vSecretPath = ?`

	rows, err := db.Query(query, groupName, config.GetConfig().Keycloak_realm, secretName)
	if err != nil {
		return nil, err
	}

	m := new(models.SecretItem)

	rows.Next()
	err = rows.Scan(&m.Name, &m.CreateDate, &m.Creator, &m.ModifyDate, &m.Modifier)
	if err != nil {
		return m, nil
	}
	defer rows.Close()

	return m, err
}

func MetricCount() (map[string]int, error) {
	query := `select 
	(select count(*) from USER_ENTITY where REALM_ID = ? AND SERVICE_ACCOUNT_CLIENT_LINK is NULL) AS users,
	(select count(*) from KEYCLOAK_GROUP where REALM_ID = ?) AS groups,
	(select count(*) from CLIENT where REALM_ID = ? AND NODE_REREG_TIMEOUT = -1) AS applicastions,
	(select count(*) from roles where REALM_ID = ?) AS roles,
	(select count(*) from authority where REALM_ID = ?) AS authorities`

	rows, err := db.Query(query,
		config.GetConfig().Keycloak_realm,
		config.GetConfig().Keycloak_realm,
		config.GetConfig().Keycloak_realm,
		config.GetConfig().Keycloak_realm,
		config.GetConfig().Keycloak_realm)
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

func GetApplications() ([]string, error) {
	query := `select CLIENT_ID from CLIENT where REALM_ID = ? AND NODE_REREG_TIMEOUT = -1 AND CLIENT_ID != ?`

	rows, err := db.Query(query, config.GetConfig().Keycloak_realm, config.GetConfig().Keycloak_client_id)
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

func GetLoginApplication(date int) ([]models.MetricItem, error) {
	query := `select B.CLIENT_ID
	, count(A.CLIENT_ID) as count
	from
	(select * FROM
	(SELECT
	CLIENT_ID, DATEADD(SECOND, EVENT_TIME/1000, '01/01/1970 09:00:00') as etime
	FROM EVENT_ENTITY
	where CLIENT_ID != ?
	AND TYPE = 'LOGIN'
	AND REALM_ID = ?
	) AA 
	where AA.etime > getdate()-?
	) A
	RIGHT OUTER JOIN
	(select CLIENT_ID from CLIENT
	where
	REALM_ID = ?
	AND NODE_REREG_TIMEOUT = -1
	AND CLIENT_ID != ?
	) B
	ON A.CLIENT_ID = B.CLIENT_ID
	group by B.client_id`

	rows, err := db.Query(query,
		config.GetConfig().Keycloak_client_id,
		config.GetConfig().Keycloak_realm,
		date,
		config.GetConfig().Keycloak_realm,
		config.GetConfig().Keycloak_client_id)

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

func GetLoginDate(date int) ([]models.MetricItem, error) {
	query := `select 
	FORMAT(C.SYSTEM_DATE, 'yyyy-MM-dd') as date,
	count(B.ID) as count
	from
	(
	select
	CONVERT(DATE, etime) as eDate,
	*
	from
	(SELECT
	E.ID,
	DATEADD(SECOND, EVENT_TIME/1000, '01/01/1970 09:00:00') as etime
	FROM EVENT_ENTITY E
	JOIN
	(select CLIENT_ID from CLIENT
	where
	REALM_ID = ?
	AND NODE_REREG_TIMEOUT = -1
	AND CLIENT_ID != ?) D
	ON E.CLIENT_ID = D.CLIENT_ID
	where  TYPE = 'LOGIN'
	) A
	where etime > getdate()-?
	) B
	right outer JOIN 
	(
	SELECT CONVERT(DATE, DATEADD(DAY, NUMBER, getdate()-?), 112) AS SYSTEM_DATE
	FROM MASTER..SPT_VALUES WITH(NOLOCK)
	WHERE TYPE = 'P'
	AND NUMBER <= DATEDIFF(DAY, getdate()-?, getdate())
	) C
	ON B.eDate = C.SYSTEM_DATE
	group by C.SYSTEM_DATE`

	rows, err := db.Query(query,
		config.GetConfig().Keycloak_realm,
		config.GetConfig().Keycloak_client_id,
		date, date, date)

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

func GetLoginError(date int) ([]models.MetricItem, error) {
	query := `declare @values table
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
	where TYPE = 'LOGIN_ERROR'
	AND REALM_ID = ?
	) AA
	where AA.etime > GETDATE() - ?) A
	RIGHT OUTER JOIN @values B
	ON A.ERROR = B.error
	GROUP BY B.errorMessage`

	rows, err := db.Query(query, config.GetConfig().Keycloak_realm, date)
	if err != nil {
		return nil, err
	}

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

func CreateUserAddRole(uid string, username string) error {
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

	_, err := db.Query(query, username, config.GetConfig().Keycloak_realm, uid)
	if err != nil {
		return err
	}

	return nil
}
