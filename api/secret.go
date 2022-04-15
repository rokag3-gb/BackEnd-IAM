package api

import (
	"fmt"
	"iam/clients"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetSecretGroup(c *gin.Context) {
	data, err := clients.VaultClient().Logical().Read("sys/mounts")
	if err != nil {
		fmt.Println(err)
	}

	var arr []map[string]interface{}

	for k, v := range data.Data {
		group := v.(map[string]interface{})

		if group["type"].(string) != "kv" {
			continue
		}

		m := make(map[string]interface{})

		if strings.HasSuffix(k, "/") {
			m["name"] = k[:len(k)-1]
		} else {
			m["name"] = k
		}
		m["id"] = group["uuid"].(string)
		m["description"] = group["description"].(string)

		arr = append(arr, m)
	}

	c.JSON(http.StatusOK, arr)
}

func CreateSecretGroup(c *gin.Context) {

}

func Secret(c *gin.Context) {
	c.String(http.StatusOK, "secret")
}
