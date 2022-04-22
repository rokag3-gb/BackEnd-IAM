package api

import (
	"encoding/json"
	"errors"
	"iam/iamdb"
	"iam/models"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetRoles(c *gin.Context) {
	rows, err := iamdb.GetRoles()

	if err != nil {
		c.Abort()
	}
	defer rows.Close()

	var arr = make([]models.RolesInfo, 0)

	for rows.Next() {
		var r models.RolesInfo

		err := rows.Scan(&r.ID, &r.Name)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		arr = append(arr, r)
	}

	c.JSON(http.StatusOK, arr)
}

func CreateRoles(c *gin.Context) {
	roles, err := getRoles(c)
	if err != nil {
		return
	}

	if roles.Name == "" {
		c.Status(http.StatusBadRequest)
		c.Abort()
		return
	}

	_, err = iamdb.CreateRoles(roles.Name)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	c.Status(http.StatusCreated)
}

func DeleteRoles(c *gin.Context) {
	roleid, err := getRoleID(c)

	if err == nil {
		return
	}
	_, err = iamdb.DeleteRoles(roleid)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}

func UpdateRoles(c *gin.Context) {
	roleid, err := getRoleID(c)
	if err != nil {
		return
	}

	roles, err := getRoles(c)
	if err != nil {
		return
	}

	if roles.Name == "" {
		c.String(http.StatusBadRequest, "required 'name'")
		c.Abort()
		return
	}

	_, err = iamdb.UpdateRoles(roles.Name, roleid)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}

func GetRolesAuth(c *gin.Context) {
	roleid, err := getRoleID(c)
	if err != nil {
		return
	}

	rows, err := iamdb.GetRolseAuth(roleid)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	var arr = make([]models.RolesInfo, 0)

	for rows.Next() {
		var r models.RolesInfo

		err := rows.Scan(&r.ID, &r.Name)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		arr = append(arr, r)
	}

	c.JSON(http.StatusOK, arr)
}

func AssignRoleAuth(c *gin.Context) {
	roleid, err := getRoleID(c)
	if err != nil {
		return
	}

	auth, err := getAuth(c)
	if err != nil {
		return
	}

	if auth.ID == "" {
		c.String(http.StatusBadRequest, "required 'id'")
		c.Abort()
		return
	}

	_, err = iamdb.AssignRoleAuth(roleid, auth.ID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	c.Status(http.StatusCreated)
}

func DismissRoleAuth(c *gin.Context) {
	roleid, err := getRoleID(c)
	if err != nil {
		return
	}

	authid, err := getAuthID(c)
	if err != nil {
		return
	}

	_, err = iamdb.DismissRoleAuth(roleid, authid)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	c.Status(http.StatusCreated)
}

func getRoles(c *gin.Context) (*models.RolesInfo, error) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Abort()
		return nil, errors.New("required 'body'")
	}

	var r *models.RolesInfo
	json.Unmarshal([]byte(value), &r)

	if r == nil {
		c.Status(http.StatusBadRequest)
		c.Abort()
		return nil, errors.New("required 'body'")
	}

	return r, nil
}

func getAuth(c *gin.Context) (*models.AutuhorityInfo, error) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Abort()
		return nil, errors.New("required 'body'")
	}

	var a *models.AutuhorityInfo
	json.Unmarshal([]byte(value), &a)

	if a == nil {
		c.Status(http.StatusBadRequest)
		c.Abort()
		return nil, errors.New("required 'body'")
	}

	return a, nil
}

func getRoleID(c *gin.Context) (string, error) {
	roleid := c.Param("roleid")

	if roleid == "" {
		c.String(http.StatusBadRequest, "required 'Role id'")
		c.Abort()
		return "", errors.New("required 'Role id'")
	}

	return roleid, nil
}

func getAuthID(c *gin.Context) (string, error) {
	authid := c.Param("authid")

	if authid == "" {
		c.String(http.StatusBadRequest, "required 'Auth id'")
		c.Abort()
		return "", errors.New("required 'Auth id'")
	}

	return authid, nil
}
