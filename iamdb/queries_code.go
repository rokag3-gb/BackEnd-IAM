package iamdb

import (
	"iam/models"
)

func GetCodeByCodeKey(codeKey string) *models.Code {
	query := `SELECT seq, kindCode, code, codeKey, codeValue, sort, 
				isUse, remark, value1, value2, value3, regDate
				FROM code WHERE codeKey = ?`
	var r models.Code
	db.QueryRow(query, codeKey).Scan(
		&r.ID, &r.KindCode, &r.Code, &r.CodeKey, &r.CodeValue, &r.Sort,
		&r.IsUse, &r.Remark, &r.Value1, &r.Value2, &r.Value3, &r.RegDate)
	if r == (models.Code{}) {
		return nil
	}
	return &r
}

func GetCodeChildsByKindCode(kindCode string) ([]models.Code, error) {
	query := `SELECT seq, kindCode, code, codeKey, codeValue, sort, 
				isUse, remark, value1, value2, value3, regDate
				FROM code WHERE kindCode = ?`

	rows, err := db.Query(query, kindCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.Code, 0)

	for rows.Next() {
		var r models.Code

		err := rows.Scan(
			&r.ID, &r.KindCode, &r.Code, &r.CodeKey, &r.CodeValue, &r.Sort,
			&r.IsUse, &r.Remark, &r.Value1, &r.Value2, &r.Value3, &r.RegDate)

		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}
	return arr, err
}

func GetCodeListByCode(code string) ([]models.Code, error) {
	query := `SELECT seq, kindCode, code, codeKey, codeValue, sort, 
				isUse, remark, value1, value2, value3, regDate
				FROM code WHERE code = ?`

	rows, err := db.Query(query, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr = make([]models.Code, 0)

	for rows.Next() {
		var r models.Code

		err := rows.Scan(
			&r.ID, &r.KindCode, &r.Code, &r.CodeKey, &r.CodeValue, &r.Sort,
			&r.IsUse, &r.Remark, &r.Value1, &r.Value2, &r.Value3, &r.RegDate)

		if err != nil {
			return nil, err
		}

		arr = append(arr, r)
	}
	return arr, err
}

func CreateCodeItem(code models.Code) error {
	query := `INSERT INTO code (kindCode, code, codeValue, sort,
				isUse, remark, value1, value2, value3)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Query(query, code.KindCode, code.Code, code.CodeValue, code.Sort,
		code.IsUse, code.Remark, code.Value1, code.Value2, code.Value3)
	return err
}

func UpdateCodeItemByCodeKey(codeKey string, code models.Code) error {
	query := `UPDATE code SET
				kindCode = ?, code = ?, codeValue = ?, sort = ?,
				isUse = ?, remark = ?, value1 = ?, value2 = ?, value3 = ?
				WHERE codeKey = ?`
	_, err := db.Query(query, code.KindCode, code.Code, code.CodeValue, code.Sort,
		code.IsUse, code.Remark, code.Value1, code.Value2, code.Value3, codeKey)
	return err
}

func DeleteCodeItemByCodeKey(codeKey string) error {
	query := `DELETE FROM code WHERE codeKey = ?`
	_, err := db.Query(query, codeKey)
	return err
}
