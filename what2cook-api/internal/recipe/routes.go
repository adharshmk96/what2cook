package recipe

import (
	"github.com/gin-gonic/gin"

	"what2cook-api/internal/auth"
)

// RegisterRoutes mounts recipe endpoints under the given /api/v1 group.
func RegisterRoutes(api *gin.RouterGroup, h *Handler, authSvc *auth.Service) {
	group := api.Group("/recipes")
	group.Use(auth.Middleware(authSvc))
	{
		group.POST("/generate", h.Generate)
	}
}
