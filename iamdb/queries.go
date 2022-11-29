package iamdb

import (
	"database/sql"
	"errors"
	"fmt"
	"iam/config"
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

func GetRoles() ([]models.RolesInfo, error) {
	db, err := DBClient()
	defer db.Close()
	if err != nil {
		return nil, err
	}

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
	where r.REALM_ID = ?
	order by r.rName`

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
	db, err := DBClient()
	defer db.Close()
	if err != nil {
		return "", err
	}

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
	db, err := DBClient()
	defer db.Close()
	if err != nil {
		return "", err
	}

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

func CreateRolesIdTx(tx *sql.Tx, id string, name string, defaultRole bool, username string) error {
	query := `INSERT INTO roles(rId, rName, defaultRole, REALM_ID, createId, modifyId) 
	select ?, ?, ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := tx.Query(query, id, name, defaultRole, config.GetConfig().Keycloak_realm, username, config.GetConfig().Keycloak_realm)
	return err
}

func DeleteRolesTx(tx *sql.Tx, id string) error {
	query := `DELETE roles where rId = ? AND REALM_ID = ?
	SELECT @@ROWCOUNT`

	rows, err := tx.Query(query, id, config.GetConfig().Keycloak_realm)

	err = resultErrorCheck(rows)
	return err
}

func DeleteRolesNameTx(tx *sql.Tx, name string) error {
	query := `DELETE roles where rName = ? AND REALM_ID = ?
	SELECT @@ROWCOUNT`

	rows, err := tx.Query(query, name, config.GetConfig().Keycloak_realm)
	err = resultErrorCheck(rows)
	return err
}

func UpdateRolesTx(tx *sql.Tx, role *models.RolesInfo, username string) error {
	var err error
	var rows *sql.Rows = nil

	//버그가 있는듯... db.Query에 nil 을 넣었을 때 IsNull 의 동작이 이상하다...
	//어쩔 수 없이 쿼리 2개로 나눠놓음
	if role.Name != nil {
		query := `UPDATE roles SET 
		rName = ?, 
		defaultRole = IsNull(?, defaultRole), 
		modifyDate=GETDATE(), 
		modifyId=(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) 
		where rId = ? AND REALM_ID = ?
		SELECT @@ROWCOUNT`
		rows, err = tx.Query(query, role.Name, role.DefaultRole, username, config.GetConfig().Keycloak_realm, role.ID, config.GetConfig().Keycloak_realm)
	} else {
		query := `UPDATE roles SET 
		defaultRole = ?, 
		modifyDate=GETDATE(), 
		modifyId=(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) 
		where rId = ? AND REALM_ID = ?
		SELECT @@ROWCOUNT`
		rows, err = tx.Query(query, role.DefaultRole, username, config.GetConfig().Keycloak_realm, role.ID, config.GetConfig().Keycloak_realm)
	}

	err = resultErrorCheck(rows)
	return err
}

func GetMyAuth(id string) ([]string, error) {
	db, err := DBClient()
	defer db.Close()
	if err != nil {
		return nil, err
	}

	query := `select a.aName
	from user_roles_mapping ur 
	join roles_authority_mapping ra 
	on ur.rId = ra.rId
	join authority a 
	on ra.aId = a.aId
	where userId = ?
	and	ur.useYn = 1
	and	ra.useYn = 1
	AND a.REALM_ID = ?
	order by a.aName`

	rows, err := db.Query(query, id, config.GetConfig().Keycloak_realm)
	if err != nil {
		return nil, err
	}

	var arr = make([]string, 0)

	for rows.Next() {
		var r string

		err := rows.Scan(&r)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}

	return arr, err
}

func GetMenuAuth(id string, site string) ([]models.MenuAutuhorityInfo, error) {
	db, err := DBClient()
	defer db.Close()
	if err != nil {
		return nil, err
	}

	query := `declare @values table
	(
		aName varchar(310)
		, url varchar(310)
		, method varchar(310)
	)
	
	insert @values(aName, url, method)
	(select a.aName
	, a.url
	, a.method
	from user_roles_mapping ur 
	join roles_authority_mapping ra 
	on ur.rId = ra.rId
	join authority a 
	on ra.aId = a.aId
	where userId = ?
	and	ur.useYn = 1
	and	ra.useYn = 1
	AND a.REALM_ID = ?
	AND (a.method = 'DISABLE')
	AND PATINDEX('SIDE_MENU/' + ? +'/%', a.url) = 1)
	
	insert @values(aName, url, method)
	(select a.aName
	, a.url
	, a.method
	from user_roles_mapping ur 
	join roles_authority_mapping ra 
	on ur.rId = ra.rId
	join authority a 
	on ra.aId = a.aId
	where userId = ?
	and	ur.useYn = 1
	and	ra.useYn = 1
	AND a.REALM_ID = ?
	AND (a.method = 'SHOW')
	AND PATINDEX('SIDE_MENU/' + ? +'/%', a.url) = 1
	AND a.url NOT IN(SELECT url FROM @values))
	
	SELECT aName, url, method FROM @values
	ORDER BY aName`

	rows, err := db.Query(query, id, config.GetConfig().Keycloak_realm, site, id, config.GetConfig().Keycloak_realm, site)
	if err != nil {
		return nil, err
	}

	var arr = make([]models.MenuAutuhorityInfo, 0)

	for rows.Next() {
		var r models.MenuAutuhorityInfo

		err := rows.Scan(&r.Name, &r.URL, &r.Method)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}

	return arr, err
}

func GetRolseAuth(id string) ([]models.RolesInfo, error) {
	db, err := DBClient()
	defer db.Close()
	if err != nil {
		return nil, err
	}

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
	where ra.rId = ?
	AND	a.REALM_ID = ?
	order by a.aName`

	rows, err := db.Query(query, id, config.GetConfig().Keycloak_realm)
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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `DELETE FROM roles_authority_mapping where rId = ? AND aId = ?
	SELECT @@ROWCOUNT`

	rows, err := db.Query(query, roleID, authID)
	err = resultErrorCheck(rows)
	return err
}

func UpdateRoleAuth(roleID string, authID string, use bool, username string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `UPDATE roles_authority_mapping SET 
	useYn = ?, 
	modifyDate=GETDATE(), 
	modifyId=(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?)
	where rId = ? 
	AND aId = ?
	SELECT @@ROWCOUNT`

	rows, err := db.Query(query, use, username, config.GetConfig().Keycloak_realm, roleID, authID)
	err = resultErrorCheck(rows)
	return err
}

func GetUserRole(userID string) ([]models.RolesInfo, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

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
	where ur.userId = ?
	AND r.REALM_ID = ?
	order by r.rName`

	rows, err := db.Query(query, userID, config.GetConfig().Keycloak_realm)
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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `DELETE FROM user_roles_mapping where userId = ? AND rId = ?
	SELECT @@ROWCOUNT`

	rows, err := db.Query(query, userID, roleID)
	err = resultErrorCheck(rows)
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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `UPDATE user_roles_mapping SET 
	useYn = ?, 
	modifyDate=GETDATE(), 
	modifyId=(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) 
	where userId = ? AND rId = ?
	SELECT @@ROWCOUNT`

	rows, err := db.Query(query, use, username, config.GetConfig().Keycloak_realm, userID, roleID)
	err = resultErrorCheck(rows)
	return err
}

func GetUserAuth(userID string) ([]models.AutuhorityInfo, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

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
	AND a.REALM_ID = ?
	order by a.aName`

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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

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
	where a.REALM_ID = ?
	order by a.aName`

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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

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
	query := `DELETE authority where aId = ? AND REALM_ID = ?
	SELECT @@ROWCOUNT`

	rows, err := tx.Query(query, id, config.GetConfig().Keycloak_realm)
	err = resultErrorCheck(rows)
	return err
}

func DeleteAuthNameTx(tx *sql.Tx, name string) error {
	query := `DELETE authority where aName = ? AND REALM_ID = ?
	SELECT @@ROWCOUNT`

	rows, err := tx.Query(query, name, config.GetConfig().Keycloak_realm)
	err = resultErrorCheck(rows)
	return err
}

func UpdateAuth(auth *models.AutuhorityInfo, username string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `UPDATE authority SET 
	aName = ?, 
	url = ?, 
	method = ?, 
	modifyDate=GETDATE(), 
	modifyId=(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) 
	where aId = ?
	SELECT @@ROWCOUNT`

	rows, err := db.Query(query, auth.Name, auth.URL, auth.Method, username, config.GetConfig().Keycloak_realm, auth.ID)
	err = resultErrorCheck(rows)
	return err
}

func GetAuthInfo(authID string) (*models.AutuhorityInfo, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `DELETE user_roles_mapping where userId = ?`

	_, err := db.Query(query, id)
	return err
}

func CheckRoleAuthID(roleID string, authID string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

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
	g.REALM_ID = ?
	order by NAME`

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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `UPDATE KEYCLOAK_GROUP SET 
	createId=B.ID,
	createDate=GETDATE(),
	modifyId=B.ID,
	modifyDate=GETDATE()
	FROM KEYCLOAK_GROUP A,
	(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) B
	where A.ID = ?
	SELECT @@ROWCOUNT`

	rows, err := db.Query(query, username, config.GetConfig().Keycloak_realm, groupId)
	err = resultErrorCheck(rows)
	return err
}

func GroupUpdate(groupId string, username string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `UPDATE KEYCLOAK_GROUP SET 
	modifyId=B.ID,
	modifyDate=GETDATE()
	FROM KEYCLOAK_GROUP A,
	(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) B
	where A.ID = ?
	SELECT @@ROWCOUNT`

	rows, err := db.Query(query, username, config.GetConfig().Keycloak_realm, groupId)
	err = resultErrorCheck(rows)
	return err
}

func GetUsers(search string, groupid string, userids []string) ([]models.GetUserInfo, error) {
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

	if len(userids) > 0 {
		placeholder := strings.Repeat("?,", len(userids))
		placeholder = placeholder[:len(placeholder)-1]
		query += " AND U.ID IN (" + placeholder + ")"
	}

	query += " ORDER BY U.USERNAME"

	queryParams := []interface{}{config.GetConfig().Keycloak_realm}
	if search != "" {
		queryParams = append(queryParams, "%"+search+"%")
	}

	if groupid != "" {
		queryParams = append(queryParams, groupid)
	}

	if len(userids) > 0 {
		for item := range userids {
			queryParams = append(queryParams, userids[item])
		}
	}

	rows, err = db.Query(query, queryParams...)

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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `UPDATE USER_ENTITY SET 
	createId=B.ID,
	createDate=GETDATE(),
	modifyId=B.ID,
	modifyDate=GETDATE()
	FROM USER_ENTITY A,
	(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) B
	where A.ID = ?
	SELECT @@ROWCOUNT`

	rows, err := db.Query(query, username, config.GetConfig().Keycloak_realm, userId)
	err = resultErrorCheck(rows)
	return err
}

func UsersUpdate(userId string, username string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `UPDATE USER_ENTITY SET 
	modifyId=B.ID,
	modifyDate=GETDATE()
	FROM USER_ENTITY A,
	(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) B
	where A.ID = ?
	SELECT @@ROWCOUNT`

	rows, err := db.Query(query, username, config.GetConfig().Keycloak_realm, userId)
	err = resultErrorCheck(rows)
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

func MergeSecret(secretPath string, secretGroupPath string, url *string, username string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `MERGE INTO vSecret A
	USING (SELECT 
	? as spath, 
	(select vSecretGroupId from vSecretGroup where vSecretGroupPath = ? AND REALM_ID = ?) as sgid,
	(select ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) as userid,
	? as url
	) B
	ON A.vSecretPath = B.spath
	AND A.vSecretGroupId = B.sgid
	WHEN MATCHED THEN
		UPDATE SET 
		modifyDate = GETDATE(),
		modifyId = B.userid,
		url = B.url
	WHEN NOT MATCHED THEN
		INSERT (vSecretPath, vSecretGroupId, url, createId, modifyId)
		VALUES(B.spath, B.sgid, B.url, B.userid, B.userid);`

	_, err := db.Query(query, secretPath, secretGroupPath, config.GetConfig().Keycloak_realm, username, config.GetConfig().Keycloak_realm, url)

	return err
}

func DeleteSecret(secretPath string, secretGroupPath string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `DELETE FROM vSecret WHERE vSecretPath = ?
	AND vSecretGroupId = (select vSecretGroupId from vSecretGroup where vSecretGroupPath = ?)
	SELECT @@ROWCOUNT`

	rows, err := db.Query(query, secretPath, secretGroupPath)
	err = resultErrorCheck(rows)
	return err
}

func GetSecretGroup(data []models.SecretGroupItem, username string) ([]models.SecretGroupItem, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	query := `declare @values table
	(
		sg varchar(310)
	)`
	for _, d := range data {
		query += "insert into @values values ('/iam/secret/" + d.Name + "/')"
	}
	query += `select C.secretGroup, 
	FORMAT(D.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(D.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
	from (select REPLACE(REPLACE(B.sg,'/iam/secret/',''),'/','') as secretGroup 
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
	ON C.secretGroup = D.vSecretGroupPath
	ORDER BY C.secretGroup`

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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

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
	WHERE s.vSecretGroupId = (SELECT vSecretGroupId FROM vSecretGroup WHERE vSecretGroupPath = ? AND REALM_ID = ?)
	ORDER BY s.vSecretPath`

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

	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return result, dbErr
	}

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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	query := `SELECT s.vSecretPath, s.url, 
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
	err = rows.Scan(&m.Name, &m.Url, &m.CreateDate, &m.Creator, &m.ModifyDate, &m.Modifier)
	if err != nil {
		return m, nil
	}
	defer rows.Close()

	return m, err
}

func MetricCount() (map[string]int, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	query := `select 
	(select count(*) from USER_ENTITY where REALM_ID = ? AND SERVICE_ACCOUNT_CLIENT_LINK is NULL) AS users,
	(select count(*) from KEYCLOAK_GROUP where REALM_ID = ?) AS groups,
	(select count(*) from CLIENT where REALM_ID = ? AND (NAME IS NULL OR LEN(NAME) = 0)) AS applicastions,
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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	query := `select CLIENT_ID from CLIENT where REALM_ID = ? AND (NAME IS NULL OR LEN(NAME) = 0)`

	rows, err := db.Query(query, config.GetConfig().Keycloak_realm)
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
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

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
	AND (NAME IS NULL OR LEN(NAME) = 0)
	) B
	ON A.CLIENT_ID = B.CLIENT_ID
	group by B.client_id`

	rows, err := db.Query(query,
		config.GetConfig().Keycloak_client_id,
		config.GetConfig().Keycloak_realm,
		date,
		config.GetConfig().Keycloak_realm)

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

func GetLoginApplicationDate(date int) ([]map[string]interface{}, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	query := `select 
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
	where
	REALM_ID = ?
	AND  (NAME IS NULL OR LEN(NAME) = 0)
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
	ORDER BY E.SYSTEM_DATE, E.CLIENT_ID`

	rows, err := db.Query(query,
		config.GetConfig().Keycloak_realm,
		date,
		date)

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

func GetLoginError(date int) ([]models.MetricItem, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

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

	_, err := db.Query(query, username, config.GetConfig().Keycloak_realm, uid)
	if err != nil {
		return err
	}

	return nil
}

func GetIdpCount() ([]models.MetricItem, error) {
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

	rows, err := db.Query(query, config.GetConfig().Keycloak_realm)
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

func GetAllSecret(data []models.SecretGroupItem) ([]models.SecretItem, error) {
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
	args = append(args, config.GetConfig().Keycloak_realm)

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
