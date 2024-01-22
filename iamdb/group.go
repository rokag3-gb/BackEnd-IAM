package iamdb

import "iam/models"

func GetGroup(realm string) ([]models.GroupItem, error) {
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

	rows, err := db.Query(query, realm, realm)
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

func GroupCreate(groupId, username, realm string) error {
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

	rows, err := db.Query(query, username, realm, groupId)
	err = resultErrorCheck(rows)
	return err
}

func GroupUpdate(groupId, username, realm string) error {
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

	rows, err := db.Query(query, username, realm, groupId)
	err = resultErrorCheck(rows)
	return err
}
