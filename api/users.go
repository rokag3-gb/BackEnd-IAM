package api

import (
	"iam/clients"
	"net/http"

	"github.com/Nerzal/gocloak/v11"
	"github.com/gin-gonic/gin"
)

func Users(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)

	params := gocloak.GetUsersParams{
		First: gocloak.IntP(c.MustGet("first").(int)),
		Max:   gocloak.IntP(c.MustGet("max").(int)),
		// BriefRepresentation: gocloak.BoolP(true),
	}

	users, err := clients.KeycloakClient().GetUsers(c,
		token.AccessToken,
		clients.KeycloakConfig().Realm,
		params)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, users)
}
