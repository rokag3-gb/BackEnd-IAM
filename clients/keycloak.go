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

func InitKeycloakClient(clientID string, clientSecret string, realm string, endpoint string) error {
	if keycloakConfig == (kcConfig{}) {
		keycloakConfig = kcConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Realm:        realm,
			Endpoint:     endpoint,
		}
	}
	if keycloakClient == nil {
		keycloakClient = gocloak.NewClient(keycloakConfig.Endpoint, gocloak.SetAuthAdminRealms("admin/realms"), gocloak.SetAuthRealms("realms"))
		//		restyClient := keycloakClient.RestyClient()
		//		restyClient.SetDebug(true)
		//		restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	var ctx = context.Background()
	_, err := KeycloakToken(ctx)
	if err != nil {
		panic("Keycloak Client Init fail " + err.Error())
	}

	return nil
}

func KeycloakClient() gocloak.GoCloak {
	return keycloakClient
}

func KeycloakConfig() kcConfig {
	return keycloakConfig
}

func KeycloakToken(ctx context.Context) (*gocloak.JWT, error) {
	token, err := keycloakClient.LoginClient(ctx,
		keycloakConfig.ClientID,
		keycloakConfig.ClientSecret,
		keycloakConfig.Realm)
	if err != nil {
		return nil, err
	}
	return token, nil
}
