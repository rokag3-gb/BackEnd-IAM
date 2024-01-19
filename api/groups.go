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
// @Summary 그룹 목록
// @Tags Groups
// @Produce  json
// @Router /groups [get]
// @Success 200 {object} []models.GroupItem
// @Failure 500
func GetGroup(c *gin.Context) {
	realm := c.GetString("realm")
	arr, err := iamdb.GetGroup(realm)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

// token godoc
// @Summary 그룹 생성
// @Tags Groups
// @Produce  json
// @Router /groups [post]
// @Param Body body models.GroupInfo true "body"
// @Success 200 {object} models.Id
// @Failure 400
// @Failure 500
func CreateGroup(c *gin.Context) {
	realm := c.GetString("realm")
	token, _ := clients.KeycloakToken(c, realm)
	var json models.GroupInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	group := gocloak.Group{
		Name: gocloak.StringP(json.Name),
	}

	newGroup, err := clients.KeycloakClient().CreateGroup(c,
		token.AccessToken,
		realm,
		group)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.GroupCreate(newGroup, c.GetString("username"), realm)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, gocloak.Group{ID: gocloak.StringP(newGroup)})
}

// token godoc
// @Summary 그룹 삭제
// @Tags Groups
// @Produce  json
// @Router /groups/{groupId} [delete]
// @Param groupId path string true "Group Id"
// @Success 204
// @Failure 500
func DeleteGroup(c *gin.Context) {
	realm := c.GetString("realm")
	token, _ := clients.KeycloakToken(c, realm)
	groupid := c.Param("groupid")

	err := clients.KeycloakClient().DeleteGroup(c, token.AccessToken, realm, groupid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	c.Status(http.StatusNoContent)
}

// token godoc
// @Summary 그룹 정보 변경
// @Tags Groups
// @Produce  json
// @Router /groups/{groupId} [put]
// @Param Body body models.GroupInfo true "body"
// @Success 204
// @Failure 400
// @Failure 500
func UpdateGroup(c *gin.Context) {
	realm := c.GetString("realm")
	token, _ := clients.KeycloakToken(c, realm)
	groupid := c.Param("groupid")
	var json models.GroupInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
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
