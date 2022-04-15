package api

import (
	"iam/clients"
	"iam/models"
	"net/http"

	"github.com/Nerzal/gocloak/v11"
	"github.com/gin-gonic/gin"
)

func GetGroup(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)

	params := gocloak.GetGroupsParams{
		First: gocloak.IntP(c.MustGet("first").(int)),
		Max:   gocloak.IntP(c.MustGet("max").(int)),
	}

	groups, err := clients.KeycloakClient().GetGroups(c,
		token.AccessToken,
		clients.KeycloakConfig().Realm,
		params)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, groups)
}

func CreateGroup(c *gin.Context) {
	// TODO: 그룹 생성 시 권한 체크 없음. 추후 Authority 기능 구현시 권한 체크 기능 넣을 것.
	// TODO: 그룹 생성 시 이름 외에 다른 인자도 받아야 하는지 추후 논의.

	token, _ := clients.KeycloakToken(c)
	var json models.GroupInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	group := gocloak.Group{
		Name: gocloak.StringP(json.Name),
	}

	newGroup, err := clients.KeycloakClient().CreateGroup(c,
		token.AccessToken,
		clients.KeycloakConfig().Realm,
		group)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gocloak.Group{ID: gocloak.StringP(newGroup)})
}

func DeleteGroup(c *gin.Context) {
	c.String(http.StatusOK, "deletegroup")
}

func UpdateGroup(c *gin.Context) {
	// TODO: 그룹 수정 시 권한 체크 없음. 추후 Authority 기능 구현시 권한 체크 기능 넣을 것.
	token, _ := clients.KeycloakToken(c)
	groupid := c.Param("groupid")
	var json models.GroupInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	groupToUpdate, err := clients.KeycloakClient().GetGroup(c, token.AccessToken, clients.KeycloakConfig().Realm, groupid)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	groupToUpdate.Name = gocloak.StringP(json.Name)

	err = clients.KeycloakClient().UpdateGroup(c, token.AccessToken, clients.KeycloakConfig().Realm, *groupToUpdate)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func GetGroupMember(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)

	groupid := c.Param("groupid")

	params := gocloak.GetGroupsParams{
		First: gocloak.IntP(c.MustGet("first").(int)),
		Max:   gocloak.IntP(c.MustGet("max").(int)),
	}

	groups, err := clients.KeycloakClient().GetGroupMembers(c,
		token.AccessToken,
		clients.KeycloakConfig().Realm,
		groupid,
		params)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, groups)
}
