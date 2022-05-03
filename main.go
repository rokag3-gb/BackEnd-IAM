package main

import (
	"iam/api"
	"iam/clients"
	"iam/iamdb"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"iam/config"
	"iam/middlewares"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

var g errgroup.Group

func main() {
	var conf config.IamConfig
	if err := conf.InitConfig(); err != nil {
		panic(err.Error())
	}

	clients.InitKeycloakClient(
		conf.Keycloak_client_id,
		conf.Keycloak_client_secret,
		conf.Keycloak_realm,
		conf.Keycloak_endpoint)

	clients.InitVaultClient(
		conf.Vault_token,
		conf.Vault_endpoint)

	iamdb.InitDbClient("mssql", conf.Db_connect_string)

	route := makeRouter(conf)

	if conf.Http_port != "" {
		http_server := &http.Server{
			Addr:         ":" + conf.Http_port,
			Handler:      route,
			ReadTimeout:  time.Duration(conf.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(conf.WriteTimeout) * time.Second,
		}
		g.Go(func() error {
			err := http_server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Fatal(err)
			}
			return err
		})
	}

	if conf.Https_port != "" && conf.Https_certfile != "" && conf.Https_keyfile != "" {
		https_server := &http.Server{
			Addr:         ":" + conf.Https_port,
			Handler:      route,
			ReadTimeout:  time.Duration(conf.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(conf.WriteTimeout) * time.Second,
		}

		g.Go(func() error {
			err := https_server.ListenAndServeTLS(conf.Https_certfile, conf.Https_keyfile)
			if err != nil && err != http.ErrServerClosed {
				log.Fatal(err)
			}
			return err
		})
	}

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}

func makeRouter(conf config.IamConfig) *gin.Engine {
	route := gin.Default()

	route.Use(middlewares.AccessControlAllowOrigin(conf))
	route.Use(middlewares.IntrospectMiddleware(conf))
	route.Use(middlewares.AuthorityCheckMiddleware(conf))

	authority := route.Group("/authority")
	{
		authority.GET("/roles", api.GetRoles)
		authority.POST("/roles", api.CreateRoles)
		authority.DELETE("/roles/:roleid", api.DeleteRoles)
		authority.PUT("/roles/:roleid", api.UpdateRoles)
		authority.GET("/roles/:roleid/auth", api.GetRolesAuth)
		authority.POST("/roles/:roleid/auth", api.AssignRoleAuth)
		authority.DELETE("/roles/:roleid/auth/:authid", api.DismissRoleAuth)
		authority.GET("/user/:userid", api.GetUserRole)
		authority.POST("/user/:userid/roles", api.AssignUserRole)
		authority.DELETE("/user/:userid/roles/:roleid", api.DismissUserRole)
		authority.GET("/user/:userid/auth", api.GetUserAuth)
		authority.GET("/user/:userid/auth/:authid", api.GetUserAuthActive) //실제로 전달되는것은 username과 role name 입니다. gin 제한사항으로 인하여 이름 변경이 불가능
		authority.GET("/auth", api.GetAuth)
		authority.POST("/auth", api.CreateAuth)
		authority.DELETE("/auth/:authid", api.DeleteAuth)
		authority.PUT("/auth/:authid", api.UpdateAuth)
		authority.GET("/auth/:authid", api.GetAuthInfo)
	}

	groups := route.Group("/groups")
	{
		groups.GET("", middlewares.ListQueryRangeMiddleware(), api.GetGroup)
		groups.POST("", api.CreateGroup)
		groups.DELETE("/:groupid", api.DeleteGroup)
		groups.PUT("/:groupid", api.UpdateGroup)
		groups.GET("/:groupid/members", middlewares.ListQueryRangeMiddleware(), api.GetGroupMember)
	}

	users := route.Group("/users")
	{
		users.GET("", middlewares.ListQueryRangeMiddleware(), api.Users)
		users.POST("", api.CreateUser)
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
		secret.GET("", api.GetSecretGroup)
		secret.POST("", api.CreateSecretGroup)
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

	for _, name := range conf.Api_host_name {
		target, err := url.Parse(conf.Api_host_list[name])
		if err != nil {
			panic(err.Error())
		}

		proxy := httputil.NewSingleHostReverseProxy(target)
		route.Any("/"+name, func(c *gin.Context) {
			proxy.ServeHTTP(c.Writer, c.Request)
		})
	}

	return route
}
