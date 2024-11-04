package middlewares

import (
	"errors"
	"fmt"
	"iam/clients"
	"iam/common"
	"iam/config"
	"iam/query"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gin-gonic/gin"
)

func GetUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.RequestURI, "/swagger") {
			return
		}
		auth := c.Request.Header.Get("Authorization")
		if auth == "" && !config.GetConfig().Developer_mode {
			common.ErrorProcess(c, nil, http.StatusForbidden, "No Authorization header provided")
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")

		username, userid, realm, tenantId, err := getDataJWT(token)
		if err != nil {
			common.ErrorProcess(c, err, http.StatusInternalServerError, "")
			return
		}

		c.Set("userId", userid)
		c.Set("username", username)
		c.Set("realm", realm)
		c.Set("accessToken", token)
		c.Set("tenantId", tenantId)
	}
}

func getDataJWT(token string) (string, string, string, string, error) {
	var username, userId, realm, tenantId string

	t, _ := jwt.Parse(token, nil)
	if t == nil {
		return username, userId, realm, tenantId, errors.New("invalid authorization")
	}

	claims, _ := t.Claims.(jwt.MapClaims)
	if claims == nil {
		return username, userId, realm, tenantId, errors.New("invalid token")
	}

	username = fmt.Sprintf("%v", claims["preferred_username"])
	if username == "" {
		return username, userId, realm, tenantId, errors.New("invalid token")
	}

	userId = fmt.Sprintf("%v", claims["sub"])
	if userId == "" {
		return username, userId, realm, tenantId, errors.New("invalid token")
	}

	realm = fmt.Sprintf("%v", claims["iss"])
	if realm == "" {
		return username, userId, realm, tenantId, errors.New("invalid token")
	}
	tmp := strings.Split(realm, "/")
	realm = tmp[len(tmp)-1]

	tenantId = fmt.Sprintf("%v", claims["tenantId"])

	return username, userId, realm, tenantId, nil
}

func ListQueryRangeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		first, firstErr := strconv.Atoi(c.DefaultQuery("first", "0"))
		if firstErr != nil {
			common.ErrorProcess(c, nil, http.StatusBadRequest, "'first' must be integer")
			return
		}
		c.Set("first", first)
		max, maxErr := strconv.Atoi(c.DefaultQuery("max", "100"))
		if maxErr != nil {
			common.ErrorProcess(c, nil, http.StatusBadRequest, "'max' must be integer")
			return
		}
		c.Set("max", max)
	}
}

func DateQueryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		date, err := strconv.Atoi(c.DefaultQuery("date", "7"))
		if err != nil {
			common.ErrorProcess(c, nil, http.StatusBadRequest, "'date' must be integer")
			return
		}
		c.Set("date", date)
	}
}

func PageNotFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	}
}

func CheckAccountRequestUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		db, err := clients.DBClient()
		if err != nil {
			common.ErrorProcess(c, err, http.StatusInternalServerError, err.Error())
			return
		}
		defer db.Close()

		result, err := query.CheckAccountUser(db, c.Param("accountId"), c.GetString("userId"), c.GetString("realm"))
		if err != nil {
			common.ErrorProcess(c, err, http.StatusInternalServerError, err.Error())
			return
		}

		if !result {
			common.ErrorProcess(c, nil, http.StatusForbidden, "invalid authorization")
			return
		}
	}
}

func CheckAccountUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		db, err := clients.DBClient()
		if err != nil {
			common.ErrorProcess(c, err, http.StatusInternalServerError, err.Error())
			return
		}
		defer db.Close()

		result, err := query.CheckAccountUser(db, c.Param("accountId"), c.Param("userid"), c.GetString("realm"))
		if err != nil {
			common.ErrorProcess(c, err, http.StatusInternalServerError, err.Error())
			return
		}

		if !result {
			common.ErrorProcess(c, nil, http.StatusBadRequest, "bad request")
			return
		}
	}
}

func GetInitInfo(token string) (string, string, error) {
	t, _ := jwt.Parse(token, nil)
	if t == nil {
		return "", "", errors.New("invalid authorization")
	}

	claims, _ := t.Claims.(jwt.MapClaims)
	if claims == nil {
		return "", "", errors.New("invalid token")
	}

	email := fmt.Sprintf("%v", claims["email"])
	if email == "" {
		return "", "", errors.New("invalid email")
	}

	client_id := fmt.Sprintf("%v", claims["azp"])
	if email == "" {
		return "", "", errors.New("invalid client_id")
	}

	return email, client_id, nil
}

func AccessControlAllowOrigin() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", config.GetConfig().Access_control_allow_origin)
		c.Header("Access-Control-Allow-Headers", config.GetConfig().Access_control_allow_headers)
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
	}
}
