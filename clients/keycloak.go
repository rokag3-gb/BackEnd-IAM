package clients

import (
	"context"

	"github.com/Nerzal/gocloak/v11"
)

var keycloakClient gocloak.GoCloak = nil
var keycloakConfig kcConfig = kcConfig{}

type kcConfig struct {
	ClientID     string
	ClientSecret string
	Realm        string
	Endpoint     string
}

func InitKeycloakClient(ctx context.Context, clientID string, clientSecret string, realm string, endpoint string) error {
	if keycloakConfig == (kcConfig{}) {
		keycloakConfig = kcConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Realm:        realm,
			Endpoint:     endpoint,
		}
	}
	if keycloakClient == nil {
		keycloakClient = gocloak.NewClient(keycloakConfig.Endpoint)
		_, err := keycloakClient.LoginClient(ctx, keycloakConfig.ClientID, keycloakConfig.ClientSecret, keycloakConfig.Realm)
		if err != nil {
			return err
		}
	}
	return nil
}

func KeycloakClient() gocloak.GoCloak {
	return keycloakClient
}

func KeycloakConfig() kcConfig {
	return keycloakConfig
}
