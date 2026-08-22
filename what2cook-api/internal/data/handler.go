package data

import (
	"errors"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"what2cook-api/internal/auth"
	"what2cook-api/internal/inventory"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type errorBody struct {
	Error string `json:"error"`
}

// Export handles GET /data/export?format=csv|xlsx
func (h *Handler) Export(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorBody{Error: "unauthorized"})
		return
	}

	format := c.Query("format")
	data, contentType, filename, err := h.svc.Export(userID, format)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, contentType, data)
}

// Import handles POST /data/import with multipart file upload.
func (h *Handler) Import(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorBody{Error: "unauthorized"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: "file is required"})
		return
	}
	if err := inventory.ValidateImportFileSize(file.Size); err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: err.Error()})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: "unable to read file"})
		return
	}
	defer src.Close()

	raw, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: "unable to read file"})
		return
	}

	format := strings.TrimSpace(c.PostForm("format"))
	if format == "" {
		format = detectFormat(file.Filename)
	}

	result, err := h.svc.Import(userID, raw, format)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func detectFormat(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".xlsx", ".xls":
		return "xlsx"
	default:
		return "csv"
	}
}

func (h *Handler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidFormat):
		c.JSON(http.StatusBadRequest, errorBody{Error: err.Error()})
	case errors.Is(err, ErrEmptyImport):
		c.JSON(http.StatusBadRequest, errorBody{Error: err.Error()})
	case errors.Is(err, auth.ErrNotFound):
		c.JSON(http.StatusNotFound, errorBody{Error: "not found"})
	default:
		log.Printf("data error: %v", err)
		c.JSON(http.StatusInternalServerError, errorBody{Error: "internal error"})
	}
}
