package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Conf struct {
	Keycloak_client_id           string
	Keycloak_client_secret       string
	Keycloak_realm               string
	Keycloak_endpoint            string
	Vault_token                  string
	Vault_endpoint               string
	Db_connect_string            string
	Http_port                    string
	Https_port                   string
	Https_certfile               string
	Https_keyfile                string
	ReadTimeout                  int
	WriteTimeout                 int
	Access_control_allow_origin  string
	Access_control_allow_headers string
	Api_host_list                map[string]string
	Api_host_name                []string
}

func (conf *Conf) InitEnvConfig() error {
	conf.Keycloak_client_id = os.Getenv("KEYCLOAK_CLIENT_ID")
	conf.Keycloak_client_secret = os.Getenv("KEYCLOAK_CLIENT_SECRET")
	conf.Keycloak_realm = os.Getenv("KEYCLOAK_REALM")
	conf.Keycloak_endpoint = os.Getenv("KEYCLOAK_ENDPOINT")

	if conf.Keycloak_client_id == "" || conf.Keycloak_client_secret == "" || conf.Keycloak_realm == "" || conf.Keycloak_endpoint == "" {
		return errors.New("check [KEYCLOAK] config")
	}

	conf.Vault_token = os.Getenv("VAULT_TOKEN")
	conf.Vault_endpoint = os.Getenv("VAULT_ENDPOINT")

	if conf.Vault_token == "" || conf.Vault_endpoint == "" {
		return errors.New("check [VAULT] config")
	}

	db_server := os.Getenv("DB_SERVER_ADDR")
	db_name := os.Getenv("DB_SERVER_NAME")
	db_port := os.Getenv("DB_SERVER_PORT")
	db_user_id := os.Getenv("DB_USER_ID")
	db_password := os.Getenv("DB_USER_PASSWORD")

	if db_server == "" || db_name == "" || db_port == "" || db_user_id == "" || db_password == "" {
		return errors.New("check [DATABASE] config")
	}

	conf.Db_connect_string = fmt.Sprintf("server=%s;database=%s;user id=%s;password=%s;port=%s", db_server, db_name, db_user_id, db_password, db_port)

	conf.Http_port = os.Getenv("HTTP_PORT")
	conf.Https_port = os.Getenv("HTTPS_PORT")
	conf.Https_certfile = os.Getenv("SSL_CERTFILE_PATH")
	conf.Https_keyfile = os.Getenv("SSL_KEYFILE_PATH")

	var err error
	conf.ReadTimeout, err = strconv.Atoi(os.Getenv("NETWROK_READ_TIMEOUT"))
	if err != nil {
		conf.ReadTimeout = 5
	}

	conf.WriteTimeout, err = strconv.Atoi(os.Getenv("NETWROK_WRITE_TIMEOUT"))
	if err != nil {
		conf.WriteTimeout = 10
	}

	conf.Access_control_allow_origin = os.Getenv("ACCESS_CONRTOL_ALLOW_ORIGIN")
	conf.Access_control_allow_headers = os.Getenv("ACCESS_CONRTOL_ALLOW_HEADERS")

	if conf.Access_control_allow_origin == "" {
		conf.Access_control_allow_origin = "*"
	}

	if conf.Access_control_allow_headers == "" {
		conf.Access_control_allow_headers = "*"
	}

	conf.Api_host_name = strings.Split(os.Getenv("API_HOST_NAME"), ",")
	api_host_url := strings.Split(os.Getenv("API_HOST_URL"), ",")

	if len(conf.Api_host_name) != len(api_host_url) {
		return errors.New("check config [API]")
	}

	conf.Api_host_list = map[string]string{}
	for i := 0; i < len(conf.Api_host_name); i++ {
		conf.Api_host_list[conf.Api_host_name[i]] = api_host_url[i]
	}

	return nil
}
