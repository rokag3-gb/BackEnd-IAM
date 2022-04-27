package api

import (
	"encoding/json"
	"fmt"
	"iam/clients"
	"iam/models"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetSecretGroup(c *gin.Context) {
	data, err := clients.VaultClient().Logical().Read("sys/mounts")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		c.Abort()
		return
	}

	var arr []models.SecretGroup

	for k, v := range data.Data {
		group := v.(map[string]interface{})

		if group["type"].(string) != "kv" {
			continue
		}

		var m models.SecretGroup

		if strings.HasSuffix(k, "/") {
			m.Name = k[:len(k)-1]
		} else {
			m.Name = k
		}
		m.ID = group["uuid"].(string)
		m.Description = group["description"].(string)

		arr = append(arr, m)
	}

	c.JSON(http.StatusOK, arr)
}

func CreateSecretGroup(c *gin.Context) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	var sg *models.SecretGroup
	json.Unmarshal([]byte(value), &sg)

	if sg.Name == "" {
		c.String(http.StatusBadRequest, "required 'body'")
		c.Abort()
		return
	}

	path := fmt.Sprintf("sys/mounts/%s", sg.Name)

	_, err = clients.VaultClient().Logical().Write(path, map[string]interface{}{
		"description": sg.Description,
		"type":        "kv",
		"options": map[string]interface{}{
			"version": "2",
		},
	})
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		c.Abort()
		return
	}

	c.Status(http.StatusCreated)
}

func DeleteSecretGroup(c *gin.Context) {
	groupName := c.Param("groupName")
	path := fmt.Sprintf("sys/mounts/%s", groupName)

	_, err := clients.VaultClient().Logical().Delete(path)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}

func GetSecretList(c *gin.Context) {
	groupName := c.Param("groupName")
	path := fmt.Sprintf("%s/metadata", groupName)

	data, err := clients.VaultClient().Logical().List(path)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		c.Abort()
		return
	}
	if data == nil || data.Data == nil || data.Data["keys"] == nil {
		c.JSON(http.StatusOK, make([]string, 0))
		return
	}

	c.JSON(http.StatusOK, data.Data["keys"])
}

func GetSecret(c *gin.Context) {
	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/data/%s", groupName, secretName)

	data, err := clients.VaultClient().Logical().Read(path)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		c.Abort()
		return
	}

	if data == nil || data.Data == nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, data.Data)
}

func MargeSecret(c *gin.Context) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/data/%s", groupName, secretName)

	data, err := clients.VaultClient().Logical().WriteBytes(path, value)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		c.Abort()
		return
	}

	if data == nil || data.Data == nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
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
		c.String(http.StatusInternalServerError, err.Error())
		c.Abort()
		return
	}

	if data == nil || data.Data == nil {
		c.Status(http.StatusInternalServerError)
		c.Abort()
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
		c.String(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if value == nil {
		c.String(http.StatusBadRequest, "required 'body'")
		c.Abort()
		return
	}

	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/delete/%s", groupName, secretName)

	_, err = clients.VaultClient().Logical().WriteBytes(path, value)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}

func UndeleteSecret(c *gin.Context) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if value == nil {
		c.String(http.StatusBadRequest, "required 'body'")
		c.Abort()
		return
	}

	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/undelete/%s", groupName, secretName)

	_, err = clients.VaultClient().Logical().WriteBytes(path, value)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}

func DestroySecret(c *gin.Context) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if value == nil {
		c.String(http.StatusBadRequest, "required 'body'")
		c.Abort()
		return
	}

	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/destroy/%s", groupName, secretName)

	_, err = clients.VaultClient().Logical().WriteBytes(path, value)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		c.Abort()
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
		c.String(http.StatusInternalServerError, err.Error())
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}
