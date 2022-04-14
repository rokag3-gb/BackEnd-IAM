package main

import (
	"context"
	"errors"
	"fmt"
	"iam/api"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var ctx = context.Background()

var clientID = "test_service"
var clientSecret = "d7c2424e-7dfc-4a74-a6c5-bd6588ba2d73"
var realm = "test_realm"
var key_uri = "https://iam.cloudmt.co.kr"

func main() {

	route := gin.Default()
	route.Use(introspect_middleware())

	groups := route.Group("/groups")
	{
		groups.GET("/", api.GetGroup)
		groups.POST("/", api.CreateGroup)
		groups.DELETE("/:groupid", api.DeleteGroup)
		groups.PUT("/:groupid", api.UpdateGroup)
		groups.GET("/:groupid/members", api.GetGroupMember)
	}

	users := route.Group("/users")
	{
		users.GET("/", api.Users)
	}

	secret := route.Group("/secret")
	{
		secret.GET("/", api.Secret)
	}

	route.Run(":8085")
}

func introspect_middleware() gin.HandlerFunc {
	return func(c *gin.Context) {

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

func badRequeset(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func getUsernameJWT(token string) string {
	t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(""), nil
	})
	if err != nil {
		return ""
	}

	if !t.Valid {
		return ""
	}

	claims := t.Claims.(jwt.MapClaims)
	tmp := claims["preferred_username"]
	username := fmt.Sprintf("%v", tmp)

	return username
}
