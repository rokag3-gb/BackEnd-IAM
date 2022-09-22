package clients

import (
	"net/http"
	"time"

	"github.com/hashicorp/vault/api"
)

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
	vaultClient := VaultClient()
	if vaultClient == nil {
		vaultClient, _ = api.NewClient(&api.Config{Address: vaultConfig.Endpoint, HttpClient: httpClient})
		vaultClient.SetToken(vaultConfig.Token)
	}

	_, err := vaultClient.Logical().Read("sys/mounts")
	if err != nil {
		panic("Vault Client Init fail " + err.Error())
	}

	return nil
}

func VaultClient() *api.Client {
	var vaultClient, _ = api.NewClient(&api.Config{Address: vaultConfig.Endpoint, HttpClient: httpClient})
	vaultClient.SetToken(vaultConfig.Token)
	return vaultClient
}

func VaultConfig() vltConfig {
	return vaultConfig
}
