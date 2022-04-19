package api

import (
	"encoding/json"
	"iam/clients"
	"iam/models"
	"net/http"

	"github.com/Nerzal/gocloak/v11"
	"github.com/gin-gonic/gin"
)

func Users(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)

	params := gocloak.GetUsersParams{
		First:               gocloak.IntP(c.MustGet("first").(int)),
		Max:                 gocloak.IntP(c.MustGet("max").(int)),
		BriefRepresentation: gocloak.BoolP(true),
	}

	users, err := clients.KeycloakClient().GetUsers(c,
		token.AccessToken,
		clients.KeycloakConfig().Realm,
		params)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	response := []gin.H{}
	for _, user := range users {
		var userMap gin.H
		inrec, _ := json.Marshal(user)
		json.Unmarshal(inrec, &userMap)
		delete(userMap, "access")
		response = append(response, userMap)
	}
	c.JSON(http.StatusOK, response)
}

func CreateUser(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	var json models.CreateUserInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	newUserId, err := clients.KeycloakClient().CreateUser(c,
		token.AccessToken,
		clients.KeycloakConfig().Realm,
		gocloak.User{
			Username:  gocloak.StringP(json.Username),
			FirstName: gocloak.StringP(json.FirstName),
			LastName:  gocloak.StringP(json.LastName),
			Email:     gocloak.StringP(json.Email),
		})
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	err = clients.KeycloakClient().SetPassword(c,
		token.AccessToken,
		newUserId,
		clients.KeycloakConfig().Realm,
		json.Password,
		false)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gocloak.User{ID: gocloak.StringP(newUserId)})
}

func UpdateUser(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	userid := c.Param("userid")

	user, err := clients.KeycloakClient().GetUserByID(c,
		token.AccessToken, clients.KeycloakConfig().Realm, userid)

	var json models.UpdateUserInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	user.Username = gocloak.StringP(json.Username)
	user.FirstName = gocloak.StringP(json.FirstName)
	user.LastName = gocloak.StringP(json.LastName)
	user.Email = gocloak.StringP(json.Email)
	user.RequiredActions = &json.RequiredActions

	err = clients.KeycloakClient().UpdateUser(c,
		token.AccessToken,
		clients.KeycloakConfig().Realm,
		*user)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func DeleteUser(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	userid := c.Param("userid")

	err := clients.KeycloakClient().DeleteUser(c,
		token.AccessToken,
		clients.KeycloakConfig().Realm,
		userid)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
