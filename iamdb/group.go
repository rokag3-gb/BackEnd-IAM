package iamdb

import (
	"errors"
	"iam/models"
)

func GetGroup() ([]models.GroupItem, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return nil, dbErr
	}

	query := `SELECT g.ID, 
	g.NAME, 
	g.REALM_ID, 
	ISNULL((select count(USER_ID) from USER_GROUP_MEMBERSHIP where GROUP_ID = g.ID group by GROUP_ID), 0) as countMembers,
	FORMAT(g.createDate, 'yyyy-MM-dd HH:mm') as createDate, 
	u1.USERNAME as Creator, 
	FORMAT(g.modifyDate, 'yyyy-MM-dd HH:mm') as modifyDate, 
	u2.USERNAME as Modifier
from KEYCLOAK_GROUP g
	LEFT OUTER JOIN USER_ENTITY u1 on g.createId = u1.ID
	LEFT OUTER JOIN USER_ENTITY u2 on g.modifyId = u2.ID
	order by NAME`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.GroupItem, 0)

	for rows.Next() {
		var r models.GroupItem

		err := rows.Scan(&r.ID, &r.Name, &r.Realm, &r.CountMembers, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
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

func GetGroupRealmById(groupId string) (string, error) {
	db, dbErr := DBClient()
	defer db.Close()
	if dbErr != nil {
		return "", dbErr
	}

	query := `SELECT g.REALM_ID from KEYCLOAK_GROUP g WHERE ID = ?`

	rows, err := db.Query(query, groupId)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var realm string
		err := rows.Scan(&realm)
		if err != nil {
			return "", err
		}

		return realm, nil
	}
	return "", errors.New("Realm not found")
}
