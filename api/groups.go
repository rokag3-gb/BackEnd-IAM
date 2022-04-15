package api

import (
	"iam/clients"
	"net/http"

	"github.com/Nerzal/gocloak/v11"
	"github.com/gin-gonic/gin"
)

func GetGroup(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)

	first := c.MustGet("first").(int)
	max := c.MustGet("max").(int)
	params := gocloak.GetGroupsParams{
		First: &first,
		Max:   &max,
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
	c.String(http.StatusOK, "creategroup")
}

func DeleteGroup(c *gin.Context) {
	c.String(http.StatusOK, "deletegroup")
}

func UpdateGroup(c *gin.Context) {
	c.String(http.StatusOK, "updategroup")
}

func GetGroupMember(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)

	groupid := c.Param("groupid")

	first := c.MustGet("first").(int)
	max := c.MustGet("max").(int)
	params := gocloak.GetGroupsParams{
		First: &first,
		Max:   &max,
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
