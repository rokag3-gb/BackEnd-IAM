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

func Token(ctx context.Context, clientid string, clientSecret string, username string, password string) (*gocloak.JWT, error) {
	token, err := keycloakClient.Login(ctx,
		clientid,
		clientSecret,
		keycloakConfig.Realm,
		username,
		password)

	if err != nil {
		return nil, err
	}
	return token, nil
}

func TokenRefresh(ctx context.Context, refreshToken string, clientid string, clientSecret string) (*gocloak.JWT, error) {
	token, err := keycloakClient.RefreshToken(ctx,
		refreshToken,
		clientid,
		clientSecret,
		keycloakConfig.Realm)

	if err != nil {
		return nil, err
	}
	return token, nil
}

func TokenLogout(ctx context.Context, refreshToken string, clientid string, clientSecret string) error {
	err := keycloakClient.Logout(ctx,
		clientid,
		clientSecret,
		keycloakConfig.Realm,
		refreshToken)

	if err != nil {
		return err
	}
	return nil
}

func TokenGetToken(ctx context.Context, data []byte, secret *string) (*gocloak.JWT, error) {
	options := gocloak.TokenOptions{}
	err := json.Unmarshal([]byte(data), &options)
	if secret != nil {
		options.ClientSecret = secret
	}

	token, err := keycloakClient.GetToken(ctx, keycloakConfig.Realm, options)

	if err != nil {
		return nil, err
	}
	return token, nil
}
