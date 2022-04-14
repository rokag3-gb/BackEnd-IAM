package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Users(c *gin.Context) {
	c.String(http.StatusOK, "users")
}
