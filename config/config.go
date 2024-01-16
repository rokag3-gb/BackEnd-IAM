package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/ini.v1"
)

var cfg *IamConfig = nil

func GetConfig() *IamConfig {
	if cfg == nil {
		InitConfig()
	}
	return cfg
}

type IamConfig struct {
	Keycloak_client_id           string
	Keycloak_client_secret       string
	Keycloak_endpoint            string
	Vault_token                  string
	Vault_endpoint               string
	Db_connect_string            string
	Sales_Reqeuest_Url           string
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
	Developer_mode               bool
	LogStdout                    bool
	UseApiDocument               bool
}

func InitConfig() error {
	ctype := os.Getenv("IAM_CONFIG_TYPE")

	if cfg == nil {
		cfg = new(IamConfig)
	}

	var err error
	if ctype == "env" {
		err = cfg.initEnvConfig()
	} else {
		err = cfg.initConf()
	}

	return err
}

func (conf *IamConfig) initConf() error {
	cfg, err := ini.Load("iam.conf")
	if err != nil {
		return err
	}

	conf.Keycloak_client_id = cfg.Section("keycloak").Key("client_id").String()
	conf.Keycloak_client_secret = cfg.Section("keycloak").Key("client_secret").String()
	conf.Keycloak_endpoint = cfg.Section("keycloak").Key("endpoint").String()

	if conf.Keycloak_client_id == "" || conf.Keycloak_client_secret == "" || conf.Keycloak_endpoint == "" {
		return errors.New("check config")
	}

	conf.Vault_token = cfg.Section("vault").Key("token").String()
	conf.Vault_endpoint = cfg.Section("vault").Key("endpoint").String()

	if conf.Vault_token == "" || conf.Vault_endpoint == "" {
		return errors.New("check config")
	}

	db_server := cfg.Section("database").Key("server_addr").String()
	db_name := cfg.Section("database").Key("name").String()
	db_user_id := cfg.Section("database").Key("user_id").String()
	db_password := cfg.Section("database").Key("user_password").String()
	db_port := cfg.Section("database").Key("server_port").String()

	if db_server == "" || db_name == "" || db_port == "" || db_user_id == "" || db_password == "" {
		return errors.New("check [DATABASE] config")
	}

	conf.Db_connect_string = fmt.Sprintf("server=%s;database=%s;user id=%s;password=%s;port=%s", db_server, db_name, db_user_id, db_password, db_port)

	conf.Sales_Reqeuest_Url = cfg.Section("sales").Key("reqeuest_Url").String()

	conf.Http_port = cfg.Section("network").Key("http_port").String()
	conf.Https_port = cfg.Section("network").Key("https_port").String()
	conf.Https_certfile = cfg.Section("network").Key("ssl_certfile").String()
	conf.Https_keyfile = cfg.Section("network").Key("ssl_keyfile").String()

	conf.ReadTimeout, err = cfg.Section("network").Key("read_timeout").Int()
	if err != nil {
		conf.ReadTimeout = 5
	}
	conf.WriteTimeout, err = cfg.Section("network").Key("write_timeout").Int()
	if err != nil {
		conf.WriteTimeout = 10
	}
	conf.Access_control_allow_origin = cfg.Section("network").Key("access_control_allow_origin").String()
	if conf.Access_control_allow_origin == "" {
		conf.Access_control_allow_origin = "*"
	}
	conf.Access_control_allow_headers = cfg.Section("network").Key("access_control_allow_headers").String()
	if conf.Access_control_allow_headers == "" {
		conf.Access_control_allow_headers = "*"
	}

	conf.Developer_mode = cfg.Section("debug").Key("developer_mode").MustBool()
	conf.LogStdout = cfg.Section("log").Key("stdout").MustBool()

	conf.UseApiDocument = cfg.Section("debug").Key("useApiDocument").MustBool()

	conf.Api_host_list = map[string]string{}

	return nil
}

func (conf *IamConfig) initEnvConfig() error {
	conf.Keycloak_client_id = os.Getenv("KEYCLOAK_CLIENT_ID")
	conf.Keycloak_client_secret = os.Getenv("KEYCLOAK_CLIENT_SECRET")
	conf.Keycloak_endpoint = os.Getenv("KEYCLOAK_ENDPOINT")

	if conf.Keycloak_client_id == "" || conf.Keycloak_client_secret == "" || conf.Keycloak_endpoint == "" {
		return errors.New("check [KEYCLOAK] config")
	}

	conf.Vault_token = os.Getenv("VAULT_TOKEN")
	conf.Vault_endpoint = os.Getenv("VAULT_ENDPOINT")

	if conf.Vault_token == "" || conf.Vault_endpoint == "" {
		return errors.New("check [VAULT] config")
	}

	db_server := os.Getenv("DB_SERVER_ADDR")
	db_name := os.Getenv("DB_NAME")
	db_port := os.Getenv("DB_SERVER_PORT")
	db_user_id := os.Getenv("DB_USER_ID")
	db_password := os.Getenv("DB_USER_PASSWORD")

	if db_server == "" || db_name == "" || db_port == "" || db_user_id == "" || db_password == "" {
		return errors.New("check [DATABASE] config")
	}

	conf.Db_connect_string = fmt.Sprintf("server=%s;database=%s;user id=%s;password=%s;port=%s", db_server, db_name, db_user_id, db_password, db_port)

	conf.Sales_Reqeuest_Url = os.Getenv("SALES_REQUEST_URL")

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

	conf.Developer_mode = false
	if os.Getenv("DEBUG_DEVELOPER_MODE") == "true" {
		conf.Developer_mode = true
	}

	conf.LogStdout = false
	if os.Getenv("LOG_STDOUT") == "true" {
		conf.LogStdout = true
	}

	conf.Api_host_list = map[string]string{}

	return nil
}
