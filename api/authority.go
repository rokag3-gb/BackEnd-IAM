package api

import (
	"encoding/json"
	"errors"
	"iam/config"
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
		logger.Error(err.Error())
		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {

			if config.GetConfig().Developer_mode {
				c.String(http.StatusInternalServerError, err.Error())
			} else {
				c.Status(http.StatusInternalServerError)
			}
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, RolesInfos)
}

func CreateRoles(c *gin.Context) {
	roles, err := getRoles(c)
	if err != nil {
		return
	}

	if roles.Name == nil {
		c.Status(http.StatusBadRequest)
		c.Abort()
		return
	}

	tx, err := iamdb.DBClient().Begin()
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}
	roleId := uuid.New()

	err = iamdb.CreateRolesIdTx(tx, roleId.String(), *roles.Name, c.GetString("username"))
	if err != nil {
		tx.Rollback()
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	if roles.AuthId != nil {
		for _, authid := range *roles.AuthId {
			err = iamdb.AssignRoleAuthTx(tx, roleId.String(), authid, c.GetString("username"))
			if err != nil {
				tx.Rollback()
				logger.Error(err.Error())

				if config.GetConfig().Developer_mode {
					c.String(http.StatusInternalServerError, err.Error())
				} else {
					c.Status(http.StatusInternalServerError)
				}
				c.Abort()
				return
			}
		}
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": roleId.String(),
	})
}

func DeleteRoles(c *gin.Context) {
	roleid, err := getRoleID(c)

	if err != nil {
		return
	}

	tx, err := iamdb.DBClient().Begin()
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	err = iamdb.DeleteRolesAuthByRoleIdTx(tx, roleid)
	if err != nil {
		tx.Rollback()
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	err = iamdb.DeleteUserRoleByRoleIdTx(tx, roleid)
	if err != nil {
		tx.Rollback()
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	err = iamdb.DeleteRolesTx(tx, roleid)
	if err != nil {
		tx.Rollback()
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
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

	if roles.Name == nil && roles.DefaultRole == nil {
		c.String(http.StatusBadRequest, "required 'body'")
		c.Abort()
		return
	}

	tx, err := iamdb.DBClient().Begin()
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	roles.ID = roleid

	err = iamdb.UpdateRolesTx(tx, roles, c.GetString("username"))
	if err != nil {
		tx.Rollback()
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	if roles.AuthId != nil {
		err = iamdb.DeleteRolesAuthByRoleIdTx(tx, roleid)
		if err != nil {
			tx.Rollback()
			logger.Error(err.Error())

			if config.GetConfig().Developer_mode {
				c.String(http.StatusInternalServerError, err.Error())
			} else {
				c.Status(http.StatusInternalServerError)
			}
			c.Abort()
			return
		}

		for _, authid := range *roles.AuthId {
			err = iamdb.AssignRoleAuthTx(tx, roleid, authid, c.GetString("username"))
			if err != nil {
				tx.Rollback()
				logger.Error(err.Error())

				if config.GetConfig().Developer_mode {
					c.String(http.StatusInternalServerError, err.Error())
				} else {
					c.Status(http.StatusInternalServerError)
				}
				c.Abort()
				return
			}
		}
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
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

	arr, err := iamdb.GetRolseAuth(roleid)
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
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
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	err = iamdb.AssignRoleAuth(roleid, auth.ID, c.GetString("username"))
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
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

	err = iamdb.DismissRoleAuth(roleid, authid)
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}

func UpdateRoleAuth(c *gin.Context) {
	use, err := getUse(c)
	if err != nil {
		return
	}

	roleid, err := getRoleID(c)
	if err != nil {
		return
	}

	authid, err := getAuthID(c)
	if err != nil {
		return
	}

	err = iamdb.UpdateRoleAuth(roleid, authid, use.Use, c.GetString("username"))
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}

func GetUserRole(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		return
	}

	arr, err := iamdb.GetUserRole(userID)
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
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
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	err = iamdb.AssignUserRole(userid, roles.ID, c.GetString("username"))
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
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

	err = iamdb.DismissUserRole(userid, roleid)
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}

func UpdateUserRole(c *gin.Context) {
	use, err := getUse(c)
	if err != nil {
		return
	}
	userid, err := getUserID(c)
	if err != nil {
		return
	}
	roleid, err := getRoleID(c)
	if err != nil {
		return
	}

	err = iamdb.UpdateUserRole(userid, roleid, use.Use, c.GetString("username"))
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}

func GetUserAuth(c *gin.Context) {
	userid, err := getUserID(c)
	if err != nil {
		return
	}

	arr, err := iamdb.GetUserAuth(userid)
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
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

	m, err := iamdb.GetUserAuthActive(userName, authName)
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, m)
}

func GetAuth(c *gin.Context) {
	arr, err := iamdb.GetAuth()
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, arr)
}

func CreateAuth(c *gin.Context) {
	auth, err := getAuth(c)
	if err != nil {
		return
	}

	if auth.Name == "" {
		c.String(http.StatusBadRequest, "required 'Name'")
		c.Abort()
		return
	}

	authId := uuid.New()
	auth.ID = authId.String()

	err = iamdb.CreateAuth(auth, c.GetString("username"))
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": authId.String(),
	})
}

func DeleteAuth(c *gin.Context) {
	authid, err := getAuthID(c)

	if err != nil {
		return
	}

	tx, err := iamdb.DBClient().Begin()
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	err = iamdb.DeleteRolesAuthByAuthIdTx(tx, authid)
	if err != nil {
		tx.Rollback()
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	err = iamdb.DeleteAuth(authid, tx)
	if err != nil {
		tx.Rollback()
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
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

	err = iamdb.UpdateAuth(auth, c.GetString("username"))
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
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

	r, err := iamdb.GetAuthInfo(authid)
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, r)
}

////////////////////////////////////////////

func getRoles(c *gin.Context) (*models.RolesInfo, error) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, "required 'body'")
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
		c.String(http.StatusBadRequest, "required 'body'")
		c.Abort()
		return nil, errors.New("required 'body'")
	}

	var a *models.AutuhorityInfo
	json.Unmarshal(value, &a)

	if a == nil {
		c.String(http.StatusBadRequest, "required 'body'")
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

func getUse(c *gin.Context) (*models.AutuhorityUse, error) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, "required 'body'")
		c.Abort()
		return nil, errors.New("required 'body'")
	}

	var a *models.AutuhorityUse
	json.Unmarshal(value, &a)

	if a == nil {
		c.String(http.StatusBadRequest, "required 'body'")
		c.Abort()
		return nil, errors.New("required 'body'")
	}

	return a, nil
}
