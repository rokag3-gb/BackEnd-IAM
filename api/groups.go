package api

import (
	"iam/clients"
	"net/http"
	"strconv"

	"github.com/Nerzal/gocloak/v11"
	"github.com/gin-gonic/gin"
)

func GetGroup(c *gin.Context) {
	token := c.MustGet("accessToken").(string)
	first, firstErr := strconv.Atoi(c.DefaultQuery("first", "0"))
	if firstErr != nil {
		c.String(http.StatusBadRequest, "'first' must be integer")
		c.Abort()
		return
	}
	max, maxErr := strconv.Atoi(c.DefaultQuery("max", "100"))
	if maxErr != nil {
		c.String(http.StatusBadRequest, "'max' must be integer")
		c.Abort()
		return
	}

	params := gocloak.GetGroupsParams{
		First: &first,
		Max:   &max,
	}

	groups, err := clients.KeycloakClient().GetGroups(c, token, clients.KeycloakRealm, params)
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
	c.String(http.StatusOK, "getgroupmember")
}
