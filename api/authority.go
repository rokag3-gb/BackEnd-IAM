package api

import (
	"encoding/json"
	"errors"
	"iam/common"
	"iam/iamdb"
	"iam/models"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// token godoc
// @Summary 전체 역할 목록 조회
// @Tags Authority
// @Produce  json
// @Router /authority/roles [get]
// @Param realm path string true "사용되지 않음"
// @Success 200 {object} []models.RolesInfo
// @Failure 500
func GetRoles(c *gin.Context) {
	RolesInfos, err := iamdb.GetRoles()

	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, RolesInfos)
}

// token godoc
// @Summary 역할 생성
// @Tags Authority
// @Produce  json
// @Param Body body models.RolesInfo true "body"
// @Router /authority/roles [post]
// @Success 200 {object} models.Id
// @Failure 400
// @Failure 500
func CreateRoles(c *gin.Context) {
	roles, err := getRoles(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	if roles.Name == nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "required 'body'")
		return
	}

	if roles.Realm == "" {
		common.ErrorProcess(c, err, http.StatusBadRequest, "required 'Realm'")
		return
	}
	db, err := iamdb.DBClient()
	defer db.Close()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	roleId := uuid.New()

	err = iamdb.CreateRolesIdTx(tx, roleId.String(), *roles.Name, roles.Realm, c.GetString("userId"), roles.DefaultRole)
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if roles.AuthId != nil {
		for _, authId := range *roles.AuthId {
			err = iamdb.AssignRoleAuthTx(tx, roleId.String(), authId, c.GetString("userId"))
			if err != nil {
				tx.Rollback()
				common.ErrorProcess(c, err, http.StatusInternalServerError, "")
				return
			}
		}
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": roleId.String(),
	})
}

// token godoc
// @Summary 역할 삭제
// @Tags Authority
// @Param roleId path string true "Role Id"
// @Router /authority/roles/{roleId} [delete]
// @Success 204
// @Failure 400
// @Failure 500
func DeleteRoles(c *gin.Context) {
	roleId, err := getRoleID(c)

	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	db, err := iamdb.DBClient()
	defer db.Close()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.DeleteRolesAuthByRoleIdTx(tx, roleId)
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.DeleteUserRoleByRoleIdTx(tx, roleId)
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.DeleteRolesTx(tx, roleId)
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	c.Status(http.StatusNoContent)
}

