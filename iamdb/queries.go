package iamdb

import (
	"database/sql"
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
