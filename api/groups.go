package api

import (
	"iam/clients"
	"net/http"

	"github.com/Nerzal/gocloak/v11"
	"github.com/gin-gonic/gin"
)

func GetGroup(c *gin.Context) {
	token := c.MustGet("accessToken").(string)

	groups, err := clients.KeycloakClient().GetGroups(c, token, clients.KeycloakRealm, gocloak.GetGroupsParams{})
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, groups)
	// c.String(http.StatusOK, "getgroup")
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
