package api

import (
	"encoding/json"
	"fmt"
	"iam/clients"
	"io/ioutil"
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
	body := c.Request.Body
	value, err := ioutil.ReadAll(body)
	if err != nil {
		fmt.Println(err.Error())
	}

	var data map[string]interface{}
	json.Unmarshal([]byte(value), &data)

	if data["name"] == nil {
		c.Status(http.StatusBadRequest)
		return
	}

	description := ""
	if data["description"] != nil {
		description = fmt.Sprintf("%s", data["description"])
	}

	path := fmt.Sprintf("sys/mounts/%s", data["name"])

	_, err = clients.VaultClient().Logical().Write(path, map[string]interface{}{
		"description": description,
		"type":        "kv",
		"options": map[string]interface{}{
			"version": "2",
		},
	})
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	c.Status(http.StatusCreated)
}

func DeleteSecretGroup(c *gin.Context) {
	groupName := c.Param("groupName")
	path := fmt.Sprintf("sys/mounts/%s", groupName)

	_, err := clients.VaultClient().Logical().Delete(path)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func GetSecretList(c *gin.Context) {
	groupName := c.Param("groupName")
	path := fmt.Sprintf("sys/mounts/%s", groupName)

	_, err := clients.VaultClient().Logical().Delete(path)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func Secret(c *gin.Context) {
	c.String(http.StatusOK, "secret")
}
