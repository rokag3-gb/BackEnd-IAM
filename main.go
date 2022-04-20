package main

import (
	"context"
	"errors"
	"fmt"
	"iam/api"
	"iam/clients"
	"net/http"
	"os"

	"iam/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var ctx = context.Background()

const VaultToken = "hvs.wiselTCiHpNilcjoAyfP7GDK"
const VaultEndpoint = "http://20.214.161.230:8200"

func main() {

	clients.InitKeycloakClient(
		os.Getenv("KEYCLOAK_CLIENT_ID"),
		os.Getenv("KEYCLOAK_CLIENT_SECRET"),
		os.Getenv("KEYCLOAK_REALM"),
		os.Getenv("KEYCLOAK_ENDPOINT"))

	clients.InitVaultClient(VaultToken, VaultEndpoint)

	route := gin.Default()

	//	route.Use(middlewares.IntrospectMiddleware())

	groups := route.Group("/groups")
	{
		groups.GET("/", middlewares.ListQueryRangeMiddleware(), api.GetGroup)
		groups.POST("/", api.CreateGroup)
		groups.DELETE("/:groupid", api.DeleteGroup)
		groups.PUT("/:groupid", api.UpdateGroup)
		groups.GET("/:groupid/members", middlewares.ListQueryRangeMiddleware(), api.GetGroupMember)
	}

	users := route.Group("/users")
	{
		users.GET("/", middlewares.ListQueryRangeMiddleware(), api.Users)
		users.POST("/", api.CreateUser)
		users.PUT("/:userid", api.UpdateUser)
		users.DELETE("/:userid", api.DeleteUser)
		users.GET("/:userid", api.GetUser)
		users.GET("/:userid/credentials", api.GetUserCredentials)
		users.PUT("/:userid/reset-password", api.ResetUserPassword)
		users.GET("/:userid/groups", middlewares.ListQueryRangeMiddleware(), api.GetUserGroups)
		users.PUT("/:userid/groups/:groupid", api.AddUserToGroup)
		users.DELETE("/:userid/groups/:groupid", api.DeleteUserFromGroup)
		users.GET("/:userid/sessions", api.GetUserSessions)
		users.DELETE("/:userid/sessions/:sessionid", api.LogoutUserSession)
		users.POST("/:userid/logout", api.LogoutAllSessions)
		users.GET("/:userid/federated-identity", api.GetUserFederatedIdentities)
	}

	secret := route.Group("/secret")
	{
		secret.GET("/", api.GetSecretGroup)
		secret.POST("/", api.CreateSecretGroup)
		secret.DELETE("/:groupName", api.DeleteSecretGroup)
		secret.GET("/:groupName", api.GetSecretList)
		secret.GET("/:groupName/data/:secretName", api.GetSecret)
		secret.POST("/:groupName/data/:secretName", api.MargeSecret)
		secret.GET("/:groupName/metadata/:secretName", api.GetMetadataSecret)
		secret.POST("/:groupName/delete/:secretName", api.DeleteSecret)
		secret.POST("/:groupName/undelete/:secretName", api.UndeleteSecret)
		secret.POST("/:groupName/destroy/:secretName", api.DestroySecret)
		secret.DELETE("/:groupName/metadata/:secretName", api.DeleteSecretMetadata)
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
