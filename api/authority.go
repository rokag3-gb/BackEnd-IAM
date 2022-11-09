package api

import (
	"encoding/json"
	"errors"
	"iam/iamdb"
	"iam/models"
	"io/ioutil"
	"net/http"

	logger "cloudmt.co.kr/mateLogger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetRoles(c *gin.Context) {
	RolesInfos, err := iamdb.GetRoles()

	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, RolesInfos)
}

func CreateRoles(c *gin.Context) {
	roles, err := getRoles(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	if roles.Name == nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "required 'body'")
		return
	}

	db, err := iamdb.DBClient()
	defer db.Close()
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	roleId := uuid.New()

	err = iamdb.CreateRolesIdTx(tx, roleId.String(), *roles.Name, roles.DefaultRole, c.GetString("username"))
	if err != nil {
		tx.Rollback()
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if roles.AuthId != nil {
		for _, authid := range *roles.AuthId {
			err = iamdb.AssignRoleAuthTx(tx, roleId.String(), authid, c.GetString("username"))
			if err != nil {
				tx.Rollback()
				logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
				return
			}
		}
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": roleId.String(),
	})
}

func DeleteRoles(c *gin.Context) {
	roleid, err := getRoleID(c)

	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	db, err := iamdb.DBClient()
	defer db.Close()
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.DeleteRolesAuthByRoleIdTx(tx, roleid)
	if err != nil {
		tx.Rollback()
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.DeleteUserRoleByRoleIdTx(tx, roleid)
	if err != nil {
		tx.Rollback()
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.DeleteRolesTx(tx, roleid)
	if err != nil {
		tx.Rollback()
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	c.Status(http.StatusNoContent)
}

func UpdateRoles(c *gin.Context) {
	roleid, err := getRoleID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	roles, err := getRoles(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	if roles.Name == nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "required 'body'")
		return
	}

	db, err := iamdb.DBClient()
	defer db.Close()
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	roles.ID = roleid

	err = iamdb.UpdateRolesTx(tx, roles, c.GetString("username"))
	if err != nil {
		tx.Rollback()
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if roles.AuthId != nil {
		err = iamdb.DeleteRolesAuthByRoleIdTx(tx, roleid)
		if err != nil {
			tx.Rollback()
			logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
			return
		}

		for _, authid := range *roles.AuthId {
			err = iamdb.AssignRoleAuthTx(tx, roleid, authid, c.GetString("username"))
			if err != nil {
				tx.Rollback()
				logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
				return
			}
		}
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	c.Status(http.StatusNoContent)
}

func GetMyAuth(c *gin.Context) {
	userId := c.GetString("userId")

	arr, err := iamdb.GetMyAuth(userId)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

func GetMenuAuth(c *gin.Context) {
	userId := c.GetString("userId")
	site := c.Param("site")

	if site == "" {
		logger.ErrorProcess(c, nil, http.StatusBadRequest, "")
	}

	arr, err := iamdb.GetMenuAuth(userId, site)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

func GetRolesAuth(c *gin.Context) {
	roleid, err := getRoleID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	arr, err := iamdb.GetRolseAuth(roleid)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

func AssignRoleAuth(c *gin.Context) {
	roleid, err := getRoleID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	auth, err := getAuth(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	if auth.ID == "" {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "required 'Name'")
		return
	}

	err = iamdb.CheckRoleAuthID(roleid, auth.ID)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.AssignRoleAuth(roleid, auth.ID, c.GetString("username"))
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusCreated)
}

func DismissRoleAuth(c *gin.Context) {
	roleid, err := getRoleID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	authid, err := getAuthID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	err = iamdb.DismissRoleAuth(roleid, authid)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

func UpdateRoleAuth(c *gin.Context) {
	use, err := getUse(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	roleid, err := getRoleID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	authid, err := getAuthID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	err = iamdb.UpdateRoleAuth(roleid, authid, use.Use, c.GetString("username"))
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

func GetUserRole(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	arr, err := iamdb.GetUserRole(userID)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

func AssignUserRole(c *gin.Context) {
	userid, err := getUserID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}
	roles, err := getRoles(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}
	if roles.ID == "" {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "required 'id'")
		return
	}

	err = iamdb.CheckUserRoleID(userid, roles.ID)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.AssignUserRole(userid, roles.ID, c.GetString("username"))
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusCreated)
}

func DismissUserRole(c *gin.Context) {
	userid, err := getUserID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}
	roleid, err := getRoleID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	err = iamdb.DismissUserRole(userid, roleid)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

func UpdateUserRole(c *gin.Context) {
	use, err := getUse(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}
	userid, err := getUserID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}
	roleid, err := getRoleID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	err = iamdb.UpdateUserRole(userid, roleid, use.Use, c.GetString("username"))
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

func GetUserAuth(c *gin.Context) {
	userid, err := getUserID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	arr, err := iamdb.GetUserAuth(userid)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

func GetUserAuthActive(c *gin.Context) {
	userName, err := getUserID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	authName, err := getAuthID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	m, err := iamdb.GetUserAuthActive(userName, authName)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, m)
}

func GetAuth(c *gin.Context) {
	arr, err := iamdb.GetAuth()
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

func CreateAuth(c *gin.Context) {
	auth, err := getAuth(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	if auth.Name == "" {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "required 'Name'")
		return
	}

	authId := uuid.New()
	auth.ID = authId.String()

	err = iamdb.CreateAuth(auth, c.GetString("username"))
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": authId.String(),
	})
}

func DeleteAuth(c *gin.Context) {
	authid, err := getAuthID(c)

	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	db, err := iamdb.DBClient()
	defer db.Close()
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.DeleteRolesAuthByAuthIdTx(tx, authid)
	if err != nil {
		tx.Rollback()
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.DeleteAuth(authid, tx)
	if err != nil {
		tx.Rollback()
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	tx.Commit()
	c.Status(http.StatusNoContent)
}

func UpdateAuth(c *gin.Context) {
	authid, err := getAuthID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	auth, err := getAuth(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	if auth.Name == "" {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "required 'Name'")
		return
	}

	auth.ID = authid

	err = iamdb.UpdateAuth(auth, c.GetString("username"))
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

func GetAuthInfo(c *gin.Context) {
	authid, err := getAuthID(c)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	r, err := iamdb.GetAuthInfo(authid)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, r)
}

////////////////////////////////////////////

func getRoles(c *gin.Context) (*models.RolesInfo, error) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return nil, errors.New("required 'body'")
	}

	var r *models.RolesInfo
	json.Unmarshal([]byte(value), &r)

	if r == nil {
		return nil, errors.New("required 'body'")
	}

	return r, nil
}

func getAuth(c *gin.Context) (*models.AutuhorityInfo, error) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return nil, errors.New("required 'body'")
	}

	var a *models.AutuhorityInfo
	json.Unmarshal(value, &a)

	if a == nil {
		return nil, errors.New("required 'body'")
	}

	return a, nil
}

func getRoleID(c *gin.Context) (string, error) {
	roleid := c.Param("roleid")

	if roleid == "" {
		return "", errors.New("required 'Role id'")
	}

	return roleid, nil
}

func getAuthID(c *gin.Context) (string, error) {
	authid := c.Param("authid")

	if authid == "" {
		return "", errors.New("required 'Auth id'")
	}

	return authid, nil
}

func getUserID(c *gin.Context) (string, error) {
	userid := c.Param("userid")

	if userid == "" {
		return "", errors.New("required 'User id'")
	}

	return userid, nil
}

func getUse(c *gin.Context) (*models.AutuhorityUse, error) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return nil, errors.New("required 'body'")
	}

	var a *models.AutuhorityUse
	json.Unmarshal(value, &a)

	if a == nil {
		return nil, errors.New("required 'body'")
	}

	return a, nil
}
