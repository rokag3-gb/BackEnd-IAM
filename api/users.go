package api

import (
	"iam/clients"
	"iam/config"
	"iam/iamdb"
	"iam/models"
	"net/http"

	logger "cloudmt.co.kr/mateLogger"
	"github.com/Nerzal/gocloak/v11"
	"github.com/gin-gonic/gin"
)

func Users(c *gin.Context) {
	search := c.Query("search")
	groupid := c.Query("groupid")

	arr, err := iamdb.GetUsers(search, groupid)
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, arr)
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
			Enabled:   gocloak.BoolP(true),
		})
	if err != nil {
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}
	err = clients.KeycloakClient().SetPassword(c,
		token.AccessToken,
		newUserId,
		clients.KeycloakConfig().Realm,
		json.Password,
		false)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}

	err = iamdb.UsersCreate(newUserId, c.GetString("username"))
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gocloak.User{ID: gocloak.StringP(newUserId)})
}

func UpdateUser(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	userid := c.Param("userid")

	user, err := clients.KeycloakClient().GetUserByID(c,
		token.AccessToken, clients.KeycloakConfig().Realm, userid)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	var json models.UpdateUserInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	user.Username = gocloak.StringP(json.Username)
	user.FirstName = gocloak.StringP(json.FirstName)
	user.LastName = gocloak.StringP(json.LastName)
	user.Email = gocloak.StringP(json.Email)
	user.Enabled = gocloak.BoolP(json.Enabled)
	user.RequiredActions = &json.RequiredActions

	err = clients.KeycloakClient().UpdateUser(c,
		token.AccessToken,
		clients.KeycloakConfig().Realm,
		*user)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}

	err = iamdb.UsersUpdate(userid, c.GetString("username"))
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
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
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}

	err = iamdb.DeleteUserRoleByUserId(userid)
	if err != nil {

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}

func GetUser(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	userid := c.Param("userid")

	user, err := clients.KeycloakClient().GetUserByID(c,
		token.AccessToken, clients.KeycloakConfig().Realm, userid)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}

	c.JSON(http.StatusOK, models.GetUserInfo{
		ID:               user.ID,
		CreatedTimestamp: user.CreatedTimestamp,
		Username:         user.Username,
		Enabled:          user.Enabled,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		Email:            user.Email,
		RequiredActions:  user.RequiredActions,
	})
}

func GetUserCredentials(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	userid := c.Param("userid")

	credentials, err := clients.KeycloakClient().GetCredentials(c,
		token.AccessToken, clients.KeycloakConfig().Realm, userid)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}

	c.JSON(http.StatusOK, credentials)
}

func ResetUserPassword(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	userid := c.Param("userid")

	var json models.ResetUserPasswordInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	err := clients.KeycloakClient().SetPassword(c,
		token.AccessToken,
		userid,
		clients.KeycloakConfig().Realm,
		json.Password,
		json.Temporary)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}

	c.Status(http.StatusNoContent)
}

func GetUserGroups(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	userid := c.Param("userid")

	groups, err := clients.KeycloakClient().GetUserGroups(c,
		token.AccessToken, clients.KeycloakConfig().Realm, userid,
		gocloak.GetGroupsParams{
			First: gocloak.IntP(c.MustGet("first").(int)),
			Max:   gocloak.IntP(c.MustGet("max").(int)),
		})
	if err != nil {
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}

	c.JSON(http.StatusOK, groups)
}

func AddUserToGroup(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	userid := c.Param("userid")
	groupid := c.Param("groupid")

	err := clients.KeycloakClient().AddUserToGroup(c,
		token.AccessToken,
		clients.KeycloakConfig().Realm,
		userid,
		groupid)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}

	c.Status(http.StatusNoContent)
}

func DeleteUserFromGroup(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	userid := c.Param("userid")
	groupid := c.Param("groupid")

	err := clients.KeycloakClient().DeleteUserFromGroup(c,
		token.AccessToken,
		clients.KeycloakConfig().Realm,
		userid,
		groupid)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}

	c.Status(http.StatusNoContent)
}

func GetUserSessions(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	userid := c.Param("userid")

	sessions, err := clients.KeycloakClient().GetUserSessions(c,
		token.AccessToken, clients.KeycloakConfig().Realm, userid)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}

	c.JSON(http.StatusOK, sessions)
}

func LogoutUserSession(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	userid := c.Param("userid")
	sessionid := c.Param("sessionid")

	sessions, err := clients.KeycloakClient().GetUserSessions(c,
		token.AccessToken, clients.KeycloakConfig().Realm, userid)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}

	hasSession := false
	for _, session := range sessions {
		if *session.ID == sessionid {
			hasSession = true
			break
		}
	}

	if !hasSession {
		c.String(http.StatusBadRequest, "Session not found from user")
		return
	}

	err = clients.KeycloakClient().LogoutUserSession(c,
		token.AccessToken,
		clients.KeycloakConfig().Realm,
		sessionid)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}

	c.Status(http.StatusNoContent)
}

func LogoutAllSessions(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	userid := c.Param("userid")

	err := clients.KeycloakClient().LogoutAllSessions(c,
		token.AccessToken,
		clients.KeycloakConfig().Realm,
		userid)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}

	c.Status(http.StatusNoContent)
}

func GetUserFederatedIdentities(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	userid := c.Param("userid")

	identities, err := clients.KeycloakClient().GetUserFederatedIdentities(c,
		token.AccessToken, clients.KeycloakConfig().Realm, userid)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		c.String(apiError.Code, apiError.Message)
		return
	}

	c.JSON(http.StatusOK, identities)
}
