package middlewares

import (
	"errors"
	"fmt"
	"iam/common"
	"iam/config"
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

		username, err := getUsernameJWT(token)
		if err != nil {
			common.ErrorProcess(c, err, http.StatusInternalServerError, "")
			return
		}

		userid, err := getUserIdJWT(token)
		if err != nil {
			common.ErrorProcess(c, err, http.StatusInternalServerError, "")
			return
		}

		c.Set("userId", userid)
		c.Set("username", username)
		c.Set("accessToken", token)
	}
}

func getUsernameJWT(token string) (string, error) {
	t, _ := jwt.Parse(token, nil)
	if t == nil {
		if config.GetConfig().Developer_mode {
			return "admin", nil
		}
		return "", errors.New("invalid authorization")
	}

	claims, _ := t.Claims.(jwt.MapClaims)
	if claims == nil {
		return "", errors.New("invalid token")
	}

	username := fmt.Sprintf("%v", claims["preferred_username"])
	if username == "" {
		return "", errors.New("invalid token")
	}

	return username, nil
}

func getUserIdJWT(token string) (string, error) {
	t, _ := jwt.Parse(token, nil)
	if t == nil {
		return "", errors.New("invalid authorization")
	}

	claims, _ := t.Claims.(jwt.MapClaims)
	if claims == nil {
		return "", errors.New("invalid token")
	}

	username := fmt.Sprintf("%v", claims["sub"])
	if username == "" {
		return "", errors.New("invalid token")
	}

	return username, nil
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
