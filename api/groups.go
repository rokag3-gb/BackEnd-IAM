package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetGroup(c *gin.Context) {
	c.String(http.StatusOK, "getgroup")
}

func CreateGroup(c *gin.Context) {
	c.String(http.StatusOK, "creategroup")
}

func DeleteGroup(c *gin.Context) {
	c.String(http.StatusOK, "deletegroup")
}

func UpdateGroup(c *gin.Context) {
	c.String(http.StatusOK, "updategroup")
}

func GetGroupMember(c *gin.Context) {
	c.String(http.StatusOK, "getgroupmember")
}
