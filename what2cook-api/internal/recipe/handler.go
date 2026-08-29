package recipe

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"what2cook-api/internal/auth"
)

// Handler exposes HTTP handlers for the recipe slice.
type Handler struct {
	svc      *Service
	savedSvc *SavedService
}

// NewHandler creates a recipe handler.
func NewHandler(svc *Service, savedSvc *SavedService) *Handler {
	return &Handler{svc: svc, savedSvc: savedSvc}
}

type errorBody struct {
	Error string `json:"error"`
}

type savedRecipeListBody struct {
	Recipes []SavedRecipe `json:"recipes"`
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

// List handles GET /recipes.
func (h *Handler) List(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorBody{Error: "unauthorized"})
		return
	}
	list, err := h.savedSvc.List(userID)
	if err != nil {
		log.Printf("saved recipe list failed: %v", err)
		c.JSON(http.StatusInternalServerError, errorBody{Error: "internal error"})
		return
	}
	if list == nil {
		list = []SavedRecipe{}
	}
	c.JSON(http.StatusOK, savedRecipeListBody{Recipes: list})
}

// Create handles POST /recipes.
func (h *Handler) Create(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorBody{Error: "unauthorized"})
		return
	}
	var req SaveRecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: "invalid JSON body"})
		return
	}
	if msg := ValidateSaveRecipe(&req); msg != "" {
		c.JSON(http.StatusBadRequest, errorBody{Error: msg})
		return
	}
	rec, err := h.savedSvc.Create(userID, &req)
	if err != nil {
		log.Printf("saved recipe create failed: %v", err)
		c.JSON(http.StatusInternalServerError, errorBody{Error: "internal error"})
		return
	}
	c.JSON(http.StatusCreated, rec)
}

// Get handles GET /recipes/:id.
func (h *Handler) Get(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorBody{Error: "unauthorized"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: "invalid recipe id"})
		return
	}
	rec, err := h.savedSvc.Get(userID, id)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, rec)
}

// Update handles PUT /recipes/:id.
func (h *Handler) Update(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorBody{Error: "unauthorized"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: "invalid recipe id"})
		return
	}
	var req SaveRecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: "invalid JSON body"})
		return
	}
	if msg := ValidateSaveRecipe(&req); msg != "" {
		c.JSON(http.StatusBadRequest, errorBody{Error: msg})
		return
	}
	rec, err := h.savedSvc.Update(userID, id, &req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, rec)
}

// Delete handles DELETE /recipes/:id.
func (h *Handler) Delete(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorBody{Error: "unauthorized"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: "invalid recipe id"})
		return
	}
	if err := h.savedSvc.Delete(userID, id); err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, errorBody{Error: "not found"})
	default:
		log.Printf("saved recipe error: %v", err)
		c.JSON(http.StatusInternalServerError, errorBody{Error: "internal error"})
	}
}
