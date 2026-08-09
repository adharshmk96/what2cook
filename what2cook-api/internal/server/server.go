package server

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"what2cook-api/internal/auth"
	"what2cook-api/internal/config"
	"what2cook-api/internal/inventory"
	"what2cook-api/internal/mail"
	"what2cook-api/internal/recipe"
	"what2cook-api/web"
)

// Server wraps the Gin engine.
type Server struct {
	engine *gin.Engine
}

// New builds the HTTP server with API routes and embedded UI.
func New(cfg *config.Config, gdb *gorm.DB, mailer *mail.Mailer) (*Server, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	repo := auth.NewRepository(gdb)
	authSvc := auth.NewService(repo, mailer, cfg.Auth.TokenSecret, cfg.Auth.SessionTTL, cfg.Auth.ResetTTL, cfg.Auth.VerifyTTL)
	authHandler := auth.NewHandler(authSvc)

	recipeSvc := recipe.NewService()
	recipeHandler := recipe.NewHandler(recipeSvc)

	invRepo := inventory.NewRepository(gdb)
	invSvc := inventory.NewService(invRepo)
	invHandler := inventory.NewHandler(invSvc)

	api := r.Group("/api/v1")
	auth.RegisterRoutes(api, authHandler, authSvc)
	recipe.RegisterRoutes(api, recipeHandler, authSvc)
	inventory.RegisterRoutes(api, invHandler, authSvc)

	if err := mountUI(r); err != nil {
		return nil, err
	}

	return &Server{engine: r}, nil
}

// Run starts listening on addr.
func (s *Server) Run(addr string) error {
	return s.engine.Run(addr)
}

func mountUI(r *gin.Engine) error {
	uiFS, err := web.UI()
	if err != nil {
		return err
	}

	if _, err := fs.Stat(uiFS, "index.html"); err != nil {
		return fmt.Errorf("embedded UI missing index.html — run `make build` from repo root: %w", err)
	}

	r.GET("/app", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/app/")
	})
	// Static assets + SPA fallback under /app; /api/v1 is registered separately.
	// Avoid http.FileServer: it redirects */index.html → ./ which loops under /app.
	r.GET("/app/*filepath", func(c *gin.Context) {
		rel := strings.TrimPrefix(c.Param("filepath"), "/")
		if rel == "" || strings.HasSuffix(rel, "/") || !uiFileExists(uiFS, rel) {
			rel = "index.html"
		}
		if err := serveUIFile(c, uiFS, rel); err != nil {
			log.Printf("serve UI %q: %v", rel, err)
			c.Status(http.StatusInternalServerError)
		}
	})

	log.Printf("UI mount ready at /app (embed: web/dist)")
	return nil
}

func uiFileExists(fsys fs.FS, name string) bool {
	f, err := fsys.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()
	st, err := f.Stat()
	return err == nil && !st.IsDir()
}

func serveUIFile(c *gin.Context, fsys fs.FS, name string) error {
	f, err := fsys.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return err
	}
	if st.IsDir() {
		return fmt.Errorf("%s is a directory", name)
	}

	ctype := mime.TypeByExtension(path.Ext(name))
	if ctype == "" {
		ctype = "application/octet-stream"
	}
	c.Header("Content-Type", ctype)

	if rs, ok := f.(io.ReadSeeker); ok {
		http.ServeContent(c.Writer, c.Request, name, st.ModTime(), rs)
		return nil
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	http.ServeContent(c.Writer, c.Request, name, st.ModTime(), bytes.NewReader(data))
	return nil
}
