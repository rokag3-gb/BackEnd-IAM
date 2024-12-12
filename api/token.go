package api

import (
	"iam/common"
	"iam/iamdb"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TokenRequest struct {
	Sub   string   `json:"sub"`
	Scope []string `json:"scope"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type TokenIntrospectRequest struct {
	Token string `json:"token"`
}

type TokenIntrospectResponse struct {
	Active bool `json:"active"`
}

// token godoc
// @Security Bearer
// @Summary Token 토큰 발급
// @Tags Token
// @Produce  json
// @Router /token [post]
// @Param Body body api.TokenRequest true "body"
// @Success 200 {object} api.TokenResponse
// @Failure 500
func PostToken(c *gin.Context) {
	var body TokenRequest
	err := c.ShouldBindJSON(&body)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, err.Error())
		return
	}

	senderID := c.GetString("userId")

	tenant, err := iamdb.GetTenantIdByRealm(c.GetString("realm"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, err.Error())
		return
	}

	if body.Scope == nil {
		body.Scope = []string{}
	}

	token, err := common.GetToken(senderID, tenant, body.Sub, "TKT-COM", "", body.Scope)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, TokenResponse{Token: token})
}

// token godoc
// @Security Bearer
// @Summary Token 토큰 검증
// @Tags Token
// @Produce  json
// @Router /token/introspect [post]
// @Param Body body api.TokenIntrospectRequest true "body"
// @Success 200 {object} api.TokenIntrospectResponse
// @Failure 401
// @Failure 500
func TokenIntrospect(c *gin.Context) {
	var body TokenIntrospectRequest
	err := c.ShouldBindJSON(&body)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, err.Error())
		return
	}

	result, err := common.TokenIntrospect(body.Token)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, err.Error())
		return
	}

	if !result {
		c.Status(http.StatusUnauthorized)
		return
	}

	c.JSON(http.StatusOK, TokenIntrospectResponse{Active: result})
}

// token godoc
// @Security Bearer
// @Summary Token 토큰 만료 요청
// @Tags Token
// @Produce  json
// @Router /token/consume [post]
// @Param Body body api.TokenIntrospectRequest true "body"
// @Success 200
// @Failure 500
func ConsumeToken(c *gin.Context) {
	var body TokenIntrospectRequest
	err := c.ShouldBindJSON(&body)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, err.Error())
		return
	}

	tokenID, err := common.TokenParse(body.Token)
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
