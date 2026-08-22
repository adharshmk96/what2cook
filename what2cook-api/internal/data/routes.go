package data

import (
	"github.com/gin-gonic/gin"

	"what2cook-api/internal/auth"
)

// RegisterRoutes mounts data export/import endpoints under /api/v1.
func RegisterRoutes(api *gin.RouterGroup, h *Handler, authSvc *auth.Service) {
	group := api.Group("/data")
	group.Use(auth.Middleware(authSvc))
	{
		group.GET("/export", h.Export)
		group.POST("/import", h.Import)
	}
}
