package main

import (
	"context"
	"errors"
	"fmt"
	"iam/api"
	"iam/clients"
	"net/http"

	"iam/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var KeycloakClientID = "test_service"
var KeycloakClientSecret = "d7c2424e-7dfc-4a74-a6c5-bd6588ba2d73"
var KeycloakRealm = "test_realm"
var KeycloakEndpoint = "https://iam.cloudmt.co.kr"
var ctx = context.Background()

func main() {

	clients.InitKeycloakClient(KeycloakClientID, KeycloakClientSecret, KeycloakRealm, KeycloakEndpoint)

	route := gin.Default()

	route.Use(middlewares.IntrospectMiddleware())

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
