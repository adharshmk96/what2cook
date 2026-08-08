package recipe

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler exposes HTTP handlers for the recipe slice.
type Handler struct {
	svc *Service
}

// NewHandler creates a recipe handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type errorBody struct {
	Error string `json:"error"`
}

// Generate handles POST /recipes/generate.
func (h *Handler) Generate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: "invalid JSON body"})
		return
	}
	if msg := ValidateGenerate(&req); msg != "" {
		c.JSON(http.StatusBadRequest, errorBody{Error: msg})
		return
	}

	result, err := h.svc.Generate(req.Ingredients)
	if err != nil {
		log.Printf("recipe generate failed: %v", err)
		c.JSON(http.StatusInternalServerError, errorBody{Error: "internal error"})
		return
	}
	c.JSON(http.StatusOK, result)
}
