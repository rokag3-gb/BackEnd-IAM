package middlewares

import (
	"errors"
	"fmt"
	"iam/clients"
	"iam/config"
	"iam/iamdb"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	logger "cloudmt.co.kr/mateLogger"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gin-gonic/gin"
)

func IntrospectMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.Request.Header.Get("Authorization")
		if auth == "" && !config.GetConfig().Developer_mode {
			c.String(http.StatusForbidden, "No Authorization header provided")
			c.Abort()
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		result, err := clients.KeycloakClient().RetrospectToken(c, token,
			clients.KeycloakConfig().ClientID,
			clients.KeycloakConfig().ClientSecret,
			clients.KeycloakConfig().Realm)

		if err != nil && !config.GetConfig().Developer_mode {
			logger.Error(err.Error())
			if config.GetConfig().Developer_mode {
				c.String(http.StatusInternalServerError, err.Error())
			} else {
				c.Status(http.StatusInternalServerError)
			}
			c.Abort()
		}

		if !*result.Active && !config.GetConfig().Developer_mode {
			c.String(http.StatusForbidden, "Invalid authorization")
			c.Abort()
		}

		username, err := getUsernameJWT(token)
		if err != nil {
			logger.Error(err.Error())
			if config.GetConfig().Developer_mode {
				c.String(http.StatusInternalServerError, err.Error())
			} else {
				c.Status(http.StatusInternalServerError)
			}
			c.Abort()
		}

		c.Set("username", username)
		c.Set("accessToken", token)
	}
}

func ListQueryRangeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		first, firstErr := strconv.Atoi(c.DefaultQuery("first", "0"))
		if firstErr != nil {
			c.String(http.StatusBadRequest, "'first' must be integer")
			c.Abort()
			return
		}
		c.Set("first", first)
		max, maxErr := strconv.Atoi(c.DefaultQuery("max", "100"))
		if maxErr != nil {
			c.String(http.StatusBadRequest, "'max' must be integer")
			c.Abort()
			return
		}
		c.Set("max", max)
	}
}

func DateQueryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		date, err := strconv.Atoi(c.DefaultQuery("date", "7"))
		if err != nil {
			c.String(http.StatusBadRequest, "'date' must be integer")
			c.Abort()
			return
		}
		c.Set("date", date)
	}
}

func AuthorityCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.MustGet("username").(string)

		rows, err := iamdb.GetUserAuthoritiesForEndpoint(username, config.GetConfig().Keycloak_realm, c.Request.Method, c.Request.URL.Path)

		if err != nil {
			c.Abort()
		}
		defer rows.Close()

		if !rows.Next() && !config.GetConfig().Developer_mode {
			c.String(http.StatusForbidden, "Access Denied")
			c.Abort()

			// 결과가 한건도 오지 않으면 권한이 없음
		}

		// 결과가 한건이라도 있으면 권한 있음
	}
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

func getUsernameJWT(token string) (string, error) {
	t, _ := jwt.Parse(token, nil)
	if t == nil {
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

func RefreshApps(c *gin.Context) {
	_, err := iamdb.GetApplicationList()
	if err != nil {
		logger.Error(err.Error())

		if config.GetConfig().Developer_mode {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.Status(http.StatusInternalServerError)
		}
		c.Abort()
	}

	c.Status(http.StatusOK)
}

func ReturnReverseProxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := strings.Split(c.Request.URL.Path, "/")

		app_url, exist := config.GetConfig().Api_host_list[path[1]]

		if !exist {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		target, err := url.Parse(app_url)
		if err != nil {
			panic(err.Error())
		}

		c.Request.Host = target.Host
		c.Request.URL.Path = strings.TrimPrefix(c.Request.RequestURI, "/"+path[1])

		c.Request.Header.Del("Authorization")
		c.Request.Header.Del("authorization")

		proxy := httputil.NewSingleHostReverseProxy(target)

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
