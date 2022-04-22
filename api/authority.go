package api

import (
	"encoding/json"
	"iam/iamdb"
	"iam/models"
	"io/ioutil"
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

	var arr []models.RolesInfo

	for rows.Next() {
		var r models.RolesInfo

		err := rows.Scan(&r.ID, &r.Name)
		if err != nil {
			log.Fatal(err)
		}

		arr = append(arr, r)
	}

	c.JSON(http.StatusOK, arr)
}

func CreateRoles(c *gin.Context) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Abort()
		return
	}

	var r models.RolesInfo
	json.Unmarshal([]byte(value), &r)

	if r.Name == "" {
		c.Status(http.StatusBadRequest)
		c.Abort()
		return
	}

	_, err = iamdb.CreateRoles(r.Name)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	c.Status(http.StatusCreated)
}

func DeleteRoles(c *gin.Context) {
	roleid := c.Param("roleid")

	if roleid == "" {
		c.Status(http.StatusBadRequest)
		c.Abort()
		return
	}

	_, err := iamdb.DeleteRoles(roleid)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}
