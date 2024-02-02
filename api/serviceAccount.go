package api

import (
	"encoding/json"
	"iam/clients"
	"iam/common"
	"iam/iamdb"
	"iam/models"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// token godoc
// @Summary 서비스 어카운트 계정 조회
// @Tags ServiceAccount
// @Produce  json
// @Router /serviceAccount [get]
// @Success 200 {object} []models.GetServiceAccountInfo
// @Failure 500
func GetServiceAccount(c *gin.Context) {
	paramPairs := c.Request.URL.Query()
	var params = map[string][]string{}

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

	arr, err := iamdb.SelectServiceAccount(params)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, arr)
}

// token godoc
// @Summary 서비스 어카운트 시크릿 조회
// @Tags ServiceAccount
// @Produce json
// @Router /serviceAccount/{realm}/{clientId}/secret [get]
// @Param realm path string true "realm ID"
// @Param clientId path string true "client ID"
// @Success 200 {object} models.ClientSecret
// @Failure 500
func GetServiceAccountSecret(c *gin.Context) {
	clientId := c.Param("clientId")
	realm := c.Param("realm")

	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	secret, err := clients.GetServiceAccountSecret(c, token.AccessToken, realm, clientId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, secret)
}

// token godoc
// @Summary 서비스 어카운트 시크릿 재생성
// @Tags ServiceAccount
// @Produce json
// @Router /serviceAccount/{realm}/{clientId}/secret/regenerate [post]
// @Param clientId path string true "client ID"
// @Success 200 {object} models.ClientSecret
// @Failure 500
func RegenerateServiceAccountSecret(c *gin.Context) {
	clientId := c.Param("clientId")
	realm := c.Param("realm")

	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	secret, err := clients.RegenerateServiceAccountSecret(c, token.AccessToken, realm, clientId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, secret)
}

// token godoc
// @Summary 서비스 어카운트 생성
// @Tags ServiceAccount
// @Produce json
// @Router /{realm}/serviceAccount [post]
// @Param Body body models.CreateServiceAccount true "body"
// @Success 201
// @Failure 500
func CreateServiceAccount(c *gin.Context) {
	realm := c.Param("realm")
	value, err := io.ReadAll(c.Request.Body)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	var r *models.CreateServiceAccount
	json.Unmarshal([]byte(value), &r)
	if r == nil || r.ClientId == "" {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = clients.CreateServiceAccount(c, token.AccessToken, realm, r.ClientId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusCreated)
}

// token godoc
// @Summary 서비스 어카운트 제거
// @Tags ServiceAccount
// @Produce json
// @Router /serviceAccount/{realm}/{clientId} [delete]
// @Param clientId path string true "client ID"
// @Success 204
// @Failure 500
func DeleteServiceAccount(c *gin.Context) {
	clientId := c.Param("clientId")
	realm := c.Param("realm")

	token, err := clients.KeycloakToken(c)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = clients.DeleteServiceAccount(c, token.AccessToken, realm, clientId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}
