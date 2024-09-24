package api

import (
	"encoding/json"
	"fmt"
	"iam/clients"
	"iam/common"
	"iam/iamdb"
	"iam/models"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// token godoc
// @Summary 서비스 어카운트 계정 조회
// @Tags ServiceAccount
// @Produce  json
// @Router /serviceAccount [get]
// @Success 200 {object} []models.GetServiceAccount
// @Failure 500
func GetServiceAccounts(c *gin.Context) {
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

	db, err := iamdb.DBClient()
	defer db.Close()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	arr, idArr, err := iamdb.SelectServiceAccounts(db)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	info, err := iamdb.SelectClientsAttribute(db, idArr)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	for i, v := range arr {
		createDate := info[*v.ClientId+"_createDate"]
		creator := info[*v.ClientId+"_creator"]
		modifyDate := info[*v.ClientId+"_modifyDate"]
		modifier := info[*v.ClientId+"_modifier"]

		arr[i].CreateDate = createDate
		arr[i].Creator = creator
		arr[i].ModifyDate = modifyDate
		arr[i].Modifier = modifier
	}

	c.JSON(http.StatusOK, arr)
}

// token godoc
// @Summary 유저 상세정보 조회
// @Tags Users
// @Produce  json
// @Router /users/{userId} [get]
// @Param userId path string true "User Id"
// @Success 200 {object} models.GetUserInfo
// @Failure 500
func GetServiceAccount(c *gin.Context) {
	db, err := iamdb.DBClient()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, err.Error())
	}
	defer db.Close()

	userid := c.Param("id")
	result, err := iamdb.SelectServiceAccount(db, userid)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	data, err := iamdb.SelectClientAttribute(db, *result.ClientId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	result.CreateDate = data["createDate"]
	result.Creator = data["creator"]
	result.ModifyDate = data["modifyDate"]
	result.Modifier = data["modifier"]

	c.JSON(http.StatusOK, result)
}

// token godoc
// @Summary 서비스 어카운트 시크릿 조회
// @Tags ServiceAccount
// @Produce json
// @Router /serviceAccount/{userId}/secret [get]
// @Param userId path string true "User Id"
// @Success 200 {object} models.ClientSecret
// @Failure 500
func GetServiceAccountSecret(c *gin.Context) {
	userId := c.Param("id")

	db, err := iamdb.DBClient()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, err.Error())
		return
	}
	defer db.Close()

	clientId, err := iamdb.SelectClientIdFromUserId(db, userId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, err.Error())
		return
	}

	realm := c.Query("realm")
	if realm == "" {
		realm = c.GetString("realm")
	}

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
// @Router /serviceAccount/{userId}/secret/regenerate [post]
// @Param userId path string true "User Id"
// @Param realm query string true "realm ID"
// @Success 200 {object} models.ClientSecret
// @Failure 500
func RegenerateServiceAccountSecret(c *gin.Context) {
	userId := c.Param("id")

	db, err := iamdb.DBClient()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, err.Error())
		return
	}
	defer db.Close()

	clientId, err := iamdb.SelectClientIdFromUserId(db, userId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, err.Error())
		return
	}

	realm := c.Query("realm")
	if realm == "" {
		realm = c.GetString("realm")
	}

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

	idOfClient, err := iamdb.SelectIdFromClientId(db, clientId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, err.Error())
		return
	}

	curTime := time.Now().Format("2006-01-02,15:04:05")

	err = iamdb.InsertUpdateClientAttribute(db, idOfClient, "modifyDate", curTime)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	err = iamdb.InsertUpdateClientAttribute(db, idOfClient, "modifier", userId)
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
// @Router /serviceAccount [post]
// @Param Body body models.CreateServiceAccount true "body"
// @Success 201
// @Failure 500
func CreateServiceAccount(c *gin.Context) {
	realm := c.Query("realm")
	if realm == "" {
		realm = c.GetString("realm")
	}
	if realm == "" {
		common.ErrorProcess(c, fmt.Errorf("required 'realm'"), http.StatusBadRequest, "")
	}

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

	err, idOfClient := clients.CreateServiceAccount(c, token.AccessToken, realm, r.ClientId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	db, dbErr := iamdb.DBClient()
	if dbErr != nil {
		common.ErrorProcess(c, dbErr, http.StatusInternalServerError, dbErr.Error())
	}
	defer db.Close()

	curTime := time.Now().Format("2006-01-02,15:04:05")
	userId := c.GetString("userId")
	{
		err = iamdb.InsertUpdateClientAttribute(db, idOfClient, "createDate", curTime)
		if err != nil {
			common.ErrorProcess(c, err, http.StatusInternalServerError, "")
			return
		}
		err = iamdb.InsertUpdateClientAttribute(db, idOfClient, "creator", userId)
		if err != nil {
			common.ErrorProcess(c, err, http.StatusInternalServerError, "")
			return
		}
		err = iamdb.InsertUpdateClientAttribute(db, idOfClient, "modifyDate", curTime)
		if err != nil {
			common.ErrorProcess(c, err, http.StatusInternalServerError, "")
			return
		}
		err = iamdb.InsertUpdateClientAttribute(db, idOfClient, "modifier", userId)
		if err != nil {
			common.ErrorProcess(c, err, http.StatusInternalServerError, "")
			return
		}
	}

	c.Status(http.StatusCreated)
}

// token godoc
// @Summary 서비스 어카운트 제거
// @Tags ServiceAccount
// @Produce json
// @Router /serviceAccount/{userId} [delete]
// @Param userId path string true "User Id"
// @Success 204
// @Failure 500
func DeleteServiceAccount(c *gin.Context) {
	userId := c.Param("id")

	db, err := iamdb.DBClient()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, err.Error())
		return
	}
	defer db.Close()

	clientId, err := iamdb.SelectClientIdFromUserId(db, userId)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, err.Error())
		return
	}

	realm := c.Query("realm")
	if realm == "" {
		realm = c.GetString("realm")
	}

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

	//키클락에서 어트리뷰트를 삭제하므로 삭제할 필요 없음

	c.Status(http.StatusNoContent)
}
