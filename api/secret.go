package api

import (
	"encoding/json"
	"fmt"
	"iam/clients"
	"iam/common"
	"iam/iamdb"
	"iam/models"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// token godoc
// @Summary 비밀 그룹 생성
// @Tags Secret
// @Produce  json
// @Router /secret [post]
// @Param Body body []models.SecretGroupInput true "body"
// @Success 201
// @Failure 400
// @Failure 500
func CreateSecretGroup(c *gin.Context) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	var sg *models.SecretGroupItem
	json.Unmarshal([]byte(value), &sg)

	if sg.Name == "" {
		common.ErrorProcess(c, err, http.StatusBadRequest, "required 'body'")
		return
	}

	db, err := iamdb.DBClient()
	defer db.Close()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.CreateSecretGroupTx(tx, sg.Name, c.GetString("username"))
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	authname := sg.Name + "_MANAGER"
	authId := uuid.New()

	rolename := sg.Name + "_Manager"
	roleId := uuid.New()

	err = iamdb.CreateAuthIdTx(tx, authId.String(), authname, "/iam/secret/"+sg.Name+"/*", "ALL", c.GetString("username"))
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.CreateRolesIdTx(tx, roleId.String(), rolename, false, c.GetString("username"))
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.AssignRoleAuthTx(tx, roleId.String(), authId.String(), c.GetString("username"))
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if sg.RoleId != nil {
		for _, role := range *sg.RoleId {
			err = iamdb.AssignRoleAuthTx(tx, role, authId.String(), c.GetString("username"))
			if err != nil {
				tx.Rollback()
				common.ErrorProcess(c, err, http.StatusInternalServerError, "")
				return
			}
		}
	}

	if sg.UserId != nil {
		for _, user := range *sg.UserId {
			err = iamdb.AssignUserRoleTx(tx, user, roleId.String(), c.GetString("username"))
			if err != nil {
				tx.Rollback()
				common.ErrorProcess(c, err, http.StatusInternalServerError, "")
				return
			}
		}
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
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusCreated)
}

