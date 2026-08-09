package inventory

import (
	"github.com/gin-gonic/gin"

	"what2cook-api/internal/auth"
)

// RegisterRoutes mounts inventory endpoints under the given /api/v1 group.
func RegisterRoutes(api *gin.RouterGroup, h *Handler, authSvc *auth.Service) {
	group := api.Group("/inventories")
	group.Use(auth.Middleware(authSvc))
	{
		group.GET("", h.List)
		group.POST("", h.Create)
		group.GET("/:id", h.Get)
		group.PATCH("/:id", h.Update)
		group.DELETE("/:id", h.Delete)
		group.POST("/:id/items", h.CreateItem)
		group.PATCH("/:id/items/:itemId", h.UpdateItem)
		group.DELETE("/:id/items/:itemId", h.DeleteItem)
	}
}
