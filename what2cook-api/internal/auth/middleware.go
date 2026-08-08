package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ctxUserIDKey    = "authUserID"
	ctxSessionIDKey = "authSessionID"
)

// Middleware validates Bearer tokens and applies sliding session refresh.
func Middleware(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorBody{Error: "unauthorized"})
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorBody{Error: "unauthorized"})
			return
		}

		userID, sessionID, err := svc.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorBody{Error: "unauthorized"})
			return
		}

		c.Set(ctxUserIDKey, userID)
		c.Set(ctxSessionIDKey, sessionID)
		c.Next()
	}
}

// UserIDFromContext returns the authenticated user id.
func UserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(ctxUserIDKey)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

// SessionIDFromContext returns the authenticated session id.
func SessionIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(ctxSessionIDKey)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}
