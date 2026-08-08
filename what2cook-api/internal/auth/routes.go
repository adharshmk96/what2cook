package auth

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts auth endpoints under the given /api/v1 group.
func RegisterRoutes(api *gin.RouterGroup, h *Handler, svc *Service) {
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.POST("/forgot-password", h.ForgotPassword)
		authGroup.POST("/reset-password", h.ResetPassword)

		protected := authGroup.Group("")
		protected.Use(Middleware(svc))
		{
			protected.POST("/logout", h.Logout)
			protected.POST("/change-password", h.ChangePassword)
			protected.GET("/me", h.Me)
		}
	}
}
