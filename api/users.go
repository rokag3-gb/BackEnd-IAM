package api

import (
	"context"
	"fmt"
	"iam/clients"
	"iam/common"
	"iam/iamdb"
	"iam/middlewares"
	"iam/models"
	"net/http"
	"strings"

	"github.com/Nerzal/gocloak/v11"
	"github.com/gin-gonic/gin"

	logger "cloudmt.co.kr/mateLogger"
)

// Roles, Groups, Account, AccountId, openid 는 하나의 유저가 여러개의 값을 가질 수 있어 콤마로 구분
var SearchUsers = map[string]string{
	"search":    "U.USERNAME",
	"username":  "U.USERNAME",
	"groupid":   "B.GROUP_ID",
	"groups":    "B.Groups",
	"roles":     "A.Roles",
	"enabled":   "U.ENABLED",
	"email":     "U.EMAIL",
	"openid":    "C.openid",
	"accounts":  "D.Account",
	"accountid": "D.AccountId",
	"ids":       "U.ID",
}

// token godoc
// @Security Bearer
// @Summary Account 유저 목록
// @Tags Account
// @Produce  json
// @Router /account/{accountId}/users [get]
// @Success 200 {object} []models.GetUserInfo
// @Failure 500

// token godoc
// @Security Bearer
// @Summary 유저 목록
// @Tags Users
// @Produce  json
// @Router /users [get]
// @Success 200 {object} []models.GetUserInfo
// @Failure 500
func Users(c *gin.Context) {
	paramPairs := c.Request.URL.Query()
	var params = map[string][]string{}

	if c.Param("accountId") != "" {
		paramPairs.Add("accountId", c.Param("accountId"))
	}

	for key, values := range paramPairs {
		col := SearchUsers[strings.ToLower(key)]
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

// token godoc
// @Security Bearer
// @Summary Account 유저 생성
// @Tags Account
// @Produce  json
// @Router /account/{accountId}/users [post]
// @Param Body body models.CreateUserInfo true "body"
// @Success 200 {object} models.Id
// @Failure 400
// @Failure 500

// token godoc
// @Security Bearer
// @Summary 유저 생성
// @Tags Users
// @Produce  json
// @Router /users [post]
// @Param Body body models.CreateUserInfo true "body"
// @Success 200 {object} models.Id
// @Failure 400
// @Failure 500
func CreateUser(c *gin.Context) {
	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	var json models.CreateUserInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}
	if json.Realm == "" {
		json.Realm = c.GetString("realm")
	}

	newUserId, err := clients.KeycloakClient().CreateUser(c,
		token.AccessToken,
		json.Realm,
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
		json.Realm,
		json.Password,
		false)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.UsersCreate(newUserId, c.GetString("userId"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if c.Param("accountId") != "" {
		err = iamdb.InsertAccountUser(c.Param("accountId"), newUserId, c.GetString("userId"))
		if err != nil {
			common.ErrorProcess(c, err, http.StatusInternalServerError, "")
			return
		}
	}

	err = iamdb.CreateUserAddDefaultRole(newUserId, json.Realm, c.GetString("userId"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, gocloak.User{ID: gocloak.StringP(newUserId)})
}

// token godoc
// @Security Bearer
// @Summary Account 유저 정보 변경
// @Tags Account
// @Produce  json
// @Router /account/{accountId}/users/{userId} [put]
// @Param accountId path string true "account Id"
// @Param userId path string true "User Id"
// @Param Body body models.CreateUserInfo true "body"
// @Success 204
// @Failure 400
// @Failure 500

// token godoc
// @Security Bearer
// @Summary 유저 정보 변경
// @Tags Users
// @Produce  json
// @Router /users/{userId} [put]
// @Param userId path string true "User Id"
// @Param Body body models.CreateUserInfo true "body"
// @Success 204
// @Failure 400
// @Failure 500
func UpdateUser(c *gin.Context) {
	realm := c.GetString("realm")
	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	userid := c.Param("userid")

	user, err := clients.KeycloakClient().GetUserByID(c,
		token.AccessToken, realm, userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	var json models.UpdateUserInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if json.Username != nil {
		user.Username = gocloak.StringP(*json.Username)
	}

	if json.FirstName != nil {
		user.FirstName = gocloak.StringP(*json.FirstName)
	}

	if json.LastName != nil {
		user.LastName = gocloak.StringP(*json.LastName)
	}

	if json.Email != nil {
		user.Email = gocloak.StringP(*json.Email)
	}

	if json.Enabled != nil {
		user.Enabled = gocloak.BoolP(*json.Enabled)
	}

	if json.RequiredActions != nil {
		user.RequiredActions = json.RequiredActions
	}

	if json.Attributes != nil {
		user.Attributes = json.Attributes
	}

	phoneNumber := ""
	if json.PhoneNumber != nil {
		phoneNumber = *gocloak.StringP(*json.PhoneNumber)
	}

	err = clients.KeycloakClient().UpdateUser(c,
		token.AccessToken,
		realm,
		*user)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.UsersUpdate(userid, phoneNumber, c.GetString("userId"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Security Bearer
// @Summary Account 자신의 계정 정보 변경
// @Tags Account
// @Produce  json
// @Router /account/{accountId}/users/me [put]
// @Param accountId path string true "account Id"
// @Param Body body models.UpdateUserInfo true "body"
// @Success 204
// @Failure 400
// @Failure 500

// token godoc
// @Security Bearer
// @Summary 자신의 계정 정보 변경
// @Tags Users
// @Produce  json
// @Router /users/me [post]
// @Param Body body models.UpdateUserInfo true "body"
// @Success 204
// @Failure 400
// @Failure 500
func UpdateMe(c *gin.Context) {
	realm := c.GetString("realm")
	userid := c.GetString("userId")

	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	user, err := clients.KeycloakClient().GetUserByID(c,
		token.AccessToken, realm, userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	var json models.UpdateUserInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if json.Username != nil {
		user.Username = gocloak.StringP(*json.Username)
	}

	if json.FirstName != nil {
		user.FirstName = gocloak.StringP(*json.FirstName)
	}

	if json.LastName != nil {
		user.LastName = gocloak.StringP(*json.LastName)
	}

	if json.Email != nil {
		user.Email = gocloak.StringP(*json.Email)
	}

	if json.Enabled != nil {
		user.Enabled = gocloak.BoolP(*json.Enabled)
	}

	if json.RequiredActions != nil {
		user.RequiredActions = json.RequiredActions
	}

	if json.Attributes != nil {
		user.Attributes = json.Attributes
	}

	phoneNumber := ""
	if json.PhoneNumber != nil {
		phoneNumber = *gocloak.StringP(*json.PhoneNumber)
	}

	err = clients.KeycloakClient().UpdateUser(c,
		token.AccessToken,
		realm,
		*user,
	)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.UsersUpdate(userid, phoneNumber, userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Security Bearer
// @Summary 유저 삭제
// @Tags Users
// @Produce  json
// @Router /users/{userId} [delete]
// @Param userId path string true "User Id"
// @Success 204
// @Failure 400
// @Failure 500
func DeleteUser(c *gin.Context) {
	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	userID := c.Param("userid")

	err = DeleteUserData(userID, token.AccessToken)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

func DeleteUserData(userID, token string) error {
	realm, err := iamdb.GetUserRealmById(userID)
	if err != nil {
		return err
	}

	arr, err := iamdb.GetAccountUserId(userID)
	if err != nil {
		logger.Error(err.Error())
	} else {
		for _, seq := range arr {
			if seq == "" {
				continue
			}

			str, err := clients.SalesDeleteAccountUser(seq, realm, token)
			fmt.Println(str)
			if err != nil {
				logger.Error("%s", err.Error())
			}
		}
	}

	ctx := context.Background()
	err = clients.KeycloakClient().DeleteUser(
		ctx,
		token,
		realm,
		userID)
	if err != nil {
		return err
	}

	err = iamdb.DeleteUserRoleByUserId(userID)
	if err != nil {
		logger.Error(err.Error())
	}

	return nil
}

// token godoc
// @Security Bearer
// @Summary Account 유저 상세정보 조회
// @Tags Account
// @Produce  json
// @Router /account/{accountId}/users/{userId} [get]
// @Param accountId path string true "account Id"
// @Param userId path string true "User Id"
// @Success 200 {object} models.GetUserInfo
// @Failure 500

// token godoc
// @Security Bearer
// @Summary 유저 상세정보 조회
// @Tags Users
// @Produce  json
// @Router /users/{userId} [get]
// @Param userId path string true "User Id"
// @Success 200 {object} models.GetUserInfo
// @Failure 500
func GetUser(c *gin.Context) {
	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	userid := c.Param("userid")
	realm, err := iamdb.GetUserRealmById(userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	user, err := clients.KeycloakClient().GetUserByID(c,
		token.AccessToken, realm, userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	var params = map[string][]string{}
	params["U.ID"] = append(params["U.ID"], *user.ID)
	arr, err := iamdb.GetUsers(params)
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
		PhoneNumber:      arr[0].PhoneNumber,
		RequiredActions:  user.RequiredActions,
		Attributes:       user.Attributes,
	})
}

// token godoc
// @Security Bearer
// @Summary Account 유저 자격증명 조회
// @Tags Account
// @Produce  json
// @Router /account/{accountId}/users/{userId}/credentials [get]
// @Param accountId path string true "account Id"
// @Param userId path string true "User Id"
// @Success 200 {object} []models.CredentialRepresentation
// @Failure 500

// token godoc
// @Security Bearer
// @Summary 유저 자격증명 조회
// @Tags Users
// @Produce  json
// @Router /users/{userId}/credentials [get]
// @Param userId path string true "User Id"
// @Success 200 {object} []models.CredentialRepresentation
// @Failure 500
func GetUserCredentials(c *gin.Context) {
	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	userid := c.Param("userid")

	realm, err := iamdb.GetUserRealmById(userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	credentials, err := clients.KeycloakClient().GetCredentials(c,
		token.AccessToken, realm, userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, credentials)
}

// token godoc
// @Security Bearer
// @Summary Account 유저 비밀번호 변경
// @Tags Account
// @Produce  json
// @Router /account/{accountId}/users/{userId}/reset-password [put]
// @Param accountId path string true "account Id"
// @Param userId path string true "User Id"
// @Param roleId body models.ResetUserPasswordInfo true "ResetUserPasswordInfo"
// @Success 204
// @Failure 500

// token godoc
// @Security Bearer
// @Summary 유저 비밀번호 변경
// @Tags Users
// @Produce  json
// @Router /users/{userId}/reset-password [put]
// @Param userId path string true "User Id"
// @Param roleId body models.ResetUserPasswordInfo true "ResetUserPasswordInfo"
// @Success 204
// @Failure 500
func ResetUserPassword(c *gin.Context) {
	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	userid := c.Param("userid")

	var json models.ResetUserPasswordInfo
	if err := c.ShouldBindJSON(&json); err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	realm, err := iamdb.GetUserRealmById(userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = clients.KeycloakClient().SetPassword(c,
		token.AccessToken,
		userid,
		realm,
		json.Password,
		json.Temporary)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Security Bearer
// @Summary 유저 그룹 조회
// @Tags Users
// @Produce  json
// @Router /users/{userId}/groups [get]
// @Param userId path string true "User Id"
// @Param first query int true "data start count"
// @Param max query int true "data max count"
// @Success 200 {object} []models.GroupData
// @Failure 500
func GetUserGroups(c *gin.Context) {
	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	userid := c.Param("userid")

	realm, err := iamdb.GetUserRealmById(userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	groups, err := clients.KeycloakClient().GetUserGroups(c,
		token.AccessToken, realm, userid,
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

// token godoc
// @Security Bearer
// @Summary 유저 그룹 가입
// @Tags Users
// @Produce  json
// @Router /users/{userId}/groups [put]
// @Param userId path string true "User Id"
// @Param groupId path string true "Group Id"
// @Success 204
// @Failure 500
func AddUserToGroup(c *gin.Context) {
	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	userid := c.Param("userid")
	groupid := c.Param("groupid")

	realm, err := iamdb.GetUserRealmById(userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = clients.KeycloakClient().AddUserToGroup(c,
		token.AccessToken,
		realm,
		userid,
		groupid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Security Bearer
// @Summary 유저 그룹 탈퇴
// @Tags Users
// @Produce  json
// @Router /users/{userId}/groups [delete]
// @Param userId path string true "User Id"
// @Param groupId path string true "Group Id"
// @Success 204
// @Failure 500
func DeleteUserFromGroup(c *gin.Context) {
	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	userid := c.Param("userid")
	groupid := c.Param("groupid")

	realm, err := iamdb.GetUserRealmById(userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = clients.KeycloakClient().DeleteUserFromGroup(c,
		token.AccessToken,
		realm,
		userid,
		groupid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Security Bearer
// @Summary 유저 세션 조회
// @Tags Users
// @Produce  json
// @Router /users/{userId}/sessions [get]
// @Param userId path string true "User Id"
// @Success 200 {object} []models.SesstionData
// @Failure 500
func GetUserSessions(c *gin.Context) {
	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	userid := c.Param("userid")

	realm, err := iamdb.GetUserRealmById(userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	sessions, err := clients.KeycloakClient().GetUserSessions(c,
		token.AccessToken, realm, userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// token godoc
// @Security Bearer
// @Summary 유저 세션 제거
// @Tags Users
// @Produce  json
// @Router /users/{userId}/sessions/{sessionId} [delete]
// @Param userId path string true "User Id"
// @Param sessionId path string true "Session Id"
// @Success 200 {object} []models.SesstionData
// @Failure 500
func LogoutUserSession(c *gin.Context) {
	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	userid := c.Param("userid")
	sessionid := c.Param("sessionid")

	realm, err := iamdb.GetUserRealmById(userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	sessions, err := clients.KeycloakClient().GetUserSessions(c,
		token.AccessToken, realm, userid)
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
		realm,
		sessionid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Security Bearer
// @Summary 유저 전체 세션 제거
// @Tags Users
// @Produce  json
// @Router /users/{userId}/logout [post]
// @Param userId path string true "User Id"
// @Success 204
// @Failure 500
func LogoutAllSessions(c *gin.Context) {
	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	userid := c.Param("userid")

	realm, err := iamdb.GetUserRealmById(userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = clients.KeycloakClient().LogoutAllSessions(c,
		token.AccessToken,
		realm,
		userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Security Bearer
// @Summary Account 유저 ID 제공자 조회
// @Tags Account
// @Produce  json
// @Router /account/{accountId}/users/{userId}/federated-identity [get]
// @Param accountId path string true "account Id"
// @Param userId path string true "User Id"
// @Success 200 {object} models.UserIdProviderData
// @Failure 500

// token godoc
// @Security Bearer
// @Summary 유저 ID 제공자 조회
// @Tags Users
// @Produce  json
// @Router /users/{userId}/federated-identity [get]
// @Param userId path string true "User Id"
// @Success 200 {object} models.UserIdProviderData
// @Failure 500
func GetUserFederatedIdentities(c *gin.Context) {
	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	userid := c.Param("userid")

	realm, err := iamdb.GetUserRealmById(userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	identities, err := clients.KeycloakClient().GetUserFederatedIdentities(c,
		token.AccessToken, realm, userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, identities)
}

// token godoc
// @Security Bearer
// @Summary Account 유저 ID 제공자 제거
// @Tags Account
// @Produce  json
// @Router /account/{accountId}/users/{userId}/federated-identity/{providerId} [delete]
// @Param accountId path string true "account Id"
// @Param userId path string true "User Id"
// @Param providerId path string true "Provider Id"
// @Success 204
// @Failure 500

// token godoc
// @Security Bearer
// @Summary 유저 ID 제공자 제거
// @Tags Users
// @Produce  json
// @Router /users/{userId}/federated-identity/{providerId} [delete]
// @Param userId path string true "User Id"
// @Param providerId path string true "Provider Id"
// @Success 204
// @Failure 500
func DeleteUserFederatedIdentity(c *gin.Context) {
	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	userid := c.Param("userid")
	providerId := c.Param("providerId")

	realm, err := iamdb.GetUserRealmById(userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = clients.KeycloakClient().DeleteUserFederatedIdentity(c, token.AccessToken, realm, userid, providerId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Security Bearer
// @Summary Users 유저 초기 설정 작업(대상이 자기 자신인 경우에만)
// @Tags Users
// @Produce  json
// @Router /users/initialize [get]
// @Success 200 {object} []int
// @Failure 400
// @Failure 500

// token godoc
// @Security Bearer
// @Summary Users 유저 초기 설정 작업(대상이 자기 자신인 경우에만)
// @Tags Users
// @Produce  json
// @Router /users/initialize [post]
// @Success 200 {object} []int
// @Failure 400
// @Failure 500
func UserInitialize(c *gin.Context) {
	token := c.GetString("accessToken")
	realm := c.GetString("realm")
	tenantId := c.Param("tenantId")

	if tenantId == "" {
		tenant, err := iamdb.GetTenantIdByRealm(realm)
		if err != nil {
			common.ErrorProcess(c, err, http.StatusInternalServerError, "")
			return
		}

		tenantId = tenant
	}

	email, client_id, err := middlewares.GetInitInfo(token)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}
	accountIds, err := iamdb.SelectDefaultAccount(email, c.GetString("userId"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	for _, id := range accountIds {
		err := iamdb.InsertAccountUser(fmt.Sprintf("%d", id), c.GetString("userId"), c.GetString("userId"))
		if err != nil {
			common.ErrorProcess(c, err, http.StatusInternalServerError, "")
			return
		}
	}

	result, err := iamdb.SelectAccount(email, c.GetString("userId"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	if result {
		roleIdList, err := iamdb.SelectNotExsistRole(client_id, c.GetString("userId"), realm)
		if err != nil {
			common.ErrorProcess(c, err, http.StatusInternalServerError, "")
			return
		}

		for _, roleId := range roleIdList {
			err = iamdb.AssignUserRole(c.GetString("userId"), tenantId, roleId, c.GetString("userId"))
			if err != nil {
				common.ErrorProcess(c, err, http.StatusInternalServerError, "")
				return
			}
		}
	}

	arr, err := iamdb.SelectAccountId(c.GetString("userId"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}
