package iamdb

func SelectAccount(email, user_id, realm string) (bool, error) {
	ret := false
	db, err := DBClient()
	defer db.Close()
	if err != nil {
		return ret, err
	}

	query := `DECLARE @AccId BIGINT
	DECLARE @UserId varchar(36)
	
	SET @UserId = ?
	SET @AccId = (SELECT AccountId FROM 
		  SALE.dbo.Account
		  WHERE ? LIKE '%' + EmailDomain)
	
	IF @AccId IS NULL
		BEGIN
			-- 계정에 대응되는 Account가 없다는 것을 알려줌
			SELECT 0
		END
	ELSE
		BEGIN
			IF NOT EXISTS (SELECT au.seq FROM SALE.dbo.Account_User au
					WHERE au.AccountId = @AccId
					AND au.UserId = @UserId)
				BEGIN
					INSERT INTO SALE.dbo.Account_User (AccountId, UserId) VALUES (@AccId, @UserId)
				END
			-- 계정에 대응되는 Account가 존재한다는 것을 알려줌
			SELECT 1
		END`

	rows, err := db.Query(query, user_id, email)
	if err != nil {
		return ret, err
	}
	defer rows.Close()

	for rows.Next() {
		var result int

		err := rows.Scan(&result)
		if err != nil {
			return ret, err
		}

		if result == 1 {
			ret = true
		}
	}

	return ret, err
}
