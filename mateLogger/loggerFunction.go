package matelogger

import (
	"fmt"
	"iam/config"

	"github.com/Nerzal/gocloak/v11"
	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/gin-gonic/gin"
)

func ErrorProcess(c *gin.Context, e error, code int, message string) {
	if e != nil {
		var err error
		var errMessage string

		dbError, dbCheck := e.(mssql.Error)
		apiError, keyCheck := e.(*gocloak.APIError)

		if message != "" {
			errMessage = message
		} else {
			errMessage = e.Error()
		}

		if dbCheck {
			ErrorFun(dbError.Error())

			if dbError.Number == 2601 {
				err = fmt.Errorf("cannot insert duplicate data")
			} else {
				err = fmt.Errorf("")
			}

			if config.GetConfig().Developer_mode {
				err = fmt.Errorf("%v\n%v", e, err)
			}

			c.String(code, err.Error())
			c.Abort()
		} else if keyCheck {
			ErrorFun(apiError.Error())
			c.String(apiError.Code, errMessage)
			c.Abort()
		} else {
			ErrorFun(e.Error())
			c.String(code, errMessage)
			c.Abort()
		}
	} else {
		ErrorFun(message)
		c.String(code, message)
		c.Abort()
	}
}
