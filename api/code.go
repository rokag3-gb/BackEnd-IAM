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

}

func CreateCodeItem(c *gin.Context) {

}

func UpdateCodeItem(c *gin.Context) {
}

func DeleteCodeItem(c *gin.Context) {
}
