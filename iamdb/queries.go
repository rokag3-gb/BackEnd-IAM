package iamdb

import (
	"database/sql"
	"fmt"
)

func GetUserAuthoritiesForEndpoint(username string, realm string, method string, endpoint string) (*sql.Rows, error) {
	query := fmt.Sprintf(`select
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
	u.USERNAME = '%s' AND
	u.REALM_ID = '%s' AND
	(a.endpointMethod = '%s' OR a.endpointMethod = 'ALL') AND
	PATINDEX(REPLACE(a.endpointUrl,'*','%%'), '%s') = 1`,
		username, realm, method, endpoint)

	rows, err := db.Query(query)
	return rows, err
}
