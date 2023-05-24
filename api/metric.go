package api

import (
	"encoding/json"
	"fmt"
	"iam/clients"
	"iam/common"
	"iam/config"
	"iam/iamdb"
	"iam/models"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// token godoc
// @Summary 리소스 수 조회
// @Tags Metric
// @Produce  json
// @Router /metric/count [get]
// @Success 200 {object} models.MetricCount
// @Failure 500
func MetricCount(c *gin.Context) {
	count, err := iamdb.MetricCount()

	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, count)
}

// token godoc
// @Summary 어플리케이션 별 현재 세션 수 조회
// @Tags Metric
// @Produce  json
// @Router /metric/session [get]
// @Success 200 {object} []models.MetricItem
// @Failure 500
func GetMetricSession(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)

	url := fmt.Sprintf("%s/admin/realms/%s/client-session-stats", config.GetConfig().Keycloak_endpoint, config.GetConfig().Keycloak_realm)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	req.Header.Add("authorization", "Bearer "+token.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)

	arr := make([]map[string]interface{}, 0)

	json.Unmarshal(bytes, &arr)

	apps, err := iamdb.GetApplications()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	ret := make([]models.MetricItem, 0)
	for _, app := range apps {
		var m models.MetricItem
		m.Key = app
		m.Value = 0
		ret = append(ret, m)
	}

	for i, app := range ret {
		for _, ar := range arr {
			if ar["clientId"] == app.Key {
				v, err := strconv.Atoi(ar["active"].(string))
				if err != nil {
					break
				}
				ret[i].Value = v
			}
		}
	}

	c.JSON(http.StatusOK, ret)
}

// token godoc
// @Summary 어플리케이션 별 유저 접속 수 조회
// @Tags Metric
// @Produce  json
// @Router /metric/login/application [get]
// @Param date query string true "Date count"
// @Success 200 {object} []models.MetricItem
// @Failure 500
func GetLoginApplication(c *gin.Context) {
	m, err := iamdb.GetLoginApplication(c.MustGet("date").(int) - 1)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, m)
}

// token godoc
// @Summary 유저 접속 로그 출력
// @Tags Metric
// @Produce  json
// @Router /metric/login/application/log [get]
// @Param date query string true "Date"
// @Success 200 {object} []models.MetricAppItem
// @Failure 500
func GetLoginApplicationLog(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		common.ErrorProcess(c, nil, http.StatusBadRequest, "required 'date'")
		return
	}
	m, err := iamdb.GetLoginApplicationLog(date)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, m)
}

// token godoc
// @Summary 일자 별 유저 접속 수 조회
// @Tags Metric
// @Produce  json
// @Router /metric/login/application/date [get]
// @Param date query string true "Date count"
// @Success 200 {object} []models.MetricAppItem
// @Failure 500
func GetLoginApplicationDate(c *gin.Context) {
	m, err := iamdb.GetLoginApplicationDate(c.MustGet("date").(int) - 1)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, m)
}

// token godoc
// @Summary 유형 별 로그인 실패 수 조회
// @Tags Metric
// @Produce  json
// @Router /metric/login/error [get]
// @Param date query string true "Date count"
// @Success 200 {object} []models.MetricItem
// @Failure 500
func GetLoginError(c *gin.Context) {
	m, err := iamdb.GetLoginError(c.MustGet("date").(int) - 1)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, m)
}

// token godoc
// @Summary ID 제공자 수 조회
// @Tags Metric
// @Produce  json
// @Router /metric/idp/count [get]
// @Success 200 {object} []models.MetricItem
// @Failure 500
func GetIdpCount(c *gin.Context) {
	m, err := iamdb.GetIdpCount()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, m)
}
