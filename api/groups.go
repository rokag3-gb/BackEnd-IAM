package api

import (
	"iam/clients"
	"iam/iamdb"
	"iam/models"
	"net/http"

	logger "cloudmt.co.kr/mateLogger"
	"github.com/Nerzal/gocloak/v11"
	"github.com/gin-gonic/gin"
)

func GetGroup(c *gin.Context) {
	arr, err := iamdb.GetGroup()
	if err != nil {
		logger.Error(err.Error())
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, arr)
}

func CreateGroup(c *gin.Context) {
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
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}

	err = iamdb.GroupCreate(newGroup, c.GetString("username"))
	if err != nil {
		logger.Error(err.Error())
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gocloak.Group{ID: gocloak.StringP(newGroup)})
}

func DeleteGroup(c *gin.Context) {
	// TODO: 그룹 삭제 시 권한 체크 없음. 추후 Authority 기능 구현시 권한 체크 기능 넣을 것.
	token, _ := clients.KeycloakToken(c)
	groupid := c.Param("groupid")

	err := clients.KeycloakClient().DeleteGroup(c, token.AccessToken, clients.KeycloakConfig().Realm, groupid)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}
	c.Status(http.StatusNoContent)
}

func UpdateGroup(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	groupid := c.Param("groupid")
	var json models.GroupInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	groupToUpdate, err := clients.KeycloakClient().GetGroup(c, token.AccessToken, clients.KeycloakConfig().Realm, groupid)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}

	groupToUpdate.Name = gocloak.StringP(json.Name)

	err = clients.KeycloakClient().UpdateGroup(c, token.AccessToken, clients.KeycloakConfig().Realm, *groupToUpdate)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}

	err = iamdb.GroupUpdate(groupid, c.GetString("username"))
	if err != nil {
		logger.Error(err.Error())
		c.Status(http.StatusInternalServerError)
		c.Abort()
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
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}
	c.JSON(http.StatusOK, groups)
}
