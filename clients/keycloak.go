package clients

import (
	"context"
	"encoding/json"
	"errors"
	"iam/models"

	logger "cloudmt.co.kr/mateLogger"
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

func KeycloakToken(ctx context.Context) (*gocloak.JWT, error) {
	token, err := keycloakClient.LoginClient(ctx,
		keycloakConfig.ClientID,
		keycloakConfig.ClientSecret,
		"master")
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

func GetServiceAccountSecret(ctx context.Context, token, realm, clientid string) (models.ClientSecret, error) {
	result := models.ClientSecret{}
	clients, err := keycloakClient.GetClients(ctx, token, realm, gocloak.GetClientsParams{ClientID: &clientid})
	if err != nil {
		return models.ClientSecret{}, err
	}

	for _, client := range clients {
		if client.ID != nil && *client.ClientID == clientid {
			credential, err := keycloakClient.GetClientSecret(ctx, token, realm, *client.ID)
			if err != nil {
				return models.ClientSecret{}, err
			}

			result.Type = credential.Type
			result.Value = credential.Value

			return result, nil
		}
	}

	return result, errors.New("ServiceAccount not found")
}

func RegenerateServiceAccountSecret(ctx context.Context, token, realm, clientid string) (models.ClientSecret, error) {
	result := models.ClientSecret{}
	clients, err := keycloakClient.GetClients(ctx, token, realm, gocloak.GetClientsParams{ClientID: &clientid})
	if err != nil {
		return models.ClientSecret{}, err
	}

	for _, client := range clients {
		if client.ID != nil && *client.ClientID == clientid {
			credential, err := keycloakClient.RegenerateClientSecret(ctx, token, realm, *client.ID)
			if err != nil {
				return models.ClientSecret{}, err
			}

			result.Type = credential.Type
			result.Value = credential.Value

			return result, nil
		}
	}

	return result, errors.New("ServiceAccount not found")
}

func CreateServiceAccount(ctx context.Context, token, realm, clientid string) (error, string) {
	ServiceAccountsEnabled := true
	idOfClient, err := keycloakClient.CreateClient(ctx, token, realm, gocloak.Client{ClientID: &clientid, ServiceAccountsEnabled: &ServiceAccountsEnabled})
	if err != nil {
		return err, ""
	}

	_, err = keycloakClient.RegenerateClientSecret(ctx, token, realm, idOfClient)
	if err != nil {
		return err, ""
	}

	logger.Info("CreateServiceAccountSecret : %s", idOfClient)

	return nil, idOfClient
}

func UpdateServiceAccount(ctx context.Context, token, realm, idOfClient, clientId string, Enabled bool) error {
	ServiceAccountsEnabled := true

	client := gocloak.Client{
		ID:                     &idOfClient,
		ClientID:               &clientId,
		Enabled:                &Enabled,
		ServiceAccountsEnabled: &ServiceAccountsEnabled,
	}

	err := keycloakClient.UpdateClient(ctx, token, realm, client)
	if err != nil {
		return err
	}

	logger.Info("CreateServiceAccountSecret : %s", idOfClient)

	return nil
}

func DeleteServiceAccount(ctx context.Context, token, realm, clientid string) error {
	clients, err := keycloakClient.GetClients(ctx, token, realm, gocloak.GetClientsParams{ClientID: &clientid})
	if err != nil {
		return err
	}

	for _, client := range clients {
		if client.ID != nil && *client.ClientID == clientid {
			err := keycloakClient.DeleteClient(ctx, token, realm, *client.ID)
			if err != nil {
				return err
			}

			return nil
		}
	}

	return errors.New("ServiceAccount not found")
}
