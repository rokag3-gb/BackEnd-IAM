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
	value, err := ioutil.ReadAll(c.Request.Body)
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
	path := fmt.Sprintf("%s/metadata", groupName)

	data, err := clients.VaultClient().Logical().List(path)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	list := data.Data["keys"]
	if list == nil {
		c.Status(http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, list)
}

func GetSecret(c *gin.Context) {
	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/data/%s", groupName, secretName)

	data, err := clients.VaultClient().Logical().Read(path)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	if data.Data == nil {
		c.Status(http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, data.Data)
}

func MargeSecret(c *gin.Context) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/data/%s", groupName, secretName)

	data, err := clients.VaultClient().Logical().WriteBytes(path, value)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	if data.Data == nil {
		c.Status(http.StatusBadRequest)
		return
	}

	c.Status(http.StatusCreated)
}

func GetMetadataSecret(c *gin.Context) {
	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/metadata/%s", groupName, secretName)

	data, err := clients.VaultClient().Logical().Read(path)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	if data.Data == nil {
		c.Status(http.StatusBadRequest)
		return
	}

	_, cas_required := data.Data["cas_required"]
	if cas_required {
		delete(data.Data, "cas_required")
	}

	_, custom_metadata := data.Data["custom_metadata"]
	if custom_metadata {
		delete(data.Data, "custom_metadata")
	}

	_, delete_version_after := data.Data["delete_version_after"]
	if delete_version_after {
		delete(data.Data, "delete_version_after")
	}

	c.JSON(http.StatusOK, data.Data)
}

func DeleteSecret(c *gin.Context) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	if value == nil {
		c.Status(http.StatusBadRequest)
		return
	}

	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/delete/%s", groupName, secretName)

	_, err = clients.VaultClient().Logical().WriteBytes(path, value)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func UndeleteSecret(c *gin.Context) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	if value == nil {
		c.Status(http.StatusBadRequest)
		return
	}

	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/undelete/%s", groupName, secretName)

	_, err = clients.VaultClient().Logical().WriteBytes(path, value)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func DestroySecret(c *gin.Context) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	if value == nil {
		c.Status(http.StatusBadRequest)
		return
	}

	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/destroy/%s", groupName, secretName)

	_, err = clients.VaultClient().Logical().WriteBytes(path, value)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func DeleteSecretMetadata(c *gin.Context) {
	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/metadata/%s", groupName, secretName)

	_, err := clients.VaultClient().Logical().Delete(path)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}
