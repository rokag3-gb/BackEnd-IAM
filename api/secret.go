package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Secret(c *gin.Context) {
	c.String(http.StatusOK, "secret")
}
