package iamdb

import (
	"database/sql"
	"errors"
	"iam/models"
)

func GetRoles(realm string) ([]models.RolesInfo, error) {
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

	rows, err := db.Query(query, realm)

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

func GetRoleIdByName(rolename, realm string) (string, error) {
	db, err := DBClient()
	defer db.Close()
	if err != nil {
		return "", err
	}

	query := `select rId from roles
	where rName = ?
	AND REALM_ID = ?`

	rows, err := db.Query(query, rolename, realm)

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

func GetAuthIdByName(authname, realm string) (string, error) {
	db, err := DBClient()
	defer db.Close()
	if err != nil {
		return "", err
	}

	query := `select aId from authority
	where aName = ?
	AND REALM_ID = ?`

	rows, err := db.Query(query, authname, realm)

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

func CreateRolesIdTx(tx *sql.Tx, id string, name string, defaultRole bool, username string, realm string) error {
	query := `INSERT INTO roles(rId, rName, defaultRole, REALM_ID, createId, modifyId) 
	select ?, ?, ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := tx.Query(query, id, name, defaultRole, realm, username, realm)
	return err
}

func DeleteRolesTx(tx *sql.Tx, id, realm string) error {
	query := `DELETE roles where rId = ? AND REALM_ID = ?
	SELECT @@ROWCOUNT`

	rows, err := tx.Query(query, id, realm)

	err = resultErrorCheck(rows)
	return err
}

func DeleteRolesNameTx(tx *sql.Tx, name, realm string) error {
	query := `DELETE roles where rName = ? AND REALM_ID = ?`

	_, err := tx.Query(query, name, realm)
	return err
}

func UpdateRolesTx(tx *sql.Tx, role *models.RolesInfo, username, realm string) error {
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
		rows, err = tx.Query(query, role.Name, role.DefaultRole, username, realm, role.ID, realm)
	} else {
		query := `UPDATE roles SET 
		defaultRole = ?, 
		modifyDate=GETDATE(), 
		modifyId=(SELECT ID FROM USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?) 
		where rId = ? AND REALM_ID = ?
		SELECT @@ROWCOUNT`
		rows, err = tx.Query(query, role.DefaultRole, username, realm, role.ID, realm)
	}

	err = resultErrorCheck(rows)
	return err
}

func GetMyAuth(id, realm string) ([]string, error) {
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

	rows, err := db.Query(query, id, realm)
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

func GetMenuAuth(id, site, realm string) ([]models.MenuAutuhorityInfo, error) {
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

	rows, err := db.Query(query, id, realm, site, id, realm, site)
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

func GetRolseAuth(id, realm string) ([]models.RolesInfo, error) {
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

	rows, err := db.Query(query, id, realm)
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

func AssignRoleAuth(roleID, authID, username, realm string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `INSERT INTO roles_authority_mapping(rId, aId, createId, modifyId)
	select ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := db.Query(query, roleID, authID, username, realm)
	return err
}

func AssignRoleAuthTx(tx *sql.Tx, roleID, authID, username, realm string) error {
	query := `INSERT INTO roles_authority_mapping(rId, aId, createId, modifyId)
	select ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := tx.Query(query, roleID, authID, username, realm)
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

func UpdateRoleAuth(roleID string, authID string, use bool, username, realm string) error {
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

	rows, err := db.Query(query, use, username, realm, roleID, authID)
	err = resultErrorCheck(rows)
	return err
}

func GetUserRole(userID, realm string) ([]models.RolesInfo, error) {
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

	rows, err := db.Query(query, userID, realm)
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

func AssignUserRole(userID, roleID, username, realm string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `INSERT INTO user_roles_mapping(userId, rId, createId, modifyId)
	select ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := db.Query(query, userID, roleID, username, realm)
	return err
}

func AssignUserRoleTx(tx *sql.Tx, userID, roleID, username, realm string) error {
	query := `INSERT INTO user_roles_mapping(userId, rId, createId, modifyId)
	select ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := tx.Query(query, userID, roleID, username, realm)
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

func DeleteUserRoleByRoleNameTx(tx *sql.Tx, roleName, realm string) error {
	query := `DELETE FROM user_roles_mapping where 
	rId = (select rId from roles where rName = ? AND REALM_ID = ?)`

	_, err := tx.Query(query, roleName, realm)
	return err
}

func DeleteUserRoleByRoleIdTx(tx *sql.Tx, roleName string) error {
	query := `DELETE FROM user_roles_mapping where rId = ?`

	_, err := tx.Query(query, roleName)
	return err
}

func UpdateUserRole(userID, roleID, username, realm string, use bool) error {
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

	rows, err := db.Query(query, use, username, realm, userID, roleID)
	err = resultErrorCheck(rows)
	return err
}

func GetUserAuth(userID, realm string) ([]models.AutuhorityInfo, error) {
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

	rows, err := db.Query(query, userID, realm)

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

func GetUserAuthActive(userName, authName, realm string) (map[string]interface{}, error) {
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

	rows, err := db.Query(query, userName, authName, realm)

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

func GetAuth(realm string) ([]models.AutuhorityInfo, error) {
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

	rows, err := db.Query(query, realm)
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

func CreateAuth(auth *models.AutuhorityInfo, username, realm string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `INSERT INTO authority(aId, aName, url, method, REALM_ID, createId, modifyId)
	select ?, ?, ?, ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := db.Query(query, auth.ID, auth.Name, auth.URL, auth.Method, realm, username, realm)
	return err
}

func CreateAuthIdTx(tx *sql.Tx, id, name, url, method, username, realm string) error {
	query := `INSERT INTO authority(aId, aName, url, method, REALM_ID, createId, modifyId)
	select ?, ?, ?, ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := tx.Query(query, id, name, url, method, realm, username, realm)
	return err
}

func DeleteAuth(tx *sql.Tx, id, realm string) error {
	query := `DELETE authority where aId = ? AND REALM_ID = ?
	SELECT @@ROWCOUNT`

	rows, err := tx.Query(query, id, realm)
	err = resultErrorCheck(rows)
	return err
}

func DeleteAuthNameTx(tx *sql.Tx, name, realm string) error {
	query := `DELETE authority where aName = ? AND REALM_ID = ?`

	_, err := tx.Query(query, name, realm)
	return err
}

func UpdateAuth(auth *models.AutuhorityInfo, username, realm string) error {
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

	rows, err := db.Query(query, auth.Name, auth.URL, auth.Method, username, realm, auth.ID)
	err = resultErrorCheck(rows)
	return err
}

func GetAuthInfo(authID, realm string) (*models.AutuhorityInfo, error) {
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

	rows, err := db.Query(query, authID, realm)
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

func DeleteRolesAuthByAuthNameTx(tx *sql.Tx, roleName, realm string) error {
	query := `DELETE roles_authority_mapping where aId =
	(select aId from authority where aName = ? AND REALM_ID = ?)`

	_, err := tx.Query(query, roleName, realm)
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

func GetAccountUserId(id, realm string) ([]string, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	query := `SELECT Seq 
	FROM Sale.dbo.Account_User AU
	JOIN IAM.dbo.USER_ENTITY U
	ON AU.UserId = U.ID AND U.REALM_ID = ?
	WHERE UserId = ?`

	rows, err := db.Query(query, realm, id)
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

func CheckRoleAuthID(roleID, authID, realm string) error {
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

	rows, err := db.Query(query, roleID, realm, authID, realm)
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

func CheckUserRoleID(userID, roleID, realm string) error {
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

	rows, err := db.Query(query, userID, realm, roleID, realm)
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
