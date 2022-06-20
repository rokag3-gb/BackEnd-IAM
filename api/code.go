package api

import (
	"iam/iamdb"
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

}

func UpdateCodeItem(c *gin.Context) {
}

func DeleteCodeItem(c *gin.Context) {
}
