package api

import (
	"iam/clients"
	"iam/common"
	"iam/iamdb"
	"iam/models"
	"net/http"
	"strings"

	"github.com/Nerzal/gocloak/v11"
	"github.com/gin-gonic/gin"
)

var SearchUsers = map[string]string{
	"search":   "U.USERNAME",
	"username": "U.USERNAME",
	"groupid":  "UG.GROUP_ID",
	"groups":   "B.Groups",
	"roles":    "A.Roles",
	"enabled":  "U.ENABLED",
	"email":    "U.EMAIL",
	"openid":   "C.openid",
	"accounts": "AC.AccountName",
	"ids":      "U.ID",
}

func Users(c *gin.Context) {
	paramPairs := c.Request.URL.Query()
	var params = map[string][]string{}

	for key, values := range paramPairs {
		col := SearchUsers[key]
		if col == "" {
			continue
		}

		for _, val := range values {
			q := strings.Split(val, ",")
			if len(q) == 0 || q[0] == "" {
				continue
			}

			params[col] = append(params[col], q...)
		}
	}

	arr, err := iamdb.GetUsers(params)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

func CreateUser(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	var json models.CreateUserInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
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
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	err = clients.KeycloakClient().SetPassword(c,
		token.AccessToken,
		newUserId,
		clients.KeycloakConfig().Realm,
		json.Password,
		false)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.UsersCreate(newUserId, c.GetString("username"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.CreateUserAddRole(newUserId, c.GetString("username"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
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
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	var json models.UpdateUserInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
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
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.UsersUpdate(userid, c.GetString("username"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
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
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.DeleteUserRoleByUserId(userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
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
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
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
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, credentials)
}

func ResetUserPassword(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	userid := c.Param("userid")

	var json models.ResetUserPasswordInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err := clients.KeycloakClient().SetPassword(c,
		token.AccessToken,
		userid,
		clients.KeycloakConfig().Realm,
		json.Password,
		json.Temporary)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
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
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
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
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
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
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
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
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
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
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
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
		common.ErrorProcess(c, nil, http.StatusBadRequest, "Session not found from user")
		return
	}

	err = clients.KeycloakClient().LogoutUserSession(c,
		token.AccessToken,
		clients.KeycloakConfig().Realm,
		sessionid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
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
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
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
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, identities)
}

func DeleteUserFederatedIdentity(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)
	userid := c.Param("userid")
	providerId := c.Param("providerId")

	err := clients.KeycloakClient().DeleteUserFederatedIdentity(c, token.AccessToken, clients.KeycloakConfig().Realm, userid, providerId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}
