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
	query := "select rId, rName, createDate, createId, modifyDate, modifyId from roles where REALM_ID = ?"

	rows, err := db.Query(query, config.GetConfig().Keycloak_realm)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.RolesInfo, 0)

	for rows.Next() {
		var r models.RolesInfo

		err := rows.Scan(&r.ID, &r.Name, &r.CreateDate, &r.CreateId, &r.ModifyDate, &r.ModifyId)
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

func CreateRoles(name string, username string) error {
	query := `INSERT INTO roles(rName, REALM_ID, createId, modifyId) 
	select ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := db.Query(query, name, config.GetConfig().Keycloak_realm, username, config.GetConfig().Keycloak_realm)
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

func UpdateRoles(name string, id string, username string) error {
	query := `UPDATE roles SET 
	rName = ?, 
	modifyDate=GETDATE(), 
	modifyId=(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) 
	where rId = ? AND REALM_ID = ?`

	_, err := db.Query(query, name, username, config.GetConfig().Keycloak_realm, id, config.GetConfig().Keycloak_realm)
	return err
}

func GetRolseAuth(id string) ([]models.RolesInfo, error) {
	query := `select
	a.aId, a.aName, ra.useYn, ra.createDate, ra.createId, ra.modifyDate, ra.modifyId
	from 
	roles_authority_mapping ra 
	join 
	authority a 
	on 
	ra.aId = a.aId 
	where 
	ra.rId = ?`

	rows, err := db.Query(query, id)
	defer rows.Close()

	var arr = make([]models.RolesInfo, 0)

	for rows.Next() {
		var r models.RolesInfo

		err := rows.Scan(&r.ID, &r.Name, &r.Use, &r.CreateDate, &r.CreateId, &r.ModifyDate, &r.ModifyId)
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

func UpdateRoleAuth(roleID string, authID string, use string, username string) error {
	query := `UPDATE roles_authority_mapping SET 
	useYn = ?, 
	modifyDate=GETDATE(), 
	modifyId=(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?)
	where rId = ? 
	AND aId = ?`

	_, err := db.Query(query, use, username, config.GetConfig().Keycloak_realm, roleID, authID)
	return err
}

func GetAuthUserList() ([]models.UserRolesInfo, error) {
	query := `select u.ID, 
	u.USERNAME, 
	ISNULL(u.FIRST_NAME, '') as FIRST_NAME, 
	ISNULL(u.LAST_NAME, '') as LAST_NAME, 
	ISNULL(u.EMAIL, '') as EMAIL, 
	ISNULL(string_agg(r.rName, ', '), '') as RoleList
		from roles r 
		join user_roles_mapping ur 
		on r.rId = ur.rId
		right outer join USER_ENTITY u
		on ur.userId = u.ID
	WHERE
		u.SERVICE_ACCOUNT_CLIENT_LINK is NULL
		AND u.REALM_ID = ?
		GROUP BY u.USERNAME, u.EMAIL, u.ID, u.FIRST_NAME, u.LAST_NAME`

	rows, err := db.Query(query, config.GetConfig().Keycloak_realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.UserRolesInfo, 0)

	for rows.Next() {
		var r models.UserRolesInfo

		err := rows.Scan(&r.ID, &r.Username, &r.FirstName, &r.LastName, &r.Email, &r.RoleList)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}

	return arr, nil
}

func GetUserRole(userID string) ([]models.RolesInfo, error) {
	query := `select r.rId, r.rName, ur.useYn, ur.createDate, ur.createId, ur.modifyDate, ur.modifyId
	from 
	roles r join
	user_roles_mapping ur 
	on r.rId = ur.rId
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

		err := rows.Scan(&r.ID, &r.Name, &r.Use, &r.CreateDate, &r.CreateId, &r.ModifyDate, &r.ModifyId)
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
	query := `DELETE FROM user_roles_mapping where 
	rId = ?`

	_, err := tx.Query(query, roleName)
	return err
}

func UpdateUserRole(userID string, roleID string, use string, username string) error {
	query := `UPDATE user_roles_mapping SET 
	useYn = ?, 
	modifyDate=GETDATE(), 
	modifyId=(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) 
	where userId = ? AND rId = ?`

	_, err := db.Query(query, use, username, config.GetConfig().Keycloak_realm, userID, roleID)
	return err
}

func GetUserAuth(userID string) ([]models.AutuhorityInfo, error) {
	query := `select a.aId, a.aName, a.createDate, a.createId, a.modifyDate, a.modifyId
	from user_roles_mapping ur 
	join roles_authority_mapping ra 
	on ur.rId = ra.rId
	join authority a 
	on ra.aId = a.aId
	where userId = ?
	and	ur.useYn = 'y'
	and	ra.useYn = 'y'
	AND a.REALM_ID = ?`

	rows, err := db.Query(query, userID, config.GetConfig().Keycloak_realm)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.AutuhorityInfo, 0)

	for rows.Next() {
		var r models.AutuhorityInfo

		err := rows.Scan(&r.ID, &r.Name, &r.CreateDate, &r.CreateId, &r.ModifyDate, &r.ModifyId)
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
	and	ur.useYn = 'y'
	and	ra.useYn = 'y'
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
	query := `select aId, aName, createDate, createId, modifyDate, modifyId from authority where REALM_ID = ?`

	rows, err := db.Query(query, config.GetConfig().Keycloak_realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.AutuhorityInfo, 0)

	for rows.Next() {
		var r models.AutuhorityInfo

		err := rows.Scan(&r.ID, &r.Name, &r.CreateDate, &r.CreateId, &r.ModifyDate, &r.ModifyId)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}
	return arr, err
}

func CreateAuth(auth *models.AutuhorityInfo, username string) error {
	query := `INSERT INTO authority(aName, url, method, REALM_ID, createId, modifyId)
	select ?, ?, ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := db.Query(query, auth.Name, auth.URL, auth.Method, config.GetConfig().Keycloak_realm, username, config.GetConfig().Keycloak_realm)
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
	query := `select aId, aName, url, method, createDate, createId, modifyDate, modifyId from authority where aId = ? AND REALM_ID = ?`

	rows, err := db.Query(query, authID, config.GetConfig().Keycloak_realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var r *models.AutuhorityInfo = new(models.AutuhorityInfo)

	if rows.Next() {
		err := rows.Scan(&r.ID, &r.Name, &r.URL, &r.Method, &r.CreateDate, &r.CreateId, &r.ModifyDate, &r.ModifyId)
		if err != nil {
			return nil, err
		}
	}
	return r, err
}

func DeleteRolesAuthByRoleId(tx *sql.Tx, id string) error {
	query := `DELETE roles_authority_mapping where rId = ?`

	_, err := tx.Query(query, id)
	return err
}

func DeleteRolesAuthByAuthId(tx *sql.Tx, id string) error {
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

func DeleteUserRoleByRoleId(tx *sql.Tx, id string) error {
	query := `DELETE user_roles_mapping where rId = ?`

	_, err := tx.Query(query, id)
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
	query := `SELECT ID, NAME, 
	ISNULL((select count(USER_ID) from USER_GROUP_MEMBERSHIP where GROUP_ID = g.ID AND REALM_ID = ? group by GROUP_ID), 0) as countMembers
	,createDate, createId, modifyDate, modifyId
	from KEYCLOAK_GROUP g
	where
	REALM_ID = ?`

	rows, err := db.Query(query, config.GetConfig().Keycloak_realm, config.GetConfig().Keycloak_realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.GroupItem, 0)

	for rows.Next() {
		var r models.GroupItem

		err := rows.Scan(&r.ID, &r.Name, &r.CountMembers, &r.CreateDate, &r.CreateId, &r.ModifyDate, &r.ModifyId)
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

func GetUsers() ([]models.GetUserInfo, error) {
	query := `SELECT ID, ENABLED, USERNAME, FIRST_NAME, LAST_NAME, EMAIL, createDate, createId, modifyDate, modifyId FROM USER_ENTITY where REALM_ID = ?`

	rows, err := db.Query(query, config.GetConfig().Keycloak_realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.GetUserInfo, 0)

	for rows.Next() {
		var r models.GetUserInfo

		err := rows.Scan(&r.ID, &r.Enabled, &r.Username, &r.FirstName, &r.LastName, &r.Email, &r.CreateDate, &r.CreateId, &r.ModifyDate, &r.ModifyId)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}
	return arr, err
}

func GetUserDetail(userId string) ([]models.GetUserInfo, error) {
	query := `SELECT ID, ENABLED, USERNAME, FIRST_NAME, LAST_NAME, EMAIL, 
	(SELECT STRING_AGG(REQUIRED_ACTION, ',') FROM USER_REQUIRED_ACTION WHERE USER_ID=U.ID) as REQUIRED_ACTION,
	createDate, createId, modifyDate, modifyId FROM USER_ENTITY U
	WHERE U.ID = ? AND REALM_ID = ?`

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
		err := rows.Scan(&r.ID, &r.Enabled, &r.Username, &r.FirstName, &r.LastName, &r.Email, &RequiredActions, &r.CreateDate, &r.CreateId, &r.ModifyDate, &r.ModifyId)
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
	query += `select C.secretGroup, D.createDate, D.createId, D.modifyDate, D.modifyId 
	from (select REPLACE(REPLACE(B.sg,'/secret/',''),'/','') as secretGroup 
	from (select
	REPLACE(url,'*','%%') as auth_url
	from USER_ENTITY u
	join user_roles_mapping ur on u.ID = ur.userId
	join roles_authority_mapping ra on ur.rId = ra.rId
	join authority a on ra.aId = a.aId
	where ur.useYn = 'y'
	and ra.useYn = 'y'
	and u.USERNAME = ?
	and u.REALM_ID = ?
	) A
	join @values B
	ON PATINDEX(A.auth_url, B.sg) = 1
	group by B.sg) C
	left outer join
	vSecretGroup D
	ON C.secretGroup = D.vSecretGroupPath`

	rows, err := db.Query(query, username, config.GetConfig().Keycloak_realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	arr := make([]models.SecretGroupItem, 0)

	for rows.Next() {
		var r models.SecretGroupItem

		err := rows.Scan(&r.Name, &r.CreateDate, &r.CreateId, &r.ModifyDate, &r.ModifyId)
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
	query := `SELECT vSecretPath, createDate, createId, modifyDate, modifyId FROM vSecret
	WHERE vSecretGroupId = (SELECT vSecretGroupId FROM vSecretGroup WHERE vSecretGroupPath = ? AND REALM_ID = ?)`

	rows, err := db.Query(query, groupName, config.GetConfig().Keycloak_realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var m = make(map[string]models.SecretItem)

	for rows.Next() {
		var r models.SecretItem

		err := rows.Scan(&r.Name, &r.CreateDate, &r.CreateId, &r.ModifyDate, &r.ModifyId)
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

	query = `SELECT createDate, createId, modifyDate, modifyId
	FROM vSecretGroup
	where vSecretGroupPath = ?
	AND REALM_ID = ?`

	rows, err = db.Query(query, groupName, config.GetConfig().Keycloak_realm)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&result.CreateDate, &result.CreateId, &result.ModifyDate, &result.ModifyId)
		if err != nil {
			return result, err
		}
	}

	return result, err
}

func GetSecretByName(groupName string, secretName string) (*models.SecretItem, error) {
	query := `SELECT vSecretPath, createDate, createId, modifyDate, modifyId FROM vSecret
	WHERE vSecretGroupId = (SELECT vSecretGroupId FROM vSecretGroup WHERE vSecretGroupPath = ? AND REALM_ID = ?) AND vSecretPath = ?`

	rows, err := db.Query(query, groupName, config.GetConfig().Keycloak_realm, secretName)
	if err != nil {
		return nil, err
	}

	m := new(models.SecretItem)

	rows.Next()
	err = rows.Scan(&m.Name, &m.CreateDate, &m.CreateId, &m.ModifyDate, &m.ModifyId)
	if err != nil {
		return m, nil
	}
	defer rows.Close()

	return m, err
}
