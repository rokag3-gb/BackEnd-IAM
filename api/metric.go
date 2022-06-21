package api

import (
	"encoding/json"
	"fmt"
	"iam/clients"
	"iam/config"
	"iam/iamdb"
	"iam/models"
	"io/ioutil"
	"net/http"
	"strconv"

	logger "cloudmt.co.kr/mateLogger"
	"github.com/gin-gonic/gin"
)

func MetricCount(c *gin.Context) {
	count, err := iamdb.MetricCount()

	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, count)
}

func GetMetricSession(c *gin.Context) {
	token, _ := clients.KeycloakToken(c)

	url := fmt.Sprintf("%s/admin/realms/%s/client-session-stats", config.GetConfig().Keycloak_endpoint, config.GetConfig().Keycloak_realm)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	req.Header.Add("authorization", "Bearer "+token.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)

	arr := make([]map[string]interface{}, 0)

	json.Unmarshal(bytes, &arr)

	apps, err := iamdb.GetApplications()
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
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

func GetLoginApplication(c *gin.Context) {
	m, err := iamdb.GetLoginApplication(c.MustGet("date").(int) - 1)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, m)
}

func GetLoginDate(c *gin.Context) {
	m, err := iamdb.GetLoginDate(c.MustGet("date").(int) - 1)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, m)
}

func GetLoginError(c *gin.Context) {
	m, err := iamdb.GetLoginError(c.MustGet("date").(int) - 1)
	if err != nil {
		logger.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.JSON(http.StatusOK, m)
}
