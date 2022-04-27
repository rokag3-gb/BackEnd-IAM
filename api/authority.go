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

	if err != nil {
		return
	}

	tx, err := iamdb.DBClient().Begin()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	_, err = iamdb.DeleteRolesAuthByRoleId(roleid, tx)
	if err != nil {
		tx.Rollback()
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	_, err = iamdb.DeleteUserRoleByRoleId(roleid, tx)
	if err != nil {
		tx.Rollback()
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	_, err = iamdb.DeleteRoles(roleid, tx)
	if err != nil {
		tx.Rollback()
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	tx.Commit()
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

	err = iamdb.CheckRoleAuthID(roleid, auth.ID)
	if err != nil {
		c.Status(http.StatusBadRequest)
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

func GetUserRole(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		return
	}

	rows, err := iamdb.GetUserRole(userID)
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

func AssignUserRole(c *gin.Context) {
	userid, err := getUserID(c)
	if err != nil {
		return
	}
	roles, err := getRoles(c)
	if err != nil {
		return
	}
	if roles.ID == "" {
		c.String(http.StatusBadRequest, "required 'id'")
		c.Abort()
		return
	}

	err = iamdb.CheckUserRoleID(userid, roles.ID)
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Abort()
		return
	}

	_, err = iamdb.AssignUserRole(userid, roles.ID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	c.Status(http.StatusCreated)
}

func DismissUserRole(c *gin.Context) {
	userid, err := getUserID(c)
	if err != nil {
		return
	}
	roleid, err := getRoleID(c)
	if err != nil {
		return
	}

	_, err = iamdb.DismissUserRole(userid, roleid)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	c.Status(http.StatusCreated)
}

func GetUserAuth(c *gin.Context) {
	userid, err := getUserID(c)
	if err != nil {
		return
	}

	rows, err := iamdb.GetUserAuth(userid)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	var arr = make([]models.AutuhorityInfo, 0)

	for rows.Next() {
		var r models.AutuhorityInfo

		err := rows.Scan(&r.ID, &r.Name)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		arr = append(arr, r)
	}

	c.JSON(http.StatusOK, arr)
}

func GetUserAuthActive(c *gin.Context) {
	userName, err := getUserID(c)
	if err != nil {
		return
	}

	authName, err := getAuthID(c)
	if err != nil {
		return
	}

	rows, err := iamdb.GetUserAuthActive(userName, authName)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	m := make(map[string]interface{})
	if rows.Next() {
		m["active"] = true
	} else {
		m["active"] = false
	}

	c.JSON(http.StatusOK, m)
}

func GetAuth(c *gin.Context) {
	rows, err := iamdb.GetAuth()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	var arr = make([]models.AutuhorityInfo, 0)

	for rows.Next() {
		var r models.AutuhorityInfo

		err := rows.Scan(&r.ID, &r.Name)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		arr = append(arr, r)
	}

	c.JSON(http.StatusOK, arr)
}

func CreateAuth(c *gin.Context) {
	auth, err := getAuth(c)
	if err != nil {
		return
	}

	if auth.Name == "" {
		c.Status(http.StatusBadRequest)
		c.Abort()
		return
	}

	_, err = iamdb.CreateAuth(auth)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	c.Status(http.StatusCreated)
}

func DeleteAuth(c *gin.Context) {
	authid, err := getAuthID(c)

	if err != nil {
		return
	}

	tx, err := iamdb.DBClient().Begin()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	_, err = iamdb.DeleteRolesAuthByAuthId(authid, tx)
	if err != nil {
		tx.Rollback()
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	_, err = iamdb.DeleteAuth(authid, tx)
	if err != nil {
		tx.Rollback()
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	tx.Commit()
	c.Status(http.StatusNoContent)
}

func UpdateAuth(c *gin.Context) {
	authid, err := getAuthID(c)
	if err != nil {
		return
	}

	auth, err := getAuth(c)
	if err != nil {
		return
	}

	auth.ID = authid

	if err != nil {
		return
	}
	_, err = iamdb.UpdateAuth(auth)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}

func GetAuthInfo(c *gin.Context) {
	authid, err := getAuthID(c)
	if err != nil {
		return
	}

	rows, err := iamdb.GetAuthInfo(authid)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	var r models.AutuhorityInfo
	if rows.Next() {
		err := rows.Scan(&r.ID, &r.Name, &r.URL, &r.Method)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	c.JSON(http.StatusOK, r)
}

////////////////////////////////////////////

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
	json.Unmarshal(value, &a)

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

func getUserID(c *gin.Context) (string, error) {
	userid := c.Param("userid")

	if userid == "" {
		c.String(http.StatusBadRequest, "required 'User id'")
		c.Abort()
		return "", errors.New("required 'User id'")
	}

	return userid, nil
}
