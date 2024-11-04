package query

import (
	"database/sql"
	"iam/models"
	"strings"
)

func CreateSecretGroupTx(tx *sql.Tx, secretGroupPath, username, realm string) error {
	query := `INSERT INTO vSecretGroup(vSecretGroupPath, REALM_ID, createId, modifyId)
	select ?, ?, ID, ID from USER_ENTITY WHERE USERNAME = ? AND REALM_ID = ?`

	_, err := tx.Query(query, secretGroupPath, realm, username, realm)
	return err
}

func DeleteSecretGroupTx(tx *sql.Tx, secretGroupPath, realm string) error {
	query := `DELETE FROM vSecretGroup WHERE vSecretGroupPath = ? AND REALM_ID = ?`

	_, err := tx.Query(query, secretGroupPath, realm)
	return err
}

func DeleteSecretBySecretGroupTx(tx *sql.Tx, secretGroupPath, realm string) error {
	query := `DELETE FROM vSecret WHERE vSecretGroupId = (SELECT vSecretGroupId FROM vSecretGroup WHERE vSecretGroupPath = ? AND REALM_ID = ?)`

	_, err := tx.Query(query, secretGroupPath, realm)
	return err
}

func MergeSecret(db *sql.DB, secretPath, secretGroupPath, username, realm string, url *string) error {
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

	_, err := db.Query(query, secretPath, secretGroupPath, realm, username, realm, url)

	return err
}

func DeleteSecret(db *sql.DB, secretPath, secretGroupPath string) error {
	query := `DELETE FROM vSecret WHERE vSecretPath = ?
	AND vSecretGroupId = (select vSecretGroupId from vSecretGroup where vSecretGroupPath = ?)
	SELECT @@ROWCOUNT`

	rows, err := db.Query(query, secretPath, secretGroupPath)
	err = resultErrorCheck(rows)
	return err
}

func GetSecretGroup(db *sql.DB, data []models.SecretGroupItem, username, realm string) ([]models.SecretGroupItem, error) {
	queryParams := []interface{}{}

	query := `declare @values table
	(
		sg varchar(310)
	)`
	for _, d := range data {
		queryParams = append(queryParams, "/iam/secret/"+d.Name+"/")
		query += `insert into @values values (?)`
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
	join UserRole ur on u.ID = ur.userId
	join roles_authority_mapping ra on ur.RoleId = ra.rId
	join authority a on ra.aId = a.aId
	where ra.useYn = 'true'
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

	queryParams = append(queryParams, username, realm)
	rows, err := db.Query(query, queryParams...)
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

func GetSecret(db *sql.DB, groupName, realm string) (map[string]models.SecretItem, error) {
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
	WHERE s.vSecretGroupId = (SELECT vSecretGroupId FROM vSecretGroup WHERE vSecretGroupPath = ? AND REALM_ID = ?)
	ORDER BY s.vSecretPath`

	rows, err := db.Query(query, groupName, realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var m = make(map[string]models.SecretItem)

	for rows.Next() {
		var r models.SecretItem

		err := rows.Scan(&r.Name, &r.Url, &r.CreateDate, &r.Creator, &r.ModifyDate, &r.Modifier)
		if err != nil {
			return nil, err
		}

		m[r.Name] = r
	}
	return m, err
}

func GetSecretGroupMetadata(db *sql.DB, groupName, realm string) (models.SecretGroupResponse, error) {
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

	rows, err := db.Query(query, groupName+"_MANAGER", realm)
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
	JOIN UserRole ur
	on u.ID = ur.userId
	JOIN roles r
	ON ur.RoleId = r.rId
	where r.rName = ?
	AND r.REALM_ID = ?`

	rows, err = db.Query(query, groupName+"_Manager", realm)
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

	rows, err = db.Query(query, groupName, realm)
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

func GetSecretByName(db *sql.DB, groupName, secretName, realm string) (*models.SecretItem, error) {
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

	rows, err := db.Query(query, groupName, realm, secretName)
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

func GetAllSecret(db *sql.DB, data []models.SecretGroupItem, realm string) ([]models.SecretItem, error) {
	var groups []string

	args := []interface{}{}
	for _, group := range data {
		groups = append(groups, group.Name)
		args = append(args, group.Name)
	}
	args = append(args, realm)

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
