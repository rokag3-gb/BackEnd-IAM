package middlewares

import (
	"fmt"
	"iam/clients"
	"iam/iamdb"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gin-gonic/gin"
)

func IntrospectMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.Request.Header.Get("Authorization")
		if auth == "" {
			c.String(http.StatusForbidden, "No Authorization header provided")
			c.Abort()
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		_, err := clients.KeycloakClient().RetrospectToken(c, token,
			clients.KeycloakConfig().ClientID,
			clients.KeycloakConfig().ClientSecret,
			clients.KeycloakConfig().Realm)
		if err != nil {
			c.String(http.StatusForbidden, "Authorization is not valid")
			c.Abort()
			panic("Inspection failed:" + err.Error())
		}

		c.Set("accessToken", token)
		c.Set("username", getUsernameJWT(token))
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

func AuthorityCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		realm := os.Getenv("KEYCLOAK_REALM")
		username := c.MustGet("username").(string)

		rows, err := iamdb.GetUserAuthoritiesForEndpoint(username, realm, c.Request.Method, c.Request.URL.Path)

		if err != nil {
			c.Abort()
		}
		defer rows.Close()

		if !rows.Next() {
			c.Status(http.StatusForbidden)
			c.Abort()

			// 결과가 한건도 오지 않으면 권한이 없음
		}

		// 결과가 한건이라도 있으면 권한 있음
	}
}

func AccessControlAllowOrigin() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "*")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
	}
}

func getUsernameJWT(token string) string {
	t, _ := jwt.Parse(token, nil)
	if t == nil {
		return ""
	}

	claims, _ := t.Claims.(jwt.MapClaims)
	return fmt.Sprintf("%v", claims["preferred_username"])
}
