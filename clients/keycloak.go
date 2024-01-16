package clients

import (
	"context"
	"encoding/json"

	"github.com/Nerzal/gocloak/v11"
)

var keycloakClient gocloak.GoCloak = nil
var keycloakConfig kcConfig = kcConfig{}

type kcConfig struct {
	ClientID     string
	ClientSecret string
	Endpoint     string
}

func InitKeycloakClient(clientID string, clientSecret string, endpoint string) error {
	if keycloakConfig == (kcConfig{}) {
		keycloakConfig = kcConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     endpoint,
		}
	}
	if keycloakClient == nil {
		keycloakClient = gocloak.NewClient(keycloakConfig.Endpoint, gocloak.SetAuthAdminRealms("admin/realms"), gocloak.SetAuthRealms("realms"))
		//		restyClient := keycloakClient.RestyClient()
		//		restyClient.SetDebug(true)
		//		restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return nil
}

func KeycloakClient() gocloak.GoCloak {
	return keycloakClient
}

func KeycloakConfig() kcConfig {
	return keycloakConfig
}

func KeycloakToken(ctx context.Context, realm string) (*gocloak.JWT, error) {
	token, err := keycloakClient.LoginClient(ctx,
		keycloakConfig.ClientID,
		keycloakConfig.ClientSecret,
		realm)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func Token(ctx context.Context, clientid string, clientSecret string, username string, password string, realm string) (*gocloak.JWT, error) {
	token, err := keycloakClient.Login(ctx,
		clientid,
		clientSecret,
		realm,
		username,
		password)

	if err != nil {
		return nil, err
	}
	return token, nil
}

func TokenRefresh(ctx context.Context, refreshToken string, clientid string, clientSecret string, realm string) (*gocloak.JWT, error) {
	token, err := keycloakClient.RefreshToken(ctx,
		refreshToken,
		clientid,
		clientSecret,
		realm)

	if err != nil {
		return nil, err
	}
	return token, nil
}

func TokenLogout(ctx context.Context, refreshToken string, clientid string, clientSecret string, realm string) error {
	err := keycloakClient.Logout(ctx,
		clientid,
		clientSecret,
		realm,
		refreshToken)

	if err != nil {
		return err
	}
	return nil
}

func TokenGetToken(ctx context.Context, data []byte, secret *string, realm string) (*gocloak.JWT, error) {
	options := gocloak.TokenOptions{}
	err := json.Unmarshal([]byte(data), &options)
	if secret != nil {
		options.ClientSecret = secret
	}

	token, err := keycloakClient.GetToken(ctx, realm, options)

	if err != nil {
		return nil, err
	}
	return token, nil
}
