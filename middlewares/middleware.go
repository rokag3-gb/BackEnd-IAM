package middlewares

import (
	"fmt"
	"iam/clients"
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

		// 클라이언트가 보낸 토큰에 대한 keycloak 인증 부분입니다.
		/*
			token := c.Request.Header.Get("token")
			if token == "" {
				c.String(http.StatusUnauthorized, "Unauthorized")
				panic("Unauthorized")
			}

			username := getUsernameJWT(token)
			if username == "" {
				c.String(http.StatusUnauthorized, "Unauthorized")
				panic("Unauthorized")
			}
			// 여기서 구한 username 으로 권한 체크를 해야합니다.
			// 다만 keycloak 설정에 따라 토큰의 내용이 변경될 수도 있으므로 이후 테스트 필요...

			var client = gocloak.NewClient(key_uri)
			rptResult, err := client.RetrospectToken(token, clientID, clientSecret, realm)
			if err != nil {
				c.String(http.StatusUnauthorized, "Unauthorized")
				panic("Unauthorized")
			}

			if !rptResult.Active {
				c.String(http.StatusUnauthorized, "Unauthorized")
				panic("Unauthorized")
			}
		*/
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

		query := fmt.Sprintf(`select
			1
			from
			USER_ENTITY u join
			user_roles_mapping ur
			on
			u.id = ur.userId
			join
			roles_authority_mapping ra
			on
			ur.rId = ra.rId
			join
			authority a
			on
			ra.aId = a.aId
			where
			u.USERNAME = '%s' AND
			u.REALM_ID = '%s' AND
			(a.endpointMethod = '%s' OR a.endpointMethod = 'ALL') AND
			PATINDEX(REPLACE(a.endpointUrl,'*','%%'), '%s') = 1`, username, realm, c.Request.Method, c.Request.URL.Path)

		rows, err := clients.DBClient().Query(query)

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

func getUsernameJWT(token string) string {
	t, _ := jwt.Parse(token, nil)
	if t == nil {
		return ""
	}

	claims, _ := t.Claims.(jwt.MapClaims)
	return fmt.Sprintf("%v", claims["preferred_username"])
}
