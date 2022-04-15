package clients

import (
	"net/http"
	"time"

	"github.com/hashicorp/vault/api"
)

var vaultClient *api.Client = nil
var vaultConfig vltConfig = vltConfig{}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

type vltConfig struct {
	Token    string
	Endpoint string
}

func InitVaultClient(token string, endpoint string) error {
	if vaultConfig == (vltConfig{}) {
		vaultConfig = vltConfig{
			Token:    token,
			Endpoint: endpoint,
		}
	}
	if vaultClient == nil {
		vaultClient, _ = api.NewClient(&api.Config{Address: vaultConfig.Endpoint, HttpClient: httpClient})
		vaultClient.SetToken(vaultConfig.Token)
	}
	return nil
}

func VaultClient() *api.Client {
	return vaultClient
}

func VaultConfig() vltConfig {
	return vaultConfig
}
