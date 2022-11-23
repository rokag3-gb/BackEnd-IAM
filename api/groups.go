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

func GetGroup(c *gin.Context) {
	arr, err := iamdb.GetGroup()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

func CreateGroup(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
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
		clients.KeycloakConfig().Realm,
		group)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.GroupCreate(newGroup, c.GetString("username"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, gocloak.Group{ID: gocloak.StringP(newGroup)})
}

func DeleteGroup(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	groupid := c.Param("groupid")

	err := clients.KeycloakClient().DeleteGroup(c, token.AccessToken, clients.KeycloakConfig().Realm, groupid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	c.Status(http.StatusNoContent)
}

func UpdateGroup(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	groupid := c.Param("groupid")
	var json models.GroupInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	groupToUpdate, err := clients.KeycloakClient().GetGroup(c, token.AccessToken, clients.KeycloakConfig().Realm, groupid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	groupToUpdate.Name = gocloak.StringP(json.Name)

	err = clients.KeycloakClient().UpdateGroup(c, token.AccessToken, clients.KeycloakConfig().Realm, *groupToUpdate)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.GroupUpdate(groupid, c.GetString("username"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}