// token godoc
// @Summary 역할 수정
// @Tags Authority
// @Produce  json
// @Param roleId path string true "Role Id"
// @Param Body body models.RolesInfo true "body"
// @Router /authority/roles/{roleId} [put]
// @Success 204
// @Failure 400
// @Failure 500
func UpdateRoles(c *gin.Context) {
	roleId, err := getRoleID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	roles, err := getRoles(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	db, err := iamdb.DBClient()
	defer db.Close()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	roles.ID = roleId

	err = iamdb.UpdateRolesTx(tx, roles, c.GetString("userId"))
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if roles.AuthId != nil {
		err = iamdb.DeleteRolesAuthByRoleIdTx(tx, roleId)
		if err != nil {
			tx.Rollback()
			common.ErrorProcess(c, err, http.StatusInternalServerError, "")
			return
		}

		for _, authId := range *roles.AuthId {
			err = iamdb.AssignRoleAuthTx(tx, roleId, authId, c.GetString("username"))
			if err != nil {
				tx.Rollback()
				common.ErrorProcess(c, err, http.StatusInternalServerError, "")
				return
			}
		}
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	c.Status(http.StatusNoContent)
}

// token godoc
// @Summary 유저 권한 조회
// @Tags Authority
// @Produce  json
// @Router /authority/user/auth [get]
// @Success 200 {object} []string
// @Failure 500
func GetMyAuth(c *gin.Context) {
	userId := c.GetString("userId")
	tenantId := c.GetString("tenantId")

	arr, err := iamdb.GetMyAuth(userId, tenantId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

// token godoc
// @Summary 메뉴 권한 조회
// @Tags Authority
// @Produce  json
// @Param site path string true "Site"
// @Router /authority/auth/menu/{site} [get]
// @Param tenantId path string true "tenantId"
// @Success 200 {object} []models.MenuAutuhorityInfo
// @Failure 400
// @Failure 500
func GetMenuAuth(c *gin.Context) {
	realm := c.GetString("realm")
	userId := c.GetString("userId")
	site := c.Param("site")

	if site == "" {
		common.ErrorProcess(c, nil, http.StatusBadRequest, "")
	}

	arr, err := iamdb.GetMenuAuth(userId, site, realm)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

// token godoc
// @Summary 역할 할당 권한 목록 조회
// @Tags Authority
// @Produce  json
// @Param roleId path string true "Role Id"
// @Router /authority/roles/{roleId}/auth [get]
// @Success 200 {object} []models.RolesInfo
// @Failure 400
// @Failure 500
func GetRolesAuth(c *gin.Context) {
	roleId, err := getRoleID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	arr, err := iamdb.GetRolseAuth(roleId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

// token godoc
// @Summary 역할 권한 할당
// @Tags Authority
// @Produce  json
// @Param roleId path string true "Role Id"
// @Router /authority/roles/{roleId}/auth [post]
// @Success 201
// @Failure 400
// @Failure 500
func AssignRoleAuth(c *gin.Context) {
	roleId, err := getRoleID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	auth, err := getAuth(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	if auth.ID == "" {
		common.ErrorProcess(c, err, http.StatusBadRequest, "required 'Name'")
		return
	}

	err = iamdb.CheckRoleAuthID(roleId, auth.ID)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.AssignRoleAuth(roleId, auth.ID, c.GetString("username"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusCreated)
}

// token godoc
// @Summary 역할 권한 제외
// @Tags Authority
// @Produce  json
// @Param roleId path string true "Role Id"
// @Param authId path string true "Auth Id"
// @Router /authority/roles/{roleId}/auth/{authId} [delete]
// @Success 204
// @Failure 400
// @Failure 500
func DismissRoleAuth(c *gin.Context) {
	roleId, err := getRoleID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	authId, err := getAuthID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	err = iamdb.DismissRoleAuth(roleId, authId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Summary 역할 권한 수정
// @Tags Authority
// @Produce  json
// @Param roleId path string true "Role Id"
// @Param authId path string true "Auth Id"
// @Router /authority/roles/{roleId}/auth/{authId} [put]
// @Success 204
// @Failure 400
// @Failure 500
func UpdateRoleAuth(c *gin.Context) {
	use, err := getUse(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	roleId, err := getRoleID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	authId, err := getAuthID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	err = iamdb.UpdateRoleAuth(c.GetString("userId"), roleId, authId, use.Use)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Summary 유저 할당 역할 목록 조회
// @Tags Authority
// @Produce  json
// @Param realm path string true "Realm Id"
// @Param userId path string true "User Id"
// @Router /authority/user/{userId} [get]
// @Success 200 {object} []models.RolesInfo
// @Failure 400
// @Failure 500
func GetUserRole(c *gin.Context) {
	userId, err := getUserID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	arr, err := iamdb.GetUserRole(userId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

// token godoc
// @Summary 유저 역할 할당
// @Tags Authority
// @Produce  json
// @Param tenantId query string true "tenantId"
// @Param userId path string true "User Id"
// @Param roleId body models.Id true "Role Id"
// @Router /authority/user/{userId}/roles [post]
// @Success 201
// @Failure 400
// @Failure 500
func AssignUserRole(c *gin.Context) {
	userid, err := getUserID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	tenantId := c.Query("tenantId")
	if tenantId == "" || tenantId == "<nil>" {
		tenantId = c.GetString("tenantId")
	}

	roles, err := getRoles(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}
	if roles.ID == "" || tenantId == "" {
		common.ErrorProcess(c, err, http.StatusBadRequest, "required 'id'")
		return
	}

	err = iamdb.CheckUserRoleID(userid, roles.ID)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.AssignUserRole(userid, tenantId, roles.ID, c.GetString("userId"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusCreated)
}

// token godoc
// @Summary 유저 역할 제외
// @Tags Authority
// @Produce  json
// @Param realm path string true "Realm Id"
// @Param tenantId query string true "tenantId"
// @Param userId path string true "User Id"
// @Param roleId path string true "Role Id"
// @Router /authority/{tenantId}/user/{userId}/roles/{roleId} [delete]
// @Success 204
// @Failure 400
// @Failure 500
func DismissUserRole(c *gin.Context) {
	userid, err := getUserID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}
	roleId, err := getRoleID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	tenantId := c.Query("tenantId")
	if tenantId == "" {
		tenantId = c.GetString("tenantId")
	}

	err = iamdb.DismissUserRole(userid, roleId, tenantId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Summary 그룹 역할 할당
// @Tags Authority
// @Produce  json
// @Param tenantId query string true "tenantId"
// @Param groupId path string true "Group Id"
// @Param roleId body models.Id true "Role Id"
// @Router /authority/group/{groupId}/roles [post]
// @Success 201
// @Failure 400
// @Failure 500
func AssignGroupRole(c *gin.Context) {
	userid, err := getGroupID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	tenantId := c.Query("tenantId")
	if tenantId == "" || tenantId == "<nil>" {
		tenantId = c.GetString("tenantId")
	}

	roles, err := getRoles(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}
	if roles.ID == "" || tenantId == "" {
		common.ErrorProcess(c, err, http.StatusBadRequest, "required 'id'")
		return
	}

	err = iamdb.CheckUserRoleID(userid, roles.ID)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.AssignUserRole(userid, tenantId, roles.ID, c.GetString("userId"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusCreated)
}

// token godoc
// @Summary 그룹 역할 제외
// @Tags Authority
// @Produce  json
// @Param realm path string true "Realm Id"
// @Param tenantId query string true "tenantId"
// @Param groupId path string true "Group Id"
// @Param roleId path string true "Role Id"
// @Router /authority/{tenantId}/group/{groupId}/roles/{roleId} [delete]
// @Success 204
// @Failure 400
// @Failure 500
func DismissGroupRole(c *gin.Context) {
	userid, err := getGroupID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}
	roleId, err := getRoleID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	tenantId := c.Query("tenantId")
	if tenantId == "" {
		tenantId = c.GetString("tenantId")
	}

	err = iamdb.DismissUserRole(userid, roleId, tenantId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Summary 유저 역할 수정 (현재 사용되지 않음)
// @Tags Authority
// @Produce  json
// @Param realm path string true "Realm Id"
// @Param userId path string true "User Id"
// @Param roleId path string true "Role Id"
// @Param tenantId path string true "tenantId"
// @Param autuhorityUse body models.AutuhorityUse true "AutuhorityUse Yn"
// @Router /authority/user/{tenantId}/{userId}/roles/{roleId} [put]
// @Success 204
// @Failure 400
// @Failure 500
func UpdateUserRole(c *gin.Context) {
	use, err := getUse(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}
	userid, err := getUserID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}
	roleId, err := getRoleID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	tenantId := c.Param("tenantId")
	if tenantId == "" {
		common.ErrorProcess(c, errors.New("required 'User id'"), http.StatusBadRequest, "")
		return
	}

	err = iamdb.UpdateUserRole(userid, roleId, tenantId, c.GetString("userId"), use.Use)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Summary 유저 활성 권한 목록 조회
// @Tags Authority
// @Produce  json
// @Param realm path string true "Realm Id"
// @Param userId path string true "User Id"
// @Param tenantId query string true "tenantId"
// @Router /authority/user/{userId}/auth [get]
// @Success 200 {object} []models.AutuhorityInfo
// @Failure 400
// @Failure 500
func GetUserAuth(c *gin.Context) {
	tenantId := c.Query("tenantId")
	if tenantId == "" {
		tenantId = c.GetString("tenantId")
	}

	userid, err := getUserID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	arr, err := iamdb.GetUserAuth(userid, tenantId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

// token godoc
// @Summary 유저 권한 보유 여부 질의
// @Tags Authority
// @Produce  json
// @Param realm path string true "Realm Id"
// @Param userName path string true "User Name"
// @Param authName path string true "Auth Name"
// @Param tenantId query string true "tenantId"
// @Router /authority/user/{userName}/auth/{authName} [get]
// @Success 200 {object} models.Active
// @Failure 400
// @Failure 500
func GetUserAuthActive(c *gin.Context) {
	tenantId := c.Query("tenantId")
	if tenantId == "" {
		tenantId = c.GetString("tenantId")
	}
	userName, err := getUserID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	authName, err := getAuthID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	m, err := iamdb.GetUserAuthActive(userName, authName, tenantId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, m)
}

// token godoc
// @Summary 전체 권한 목록 조회
// @Tags Authority
// @Produce  json
// @Param realm path string true "Realm Id"
// @Router /authority/auth [get]
// @Success 200 {object} []models.AutuhorityInfo
// @Failure 500
func GetAuth(c *gin.Context) {
	arr, err := iamdb.GetAuth()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

// token godoc
// @Summary 권한 생성
// @Tags Authority
// @Produce  json
// @Param realm path string true "Realm Id"
// @Param Body body models.AutuhorityInfo true "body"
// @Router /authority/auth [post]
// @Success 200 {object} models.Id
// @Failure 400
// @Failure 500
func CreateAuth(c *gin.Context) {
	auth, err := getAuth(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	if auth.Name == "" {
		common.ErrorProcess(c, err, http.StatusBadRequest, "required 'Name'")
		return
	}
	if auth.Realm == "" {
		common.ErrorProcess(c, err, http.StatusBadRequest, "required 'Realm'")
		return
	}

	authId := uuid.New()
	auth.ID = authId.String()

	err = iamdb.CreateAuth(auth, c.GetString("userId"), auth.Realm)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": authId.String(),
	})
}

// token godoc
// @Summary 권한 삭제
// @Tags Authority
// @Produce  json
// @Param realm path string true "Realm Id"
// @Param authId path string true "Auth Name"
// @Router /authority/auth/{authId} [delete]
// @Success 200 {object} models.Id
// @Failure 400
// @Failure 500
func DeleteAuth(c *gin.Context) {
	realm := c.GetString("realm")
	authId, err := getAuthID(c)

	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	db, err := iamdb.DBClient()
	defer db.Close()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.DeleteRolesAuthByAuthIdTx(tx, authId)
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.DeleteAuth(tx, authId, realm)
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	tx.Commit()
	c.Status(http.StatusNoContent)
}

// token godoc
// @Summary 권한 수정
// @Tags Authority
// @Produce  json
// @Param realm path string true "Realm Id"
// @Param authId path string true "Auth Name"
// @Param Body body models.AutuhorityInfo true "body"
// @Router /authority/auth/{authId} [put]
// @Success 204
// @Failure 400
// @Failure 500
func UpdateAuth(c *gin.Context) {
	authId, err := getAuthID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	auth, err := getAuth(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	if auth.Name == "" {
		common.ErrorProcess(c, err, http.StatusBadRequest, "required 'Name'")
		return
	}

	auth.ID = authId

	err = iamdb.UpdateAuth(auth, c.GetString("userId"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Summary 권한 정보 조회
// @Tags Authority
// @Produce  json
// @Param authId path string true "Auth Id"
// @Router /authority/auth/{authId} [get]
// @Success 200 {object} models.AutuhorityInfo
// @Failure 400
// @Failure 500
func GetAuthInfo(c *gin.Context) {
	authId, err := getAuthID(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	r, err := iamdb.GetAuthInfo(authId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, r)
}

////////////////////////////////////////////

func getRoles(c *gin.Context) (*models.RolesInfo, error) {
	value, err := io.ReadAll(c.Request.Body)
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
	value, err := io.ReadAll(c.Request.Body)
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
	roleId := c.Param("roleId")

	if roleId == "" {
		return "", errors.New("required 'Role id'")
	}

	return roleId, nil
}

func getAuthID(c *gin.Context) (string, error) {
	authId := c.Param("authId")

	if authId == "" {
		return "", errors.New("required 'Auth id'")
	}

	return authId, nil
}

func getUserID(c *gin.Context) (string, error) {
	userid := c.Param("userid")

	if userid == "" {
		return "", errors.New("required 'User id'")
	}

	return userid, nil
}

func getGroupID(c *gin.Context) (string, error) {
	userid := c.Param("groupId")

	if userid == "" {
		return "", errors.New("required 'Group id'")
	}

	return userid, nil
}

func getUse(c *gin.Context) (*models.AutuhorityUse, error) {
	value, err := io.ReadAll(c.Request.Body)
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
