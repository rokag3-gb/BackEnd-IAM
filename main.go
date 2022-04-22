package main

import (
	"context"
	"iam/api"
	"iam/clients"
	"iam/iamdb"
	"os"

	"iam/middlewares"

	"github.com/gin-gonic/gin"
)

var ctx = context.Background()

func main() {
	clients.InitKeycloakClient(
		os.Getenv("KEYCLOAK_CLIENT_ID"),
		os.Getenv("KEYCLOAK_CLIENT_SECRET"),
		os.Getenv("KEYCLOAK_REALM"),
		os.Getenv("KEYCLOAK_ENDPOINT"))

	clients.InitVaultClient(
		os.Getenv("VAULT_TOKEN"),
		os.Getenv("VAULT_ENDPOINT"))

	iamdb.InitDbClient("mssql", os.Getenv("DB_CONNECT_STRING"))

	route := gin.Default()

	route.Use(middlewares.IntrospectMiddleware())
	route.Use(middlewares.AuthorityCheckMiddleware())

	authority := route.Group("/authority")
	{
		authority.GET("/roles", api.GetRoles)
		authority.POST("/roles", api.CreateRoles)
		authority.DELETE("/roles/:roleid", api.DeleteRoles)
		authority.PUT("/roles/:roleid", api.GetRoles)
		authority.GET("/roles/:roleid/auth", api.GetRoles)
		authority.POST("/roles/:roleid/auth", api.GetRoles)
		authority.DELETE("/roles/:roleid/auth/:authid", api.GetRoles)
		authority.GET("/user/:userid", api.GetRoles)
		authority.POST("/user/:userid/roles", api.GetRoles)
		authority.DELETE("/user/:userid/roles/:roleid", api.GetRoles)
		authority.GET("/user/:userid/auth/:roleid", api.GetRoles) //실제로 전달되는것은 username과 role name 입니다. gin 제한사항으로 인하여 이름 변경이 불가능
		authority.GET("/user/:userid/auth", api.GetRoles)
		authority.GET("/auth", api.GetRoles)
		authority.POST("/auth", api.GetRoles)
		authority.DELETE("/auth/:authid", api.GetRoles)
		authority.PUT("/auth/:authid", api.GetRoles)
		authority.GET("/auth/:authid", api.GetRoles)
	}

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
