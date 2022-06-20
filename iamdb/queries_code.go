package iamdb

import (
	"iam/models"
)

func GetCodeByCodeKey(codeKey string) *models.Code {
	query := `SELECT seq, kindCode, code, codeKey, codeValue, sort, isUse, remark, value1, value2, value3, regDate
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
