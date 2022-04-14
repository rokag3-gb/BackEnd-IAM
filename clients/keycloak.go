package clients

import (
	"context"

	"github.com/Nerzal/gocloak/v11"
)

var KeycloakClientID = "vault"
var KeycloakClientSecret = "ahvVBsKLPZXbt5PA5JicckZdx1sTriCR"
var KeycloakRealm = "iam"
var KeycloakEndpoint = "http://127.0.0.1:8080"

// var KeycloakClientID = "test_service"
// var KeyCloakClientSecret = "d7c2424e-7dfc-4a74-a6c5-bd6588ba2d73"
// var KeyCloakRealm = "test_realm"
// var KeycloakEndpoint = "https://iam.cloudmt.co.kr"

var keycloakClient gocloak.GoCloak = nil

func InitKeycloakClient(ctx context.Context) {
	if keycloakClient == nil {
		keycloakClient = gocloak.NewClient(KeycloakEndpoint)
		keycloakClient.LoginClient(ctx, KeycloakClientID, KeycloakClientSecret, KeycloakRealm)
	}
}

func KeycloakClient() gocloak.GoCloak {
	return keycloakClient
}
