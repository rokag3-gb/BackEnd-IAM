package api

import (
	"iam/iamdb"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetRoles(c *gin.Context) {
	rows, err := iamdb.GetRoles()

	if err != nil {
		c.Abort()
	}
	defer rows.Close()

	var arr []map[string]interface{}

	for rows.Next() {
		var rId int
		var rName string

		err := rows.Scan(&rId, &rName)
		if err != nil {
			log.Fatal(err)
		}
		m := make(map[string]interface{})
		m["rId"] = rId
		m["rName"] = rName

		arr = append(arr, m)
	}

	c.JSON(http.StatusOK, arr)
}
