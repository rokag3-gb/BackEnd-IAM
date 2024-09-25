package iamdb

import (
	"database/sql"
	"errors"
	"iam/models"
)

func GetRoles() ([]models.RolesInfo, error) {
	db, err := DBClient()
	defer db.Close()
	if err != nil {
		return nil, err
	}

	query := `select r.rId, 
	r.rName, 
	r.defaultRole, 
	r.REALM_ID,
	FORMAT(r.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(r.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
from roles r
	LEFT OUTER JOIN USER_ENTITY u1 on r.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2 on r.modifyId = u2.ID
order by r.rName`

	rows, err := db.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.RolesInfo, 0)

	for rows.Next() {
		var r models.RolesInfo

		err := rows.Scan(&r.ID, &r.Name, &r.DefaultRole, &r.Realm, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
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

func CreateRolesIdTx(tx *sql.Tx, id, name, userId, realm string, defaultRole bool) error {
	query := `INSERT INTO roles(rId, rName, defaultRole, REALM_ID, createId, modifyId) 
	VALUES(?, ?, ?, ?, ?, ?)`

	_, err := tx.Query(query, id, name, defaultRole, realm, userId, userId)
	return err
}

func DeleteRolesTx(tx *sql.Tx, id string) error {
	query := `DELETE roles where rId = ? SELECT @@ROWCOUNT`

	rows, err := tx.Query(query, id)

	err = resultErrorCheck(rows)
	return err
}

func DeleteRolesNameTx(tx *sql.Tx, name, realm string) error {
	query := `DELETE roles where rName = ? AND REALM_ID = ?`

	_, err := tx.Query(query, name, realm)
	return err
}

func UpdateRolesTx(tx *sql.Tx, role *models.RolesInfo, userid string) error {
	var err error
	var rows *sql.Rows = nil

	//버그가 있는듯... db.Query에 nil 을 넣었을 때 IsNull 의 동작이 이상하다...
	//어쩔 수 없이 쿼리 2개로 나눠놓음
	if role.Name != nil {
		query := `UPDATE roles SET 
		rName = ?, 
		defaultRole = IsNull(?, defaultRole), 
		modifyDate=GETDATE(), 
		modifyId=?
		where rId = ?
		SELECT @@ROWCOUNT`
		rows, err = tx.Query(query, role.Name, role.DefaultRole, userid, role.ID)
	} else {
		query := `UPDATE roles SET 
		defaultRole = ?, 
		modifyDate=GETDATE(), 
		modifyId=?
		where rId = ?
		SELECT @@ROWCOUNT`
		rows, err = tx.Query(query, role.DefaultRole, userid, role.ID)
	}

	err = resultErrorCheck(rows)
	return err
}

// 토큰에서 테넌트ID를 받아와야 하는 유형
func GetMyAuth(id, tenantId string) ([]string, error) {
	db, err := DBClient()
	defer db.Close()
	if err != nil {
		return nil, err
	}

	query := `select a.aName
	from UserRole ur 
		join roles_authority_mapping ra on ur.RoleId = ra.rId
		join authority a on ra.aId = a.aId
		JOIN Tenant t ON ur.TenantId = t.TenantId AND t.TenantId = ?
	where userId = ?
		and	ra.useYn = 1
	order by a.aName`

	rows, err := db.Query(query, tenantId, id)
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

// 토큰에서 테넌트ID를 받아와야 하는 유형
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
from UserRole ur 
	join roles_authority_mapping ra on ur.RoleId = ra.rId
	join authority a on ra.aId = a.aId
where userId = ?
	and	ra.useYn = 1
	AND a.REALM_ID = ?
	AND (a.method = 'DISABLE')
	AND PATINDEX('SIDE_MENU/' + ? +'/%', a.url) = 1)
	
insert @values(aName, url, method)
	(select a.aName
	, a.url
	, a.method
from UserRole ur 
	join roles_authority_mapping ra on ur.RoleId = ra.rId
	join authority a on ra.aId = a.aId
where userId = ?
	and	ra.useYn = 1
	AND a.REALM_ID = ?
	AND (a.method = 'SHOW')
	AND PATINDEX('SIDE_MENU/' + ? +'/%', a.url) = 1
	AND a.url NOT IN(SELECT url FROM @values))
	
SELECT aName, url, method FROM @values ORDER BY aName`

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

func GetRolseAuth(id string) ([]models.RolesInfo, error) {
	db, err := DBClient()
	defer db.Close()
	if err != nil {
		return nil, err
	}

	query := `select a.aId, 
	a.aName, 
	ra.useYn, 
	a.REALM_ID, 
	FORMAT(ra.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(ra.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
from roles_authority_mapping ra 
	join authority a on ra.aId = a.aId 
	LEFT OUTER JOIN USER_ENTITY u1 on ra.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2 on ra.modifyId = u2.ID
where ra.rId = ?
order by a.aName`

	rows, err := db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.RolesInfo, 0)

	for rows.Next() {
		var r models.RolesInfo

		err := rows.Scan(&r.ID, &r.Name, &r.Use, &r.Realm, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}

	return arr, err
}

func AssignRoleAuth(roleId, authId, userid string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `INSERT INTO roles_authority_mapping(rId, aId, createId, modifyId)
	VALUES(?, ?, ?, ?)`

	_, err := db.Query(query, roleId, authId, userid, userid)
	return err
}

func AssignRoleAuthTx(tx *sql.Tx, roleID, authID, userid string) error {
	query := `INSERT INTO roles_authority_mapping(rId, aId, createId, modifyId)
	VALUES(?, ?, ?, ?)`

	_, err := tx.Query(query, roleID, authID, userid, userid)
	return err
}

func DismissRoleAuth(roleID string, authID string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `DELETE FROM roles_authority_mapping where rId = ? AND aId = ? SELECT @@ROWCOUNT`

	rows, err := db.Query(query, roleID, authID)
	err = resultErrorCheck(rows)
	return err
}

func UpdateRoleAuth(userId, roleId, authId string, use bool) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `UPDATE roles_authority_mapping SET 
	useYn = ?, 
	modifyDate=GETDATE(), 
	modifyId=?
	where rId = ? 
	AND aId = ?
	SELECT @@ROWCOUNT`

	rows, err := db.Query(query, use, userId, roleId, authId)
	err = resultErrorCheck(rows)
	return err
}

func GetUserRole(userId string) ([]models.RolesInfo, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	query := `select r.rId, 
	r.rName, 
	r.defaultRole, 
	t.TenantId,
	FORMAT(ur.SavedAt, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as SaveUser
from roles r 
	join UserRole ur on r.rId = ur.RoleId
	join Tenant t on ur.TenantId = t.TenantId
	LEFT OUTER JOIN USER_ENTITY u1 on ur.SaverId = u1.ID
where ur.userId = ?
order by r.rName`

	rows, err := db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.RolesInfo, 0)

	for rows.Next() {
		var r models.RolesInfo

		err := rows.Scan(&r.ID, &r.Name, &r.DefaultRole, &r.TenantId, &r.CreateDate, &r.Creator)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}
	return arr, err
}

func AssignUserRole(userID, tenantId, roleId, reqUserId string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `INSERT INTO UserRole(tenantId, UserId, RoleId, SavedAt, SaverId)
	VALUES(?, ?, ?, GETDATE(), ?)`

	_, err := db.Query(query, tenantId, userID, roleId, reqUserId)
	return err
}

func AssignUserRoleTx(tx *sql.Tx, userID, tenantId, roleId, reqUserId string) error {
	query := `INSERT INTO UserRole(tenantId, UserId, RoleId, SavedAt, SaverId)
	VALUES(?, ?, ?, GETDATE(), ?)`

	_, err := tx.Query(query, tenantId, userID, roleId, reqUserId)
	return err
}

func DismissUserRole(userID, roleID, tenantId string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `DELETE FROM UserRole where userId = ? AND RoleId = ? AND TenantId = ?
	SELECT @@ROWCOUNT`

	rows, err := db.Query(query, userID, roleID, tenantId)
	err = resultErrorCheck(rows)
	return err
}

func DeleteUserRoleByRoleNameTx(tx *sql.Tx, roleName, realm string) error {
	query := `DELETE FROM UserRole where 
	RoleId = (select rId from roles where rName = ? AND REALM_ID = ?)`

	_, err := tx.Query(query, roleName, realm)
	return err
}

func DeleteUserRoleByRoleIdTx(tx *sql.Tx, roleName string) error {
	query := `DELETE FROM UserRole where RoleId = ?`

	_, err := tx.Query(query, roleName)
	return err
}

func UpdateUserRole(userId, roleId, tenantId, reqUserId string, use bool) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `UPDATE UserRole SET 
	useYn = ?, 
	modifyDate=GETDATE(), 
	modifyId=? 
	where userId = ? AND RoleId = ? AND TenantId = ?
	SELECT @@ROWCOUNT`

	rows, err := db.Query(query, use, reqUserId, userId, roleId, tenantId)
	err = resultErrorCheck(rows)
	return err
}

func GetUserAuth(userID, tenantId string) ([]models.AutuhorityInfo, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	query := `SELECT a.aId
	, a.aName
	, a.url
	, a.method
	, a.REALM_ID
	, FORMAT(a.createDate, 'yyyy-MM-dd HH:mm') as createDate
	, u1.USERNAME as Creator
	, FORMAT(a.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate 
	, u2.USERNAME as Modifier
FROM UserRole ur 
	join roles_authority_mapping ra on ur.RoleId = ra.rId
	join authority a on ra.aId = a.aId
	JOIN Tenant t ON ur.TenantId = t.TenantId
	LEFT OUTER JOIN USER_ENTITY u1 on a.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2 on a.modifyId = u2.ID
WHERE ur.userId = ?
	and	ra.useYn = 1
	AND t.TenantId = ?
order by a.aName`

	rows, err := db.Query(query, userID, tenantId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.AutuhorityInfo, 0)

	for rows.Next() {
		var r models.AutuhorityInfo

		err := rows.Scan(&r.ID, &r.Name, &r.URL, &r.Method, &r.Realm, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}

	return arr, nil
}

func GetUserAuthActive(userName, authName, tenantId string) (map[string]interface{}, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	query := `select 1
from USER_ENTITY u
	join UserRole ur on u.ID = ur.userId
	join roles_authority_mapping ra on ur.RoleId = ra.rId
	join authority a on ra.aId = a.aId
where u.USERNAME = ?
	AND a.aName = ?
	AND ur.TenantId = ?`

	rows, err := db.Query(query, userName, authName, tenantId)

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

	query := `select 
	a.aId, 
	a.aName, 
	a.url, 
	a.method, 
	a.REALM_ID, 
	FORMAT(a.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(a.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
from authority a
	LEFT OUTER JOIN USER_ENTITY u1 on a.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2 on a.modifyId = u2.ID
order by a.aName`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.AutuhorityInfo, 0)

	for rows.Next() {
		var r models.AutuhorityInfo

		err := rows.Scan(&r.ID, &r.Name, &r.URL, &r.Method, &r.Realm, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}
	return arr, err
}

func CreateAuth(auth *models.AutuhorityInfo, userId, realm string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `INSERT INTO authority(aId, aName, url, method, REALM_ID, createId, modifyId)
	VALUES(?, ?, ?, ?, ?, ?, ?)`

	_, err := db.Query(query, auth.ID, auth.Name, auth.URL, auth.Method, realm, userId, userId)
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

func UpdateAuth(auth *models.AutuhorityInfo, userId string) error {
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
	modifyId=?
	where aId = ?
	SELECT @@ROWCOUNT`

	rows, err := db.Query(query, auth.Name, auth.URL, auth.Method, userId, auth.ID)
	err = resultErrorCheck(rows)
	return err
}

func GetAuthInfo(authID string) (*models.AutuhorityInfo, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	query := `select 
	a.aId, 
	a.aName, 
	a.url, 
	a.method, 
	a.REALM_ID, 
	FORMAT(a.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(a.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
from authority a
	LEFT OUTER JOIN USER_ENTITY u1 on a.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2 on a.modifyId = u2.ID
where a.aId = ?`

	rows, err := db.Query(query, authID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var r *models.AutuhorityInfo = new(models.AutuhorityInfo)

	if rows.Next() {
		err := rows.Scan(&r.ID, &r.Name, &r.URL, &r.Method, &r.Realm, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
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

	query := `DELETE UserRole where userId = ?`

	_, err := db.Query(query, id)
	return err
}

func CheckRoleAuthID(roleID, authID string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `select count(*) as result
	from
	(
	select rid as id from roles where rId = ?
	union 
	select aid as id from authority where aId = ?
	) a`

	rows, err := db.Query(query, roleID, authID)
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

func CheckUserRoleID(userID, roleID string) error {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return dbErr
	}

	query := `select count(*) as result
from
(
select ID as id from USER_ENTITY where ID = ?
union 
select rId as id from roles where rId = ?
) a`

	rows, err := db.Query(query, userID, roleID)
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
