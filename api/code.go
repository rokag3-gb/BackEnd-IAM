package api

import (
	"iam/iamdb"
	"iam/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetCodeItem(c *gin.Context) {
	codeKey := c.Param("codeKey")
	if codeKey == "root" {
		codeKey = "000-000"
	}
	codeItem := iamdb.GetCodeByCodeKey(codeKey)
	if codeItem == nil {
		c.Status(http.StatusNotFound)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, codeItem)
}

func GetCodeChilds(c *gin.Context) {
	kindCodeOrCodeKey := c.Param("codeKey")
	if kindCodeOrCodeKey == "root" {
		kindCodeOrCodeKey = "000"
	} else if len(kindCodeOrCodeKey) == 7 && kindCodeOrCodeKey[3] == '-' {
		kindCodeOrCodeKey = kindCodeOrCodeKey[4:]
	} else if len(kindCodeOrCodeKey) != 3 {
		c.String(http.StatusBadRequest, "Only root, 3-digit kindCode or 7-digit codeKey is allowed")
		c.Abort()
		return
	}

	codeItems, err := iamdb.GetCodeChildsByKindCode(kindCodeOrCodeKey)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, codeItems)
}

func CreateCodeItem(c *gin.Context) {
	var json models.Code
	if err := c.ShouldBindJSON(&json); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	codes, err := iamdb.GetCodeListByCode(json.KindCode)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if len(codes) <= 0 {
		c.String(http.StatusBadRequest, "Parent KindCode is not found")
		return
	}
	err = iamdb.CreateCodeItem(json)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusCreated)
}

func UpdateCodeItem(c *gin.Context) {
	codeKey := c.Param("codeKey")
	var json models.Code
	if err := c.ShouldBindJSON(&json); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	codes, err := iamdb.GetCodeListByCode(json.KindCode)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if len(codes) <= 0 {
		c.String(http.StatusBadRequest, "Parent KindCode is not found")
		return
	}
	err = iamdb.UpdateCodeItemByCodeKey(codeKey, json)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusCreated)
}

func DeleteCodeItem(c *gin.Context) {
	codeKey := c.Param("codeKey")
	err := iamdb.DeleteCodeItemByCodeKey(codeKey)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusOK)
}