// token godoc
// @Summary 비밀 그룹 삭제
// @Tags Secret
// @Produce  json
// @Router /secret/{SecretGroupName} [delete]
// @Param secretGroupName path string true "Secret Group Name"
// @Param Body body []models.SecretGroupInput true "body"
// @Success 204
// @Failure 400
// @Failure 500
func DeleteSecretGroup(c *gin.Context) {
	groupName := c.Param("groupName")

	db, err := iamdb.DBClient()
	defer db.Close()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	authname := groupName + "_MANAGER"
	rolename := groupName + "_Manager"

	err = iamdb.DeleteUserRoleByRoleNameTx(tx, rolename)
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.DeleteRolesAuthByAuthNameTx(tx, authname)
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.DeleteAuthNameTx(tx, authname)
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.DeleteRolesNameTx(tx, rolename)
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.DeleteSecretBySecretGroupTx(tx, groupName)
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.DeleteSecretGroupTx(tx, groupName)
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	path := fmt.Sprintf("sys/mounts/%s", groupName)
	_, err = clients.VaultClient().Logical().Delete(path)
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Summary 비밀 목록 조회
// @Tags Secret
// @Produce  json
// @Router /secret/{SecretGroupName} [get]
// @Param secretGroupName path string true "Secret Group Name"
// @Success 200 {object} []models.SecretItem
// @Failure 500
func GetSecretList(c *gin.Context) {
	groupName := c.Param("groupName")

	if err := CheckGroupName(groupName); err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	path := fmt.Sprintf("%s/metadata", groupName)

	data, err := clients.VaultClient().Logical().List(path)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	arr := make([]models.SecretItem, 0)

	if data == nil || data.Data == nil || data.Data["keys"] == nil {
		c.JSON(http.StatusOK, arr)
		return
	}

	secrets, err := iamdb.GetSecret(groupName)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	tmp := data.Data["keys"].([]interface{})
	vItem := make([]string, len(tmp))
	for i, v := range tmp {
		vItem[i] = v.(string)
	}

	for _, item := range vItem {
		v := secrets[item]
		v.Name = item
		arr = append(arr, v)
	}

	c.JSON(http.StatusOK, arr)
}

// token godoc
// @Summary 비밀 그룹 수정
// @Tags Secret
// @Produce  json
// @Router /secret/{SecretGroupName} [put]
// @Param secretGroupName path string true "Secret Group Name"
// @Success 200 {object} []models.SecretGroupUpdate
// @Success 204
// @Failure 400
// @Failure 500
func UpdateSecretGroup(c *gin.Context) {
	groupName := c.Param("groupName")
	authorityMessage := ""
	roleMessage := ""

	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	var sg *models.SecretGroupItem
	json.Unmarshal([]byte(value), &sg)

	if sg == nil {
		common.ErrorProcess(c, nil, http.StatusBadRequest, "required 'body'")
		return
	}

	db, err := iamdb.DBClient()
	defer db.Close()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	authName := groupName + "_MANAGER"
	roleName := groupName + "_Manager"

	authId, err := iamdb.GetAuthIdByName(authName)
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	roleId, err := iamdb.GetRoleIdByName(roleName)
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if authId != "" {
		if sg.RoleId != nil {
			err = iamdb.DeleteRolesAuthByAuthIdTx(tx, authId)
			if err != nil {
				tx.Rollback()
				common.ErrorProcess(c, err, http.StatusInternalServerError, "")
				return
			}

			for _, role := range *sg.RoleId {
				err = iamdb.AssignRoleAuthTx(tx, role, authId, c.GetString("username"))
				if err != nil {
					tx.Rollback()
					common.ErrorProcess(c, err, http.StatusInternalServerError, "")
					return
				}
			}
		}
	} else {
		authorityMessage += fmt.Sprint("Authoriy[" + authName + "] does not exist.")
	}

	if roleId != "" {
		if sg.UserId != nil {
			err = iamdb.DeleteUserRoleByRoleIdTx(tx, roleId)
			if err != nil {
				tx.Rollback()
				common.ErrorProcess(c, err, http.StatusInternalServerError, "")
				return
			}

			for _, user := range *sg.UserId {
				err = iamdb.AssignUserRoleTx(tx, user, roleId, c.GetString("username"))
				if err != nil {
					tx.Rollback()
					common.ErrorProcess(c, err, http.StatusInternalServerError, "")
					return
				}
			}
		}
	} else {
		roleMessage += fmt.Sprint("Role[" + roleName + "] does not exist.")
	}

	path := fmt.Sprintf("sys/mounts/%s/tune", groupName)

	_, err = clients.VaultClient().Logical().Write(path, map[string]interface{}{
		"description": sg.Description,
	})
	if err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if roleMessage != "" || authorityMessage != "" {
		m := make(map[string]interface{})
		if roleMessage != "" {
			m["role"] = roleMessage
		}
		if authorityMessage != "" {
			m["authority"] = authorityMessage
		}
		c.JSON(http.StatusOK, m)
	} else {
		c.Status(http.StatusNoContent)
	}
}

// token godoc
// @Summary 비밀 그룹 상세 정보 조회
// @Tags Secret
// @Produce  json
// @Router /secret/{SecretGroupName}/data [get]
// @Param secretGroupName path string true "Secret Group Name"
// @Success 200 {object} models.SecretGroupResponse
// @Failure 400
// @Failure 500
func GetSecretGroupMetadata(c *gin.Context) {
	groupName := c.Param("groupName")

	if err := CheckGroupName(groupName); err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	path := fmt.Sprintf("/sys/mounts/%s", groupName)

	data, err := clients.VaultClient().Logical().Read(path)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	arr := make([]models.SecretItem, 0)

	if data == nil || data.Data == nil || data.Data["description"] == nil {
		c.JSON(http.StatusOK, arr)
		return
	}

	secretGroup, err := iamdb.GetSecretGroupMetadata(groupName)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	secretGroup.Description = data.Data["description"].(string)

	c.JSON(http.StatusOK, secretGroup)
}

// token godoc
// @Summary 비밀 내용 조회
// @Tags Secret
// @Produce  json
// @Router /secret/{SecretGroupName}/data/{Secretname} [get]
// @Param secretGroupName path string true "Secret Group Name"
// @Param secretName path string true "Secret Name"
// @Success 200 {object} models.SecretData
// @Failure 404
// @Failure 500
func GetSecret(c *gin.Context) {
	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/data/%s", groupName, secretName)

	data, err := clients.VaultClient().Logical().Read(path)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if data == nil || data.Data == nil {
		common.ErrorProcess(c, err, http.StatusNotFound, "Data Not Found")
		return
	}

	secret, err := iamdb.GetSecretByName(groupName, secretName)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	data.Data["url"] = secret.Url
	data.Data["createDate"] = secret.CreateDate
	data.Data["creator"] = secret.Creator
	data.Data["modifyDate"] = secret.ModifyDate
	data.Data["modifier"] = secret.Modifier

	c.JSON(http.StatusOK, data.Data)
}

// token godoc
// @Summary 비밀 생성 / 수정
// @Tags Secret
// @Produce  json
// @Router /secret/{SecretGroupName}/data/{Secretname} [post]
// @Param secretGroupName path string true "Secret Group Name"
// @Param secretName path string true "Secret Name"
// @Param Body body models.SecretInput true "body"
// @Success 201
// @Failure 400
// @Failure 500
func MargeSecret(c *gin.Context) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	var url models.SecretURL
	json.Unmarshal([]byte(value), &url)

	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/data/%s", groupName, secretName)

	data, err := clients.VaultClient().Logical().WriteBytes(path, value)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if data == nil || data.Data == nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "return data is null")
		return
	}

	err = iamdb.MergeSecret(secretName, groupName, url.URL, c.GetString("username"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusCreated)
}

