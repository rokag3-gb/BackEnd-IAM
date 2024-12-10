package api

import (
	"fmt"
	"iam/clients"
	"iam/common"
	"iam/config"
	"iam/iamdb"
	"iam/middlewares"
	"net/http"

	logger "cloudmt.co.kr/mateLogger"
	"github.com/Nerzal/gocloak/v11"
	"github.com/gin-gonic/gin"
)

type UserInviteRequest struct {
	Email     string `json:"email"`
	AccountID int64  `json:"accountId"`
}

type PostChangePasswordRequest struct {
	Password        string `json:"password" binding:"required,eqfield=PasswordConfirm"`
	PasswordConfirm string `json:"passwordConfirm" binding:"required"`
	Token           string `json:"token"`
}

// token godoc
// @Security Bearer
// @Summary User 기본 정보 추가
// @Tags User
// @Produce  json
// @Router /user-initialize [post]
// @Success 200 {object} []string
// @Failure 400
// @Failure 500
func UserInitializeKey(c *gin.Context) {
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

	arr, err := iamdb.SelectAccountKey(c.GetString("userId"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

// token godoc
// @Security Bearer
// @Summary 유저 초대
// @Tags User
// @Produce  json
// @Router /user-invite [post]
// @Param Body body api.UserInviteRequest true "body"
// @Success 200
// @Failure 500
func PostUserInvite(c *gin.Context) {
	var r UserInviteRequest
	err := c.ShouldBindJSON(&r)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	senderID := c.GetString("userId")
	realm := c.GetString("realm")
	accessToken := c.GetString("accessToken")

	tenant, err := iamdb.GetTenantIdByRealm(realm)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	db, err := iamdb.DBClient()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	conf := config.GetConfig()
	keycloakToken, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	targetUserID, err := iamdb.SelectUserByEmail(db, r.Email)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if targetUserID != "" {
		result, err := iamdb.SelectAccountUserByEmail(db, r.Email, r.AccountID)
		if err != nil {
			common.ErrorProcess(c, err, http.StatusInternalServerError, "")
			return
		}

		//유저가 이미 등록되어있고 AccountUser 또한 등록되어있다면 StatusConflict를 리턴한다.
		if result {
			common.ErrorProcess(c, err, http.StatusConflict, "")
			return
		}

		_, err = clients.SalesPostAccountUser(accessToken, realm, clients.PostAccountUser{AccountId: r.AccountID, UserId: targetUserID, IsUse: true})
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		c.Status(http.StatusOK)
		return
	}

	userID, err := clients.KeycloakClient().CreateUser(c,
		keycloakToken.AccessToken,
		realm,
		gocloak.User{
			Username: gocloak.StringP(r.Email),
			Email:    gocloak.StringP(r.Email),
			Enabled:  gocloak.BoolP(false),
		})
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	_, err = clients.SalesPostAccountUser(accessToken, realm, clients.PostAccountUser{AccountId: r.AccountID, UserId: userID, IsUse: true})
	if err != nil {
		err := DeleteUserData(userID, accessToken)
		if err != nil {
			logger.Error(err.Error())
		}
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	token, err := common.GetToken(senderID, tenant, userID, "TKT-PWD", []string{
		fmt.Sprintf("POST /%s/user/change-password", conf.MerlinDefaultURL),
	})
	if err != nil {
		err := DeleteUserData(userID, accessToken)
		if err != nil {
			logger.Error(err.Error())
		}
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	url := fmt.Sprintf(conf.UserInviteURL, token)

	_, err = clients.SalesSendEmail(accessToken, realm, clients.EmailRequest{
		Subject:    conf.UserInviteSubject,
		SenderName: conf.UserInviteSenderName,
		To:         []string{r.Email},
		Body:       fmt.Sprintf(conf.UserInviteMessage, url),
		IsBodyHtml: true,
	})
	if err != nil {
		err := DeleteUserData(userID, accessToken)
		if err != nil {
			logger.Error(err.Error())
		}
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusOK)
}

// token godoc
// @Security Bearer
// @Summary 유저 패스워드 변경 요청
// @Tags User
// @Produce  json
// @Router /user/{userID}/forgot-password [post]
// @Param userID path string true "User Id"
// @Success 200
// @Failure 500
func PostForgotPassword(c *gin.Context) {
	userID := c.Param("userid")

	senderID := c.GetString("userId")
	realm := c.GetString("realm")

	db, err := iamdb.DBClient()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	email, err := iamdb.SelectEmailByUser(db, userID)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	if email == "" {
		common.ErrorProcess(c, nil, http.StatusBadRequest, "user does not have an email address set")
		return
	}

	conf := config.GetConfig()
	clientData := conf.Keycloak_realm_client_secret[realm]
	if clientData.ClientID == "" || clientData.ClientSecret == "" {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "service config error")
		return
	}

	accessToken, err := clients.KeycloakRealmToken(c, clientData.ClientID, clientData.ClientSecret, realm)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = clients.DeleteUserCredential(c, accessToken.AccessToken, realm, userID)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = clients.SetUserDiabled(c, accessToken.AccessToken, realm, userID)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	tenant, err := iamdb.GetTenantIdByRealm(realm)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	token, err := common.GetToken(senderID, tenant, userID, "TKT-PWD", []string{
		fmt.Sprintf("POST /%s/user/change-password", conf.MerlinDefaultURL),
	})
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	url := fmt.Sprintf(conf.ChangePasswordURL, token)

	clients.SalesSendEmail(accessToken.AccessToken, realm, clients.EmailRequest{
		Subject:    conf.ChangePasswordSubject,
		SenderName: conf.ChangePasswordSenderName,
		To:         []string{email},
		Body:       fmt.Sprintf(conf.ChangePasswordMessage, url),
		IsBodyHtml: true,
	})

	c.Status(http.StatusOK)
}

// token godoc
// @Security Bearer
// @Summary 유저 패스워드 변경
// @Tags User
// @Produce  json
// @Router /user/change-password [post]
// @Param Body body api.PostChangePasswordRequest true "PostChangePasswordRequest"
// @Success 200
// @Failure 400
// @Failure 500
func PostChangePassword(c *gin.Context) {
	realm := c.GetString("realm")
	userid := c.GetString("userId")

	var data PostChangePasswordRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}
	getToken := c.GetString("accessToken")

	result, err := common.TokenIntrospect(getToken)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, err.Error())
		return
	}
	if !result {
		common.ErrorProcess(c, nil, http.StatusUnauthorized, "invalid token")
		return
	}

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

	user.Enabled = gocloak.BoolP(true)

	err = clients.KeycloakClient().UpdateUser(c,
		token.AccessToken,
		realm,
		*user,
	)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = clients.KeycloakClient().SetPassword(c,
		token.AccessToken,
		userid,
		realm,
		data.Password,
		false)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	tokenID, err := common.TokenParse(data.Token)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, err.Error())
		return
	}

	db, err := iamdb.DBClient()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.UpdateTokenConsume(db, tokenID)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusOK)
}
