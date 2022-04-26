package iamdb

import (
	"database/sql"
	"errors"
	"iam/models"
)

func ConnectionTest() (*sql.Rows, error) {
	query := "select 1"

	rows, err := db.Query(query)
	return rows, err
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
	PATINDEX(REPLACE(a.url,'*','%%'), ?) = 1`

	rows, err := db.Query(query, username, realm, method, endpoint)
	return rows, err
}

func GetRoles() (*sql.Rows, error) {
	query := "select rId, rName from roles"

	rows, err := db.Query(query)
	return rows, err
}

func CreateRoles(name string) (*sql.Rows, error) {
	query := `INSERT INTO roles(rName) VALUES(?)`

	rows, err := db.Query(query, name)
	return rows, err
}

func DeleteRoles(id string) (*sql.Rows, error) {
	query := `DELETE roles where rId=?`

	rows, err := db.Query(query, id)
	return rows, err
}

func UpdateRoles(name string, id string) (*sql.Rows, error) {
	query := `UPDATE roles SET rName=? where rId=?`

	rows, err := db.Query(query, name, id)
	return rows, err
}

func GetRolseAuth(id string) (*sql.Rows, error) {
	query := `select
	a.aId, a.aName 
	from 
	roles_authority_mapping ra 
	join 
	authority a 
	on 
	ra.aId = a.aId 
	where 
	ra.rId = ?`

	rows, err := db.Query(query, id)
	return rows, err
}

func AssignRoleAuth(roleID string, authID string) (*sql.Rows, error) {
	query := `INSERT INTO roles_authority_mapping(rId, aId) VALUES(?, ?)`

	rows, err := db.Query(query, roleID, authID)
	return rows, err
}

func DismissRoleAuth(roleID string, authID string) (*sql.Rows, error) {
	query := `DELETE FROM roles_authority_mapping where rId = ? AND aId = ?`

	rows, err := db.Query(query, roleID, authID)
	return rows, err
}

func GetUserRole(userID string) (*sql.Rows, error) {
	query := `select r.rId, r.rName
	from 
	roles r join
	user_roles_mapping ur 
	on r.rId = ur.rId
	where
	ur.userId = ?
	AND
	ur.useYn = 'y'`

	rows, err := db.Query(query, userID)
	return rows, err
}

func AssignUserRole(userID string, roleID string) (*sql.Rows, error) {
	query := `INSERT INTO user_roles_mapping(userId, rId) VALUES(?, ?)`

	rows, err := db.Query(query, userID, roleID)
	return rows, err
}

func DismissUserRole(userID string, roleID string) (*sql.Rows, error) {
	query := `DELETE FROM user_roles_mapping where userId = ? AND rId = ?`

	rows, err := db.Query(query, userID, roleID)
	return rows, err
}

func GetUserAuth(userID string) (*sql.Rows, error) {
	query := `select a.aId, a.aName 
	from 
	user_roles_mapping ur 
	join 
	roles_authority_mapping ra 
	on ur.rId = ra.rId
	join 
	authority a 
	on ra.aId = a.aId
	where 
	userId = ?
	and
	ur.useYn = 'y'
	and
	ra.useYn = 'y'`

	rows, err := db.Query(query, userID)
	return rows, err
}

func GetUserAuthActive(userName string, authName string) (*sql.Rows, error) {
	query := `select 1
	from 
	USER_ENTITY u
	join
	user_roles_mapping ur 
	on u.ID = ur.userId
	join 
	roles_authority_mapping ra 
	on ur.rId = ra.rId
	join 
	authority a 
	on ra.aId = a.aId
	where u.USERNAME = ?
	AND
	a.aName = ?
	and
	ur.useYn = 'y'
	and
	ra.useYn = 'y'`

	rows, err := db.Query(query, userName, authName)
	return rows, err
}

func GetAuth() (*sql.Rows, error) {
	query := `select aId, aName from authority`

	rows, err := db.Query(query)
	return rows, err
}

func CreateAuth(auth *models.AutuhorityInfo) (*sql.Rows, error) {
	query := `INSERT INTO authority(aName, url, method) VALUES(?, ?, ?)`

	rows, err := db.Query(query, auth.Name, auth.URL, auth.Method)
	return rows, err
}

func DeleteAuth(id string) (*sql.Rows, error) {
	query := `DELETE authority where aId=?`

	rows, err := db.Query(query, id)
	return rows, err
}

func UpdateAuth(auth *models.AutuhorityInfo) (*sql.Rows, error) {
	query := `UPDATE authority SET aName=?, url=?, method=? where aId=?`

	rows, err := db.Query(query, auth.Name, auth.URL, auth.Method, auth.ID)
	return rows, err
}

func GetAuthInfo(authID string) (*sql.Rows, error) {
	query := `select aId, aName, url, method from authority where aId = ?`

	rows, err := db.Query(query, authID)
	return rows, err
}

func DeleteRolesAuthByRoleId(id string) (*sql.Rows, error) {
	query := `DELETE roles_authority_mapping where rId=?`

	rows, err := db.Query(query, id)
	return rows, err
}

func DeleteRolesAuthByAuthId(id string) (*sql.Rows, error) {
	query := `DELETE roles_authority_mapping where aId=?`

	rows, err := db.Query(query, id)
	return rows, err
}

func DeleteUserRoleByUserId(id string) (*sql.Rows, error) {
	query := `DELETE user_roles_mapping where userId=?`

	rows, err := db.Query(query, id)
	return rows, err
}

func DeleteUserRoleByRoleId(id string) (*sql.Rows, error) {
	query := `DELETE user_roles_mapping where rId=?`

	rows, err := db.Query(query, id)
	return rows, err
}

func CheckRoleAuthID(roleID string, authID string) error {
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
		return errors.New("value is wrong")
	}

	return nil
}

func CheckUserRoleID(userID string, roleID string) error {
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
		return errors.New("value is wrong")
	}

	return nil
}

func GetGroupMembersCountMap() (map[string]int, error) {
	query := `select GROUP_ID, count(USER_ID) as countMembers from USER_GROUP_MEMBERSHIP group by GROUP_ID`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	countMap := make(map[string]int)
	for rows.Next() {
		var groupID string
		var countMembers int
		err = rows.Scan(&groupID, &countMembers)
		if err != nil {
			return nil, err
		}
		countMap[groupID] = countMembers
	}

	return countMap, nil
}
