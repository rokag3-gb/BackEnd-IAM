package iamdb

import (
	"database/sql"
	"iam/models"
)

func InsertToken(db *sql.DB, data models.TokenData) error {
	query := `INSERT INTO Token(
		TokenId
		, TokenTypeCode
		, TenantId
		, IssuedAtUTC
		, iat
		, Issuer
		, IssuedUserId
		, ExpiredAtUTC
		, exp
		, SubjectUserId
		, Token
	)
	VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.Query(query,
		data.TokenId,
		data.TokenTypeCode,
		data.TenantId,
		data.IssuedAtUTC,
		data.Iat,
		data.Issuer,
		data.IssuedUserId,
		data.ExpiredAtUTC,
		data.Exp,
		data.SubjectUserId,
		data.Token,
	)
	return err
}

func UpdateTokenConsume(db *sql.DB, jit string) error {
	query := `UPDATE Token SET IsConsumed = 1, ConsumedAtUTC = GETDATE() WHERE TokenId = ? AND IsConsumed = 0`

	_, err := db.Query(query, jit)
	return err
}