// token godoc
// @Summary 비밀 상세 정보 조회
// @Tags Secret
// @Produce  json
// @Router /secret/{SecretGroupName}/metadata/{Secretname} [get]
// @Param secretGroupName path string true "Secret Group Name"
// @Param secretName path string true "Secret Name"
// @Success 200 {object} models.SecretMetadata
// @Failure 404
// @Failure 500
func GetMetadataSecret(c *gin.Context) {
	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/metadata/%s", groupName, secretName)

	data, err := clients.VaultClient().Logical().Read(path)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	if data == nil || data.Data == nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "vault return data is null")
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

// token godoc
// @Summary 비밀특정버전삭제(복구가능)
// @Tags Secret
// @Produce  json
// @Router /secret/{SecretGroupName}/delete/{Secretname} [POST]
// @Param secretGroupName path string true "Secret Group Name"
// @Param secretName path string true "Secret Name"
// @Param Body body models.SecretVersion true "body"
// @Success 204
// @Failure 400
// @Failure 500
func DeleteSecret(c *gin.Context) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	if value == nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "required 'body'")
		return
	}

	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/delete/%s", groupName, secretName)

	_, err = clients.VaultClient().Logical().WriteBytes(path, value)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Summary 비밀 복구
// @Tags Secret
// @Produce  json
// @Router /secret/{SecretGroupName}/undelete/{Secretname} [POST]
// @Param secretGroupName path string true "Secret Group Name"
// @Param secretName path string true "Secret Name"
// @Param Body body models.SecretVersion true "body"
// @Success 204
// @Failure 400
// @Failure 500
func UndeleteSecret(c *gin.Context) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	if value == nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "required 'body'")
		return
	}

	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/undelete/%s", groupName, secretName)

	_, err = clients.VaultClient().Logical().WriteBytes(path, value)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Summary 비밀특정버전삭제(복구불가)
// @Tags Secret
// @Produce  json
// @Router /secret/{SecretGroupName}/destroy/{Secretname} [POST]
// @Param secretGroupName path string true "Secret Group Name"
// @Param secretName path string true "Secret Name"
// @Param Body body models.SecretVersion true "body"
// @Success 204
// @Failure 400
// @Failure 500
func DestroySecret(c *gin.Context) {
	value, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "")
		return
	}

	if value == nil {
		common.ErrorProcess(c, err, http.StatusBadRequest, "required 'body'")
		return
	}

	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/destroy/%s", groupName, secretName)

	_, err = clients.VaultClient().Logical().WriteBytes(path, value)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// token godoc
// @Summary 비밀 삭제
// @Tags Secret
// @Produce  json
// @Router /secret/{SecretGroupName}/metadata/{Secretname} [delete]
// @Param secretGroupName path string true "Secret Group Name"
// @Param secretName path string true "Secret Name"
// @Success 204
// @Failure 400
// @Failure 500
func DeleteSecretMetadata(c *gin.Context) {
	groupName := c.Param("groupName")
	secretName := c.Param("secretName")
	path := fmt.Sprintf("%s/metadata/%s", groupName, secretName)

	_, err := clients.VaultClient().Logical().Delete(path)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	err = iamdb.DeleteSecret(secretName, groupName)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	c.Status(http.StatusNoContent)
}

func CheckGroupName(groupName string) error {
	path := fmt.Sprintf("sys/mounts/%s", groupName)
	_, err := clients.VaultClient().Logical().Read(path)
	if err != nil {
		return err
	}

	return nil
}

// token godoc
// @Summary 비밀 그룹 목록 조회
// @Tags Secret
// @Produce  json
// @Router /secret [get]
// @Success 200 {object} []models.SecretGroupItem
// @Failure 500
func GetAllSecretList(c *gin.Context) {
	data, err := clients.VaultClient().Logical().Read("sys/mounts")
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	arr := make([]models.SecretGroupItem, 0)

	for k, v := range data.Data {
		group := v.(map[string]interface{})

		if group["type"].(string) != "kv" {
			continue
		}

		name := ""
		if strings.HasSuffix(k, "/") {
			name = k[:len(k)-1]
		} else {
			name = k
		}

		var m models.SecretGroupItem

		m.Name = name
		m.Description = group["description"].(string)

		arr = append(arr, m)
	}

	secretGroup, err := iamdb.GetSecretGroup(arr, c.GetString("username"))
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}
	if len(secretGroup) == 0 {
		c.JSON(http.StatusOK, secretGroup)
		return
	}

	secrets, err := iamdb.GetAllSecret(secretGroup)
	if err != nil {
		common.ErrorProcess(c, err, http.StatusInternalServerError, "")
		return
	}

	for i := range secretGroup {
		tmp := make([]models.SecretItem, 0)
		secretGroup[i].Secrets = &tmp
	}

	for _, s := range secrets {
		for _, sg := range secretGroup {
			if sg.Name == s.SecretGroup {
				*sg.Secrets = append(*sg.Secrets, s)
				break
			}
		}
	}

	c.JSON(http.StatusOK, secretGroup)
}
