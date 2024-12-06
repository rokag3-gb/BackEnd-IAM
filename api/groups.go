package api

import (
	"iam/clients"
	"iam/common"
	"iam/iamdb"
	"iam/models"
	"net/http"

	"github.com/Nerzal/gocloak/v11"
	"github.com/gin-gonic/gin"
)

// token godoc
// @Security Bearer
// @Summary 그룹 목록
// @Tags Groups
// @Produce  json
// @Router /groups [get]
// @Success 200 {object} []models.GroupItem
// @Failure 500
func GetGroup(c *gin.Context) {
	arr, err := iamdb.GetGroup()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

// token godoc
// @Security Bearer
// @Summary 그룹 생성
// @Tags Groups
// @Produce  json
// @Router /groups [post]
// @Param Body body models.GroupInfo true "body"
// @Success 200 {object} models.Id
// @Failure 400
// @Failure 500
func CreateGroup(c *gin.Context) {
	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	var json models.GroupInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}
	if json.Realm == "" {
		json.Realm = c.GetString("realm")
	}

	group := gocloak.Group{
		Name: gocloak.StringP(json.Name),
	}

	newGroup, err := clients.KeycloakClient().CreateGroup(c,
		token.AccessToken,
		json.Realm,
		group)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.GroupCreate(newGroup, c.GetString("username"), json.Realm)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, gocloak.Group{ID: gocloak.StringP(newGroup)})
}

// token godoc
// @Security Bearer
// @Summary 그룹 삭제
// @Tags Groups
// @Produce  json
// @Router /groups/{groupId} [delete]
// @Param realm path string true "Realm Id"
// @Param groupId path string true "Group Id"
// @Success 204
// @Failure 500
func DeleteGroup(c *gin.Context) {
	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	groupid := c.Param("groupid")

	realm, err := iamdb.GetGroupRealmById(groupid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = clients.KeycloakClient().DeleteGroup(c, token.AccessToken, realm, groupid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	c.Status(http.StatusNoContent)
}

// token godoc
// @Security Bearer
// @Summary 그룹 정보 변경
// @Tags Groups
// @Produce  json
// @Router /groups/{groupId} [put]
// @Param realm path string true "Realm Id"
// @Param Body body models.GroupInfo true "body"
// @Success 204
// @Failure 400
// @Failure 500
func UpdateGroup(c *gin.Context) {
	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	groupid := c.Param("groupid")
	var json models.GroupInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	realm, err := iamdb.GetGroupRealmById(groupid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	groupToUpdate, err := clients.KeycloakClient().GetGroup(c, token.AccessToken, realm, groupid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	groupToUpdate.Name = gocloak.StringP(json.Name)

	err = clients.KeycloakClient().UpdateGroup(c, token.AccessToken, realm, *groupToUpdate)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.GroupUpdate(groupid, c.GetString("username"), realm)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}
