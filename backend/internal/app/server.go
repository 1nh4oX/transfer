package app

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"transfer/backend/internal/config"
	"transfer/backend/internal/handler"
	"transfer/backend/internal/middleware"
	"transfer/backend/internal/repo"
	"transfer/backend/internal/service"
)

type Server struct {
	engine *gin.Engine
	addr   string
}

func NewServer() (*Server, error) {
	cfg := config.Load()

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	healthHandler := handler.NewHealthHandler()
	authService := service.NewAuthService(cfg)
	authHandler := handler.NewAuthHandler(authService)

	db, err := repo.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	shareRepo := repo.NewPostgresShareRepository(db)
	fileRepo := repo.NewPostgresFileRepository(db)
	noteRepo := repo.NewPostgresNoteRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := shareRepo.InitSchema(ctx); err != nil {
		return nil, err
	}
	if err := fileRepo.InitSchema(ctx); err != nil {
		return nil, err
	}
	if err := noteRepo.InitSchema(ctx); err != nil {
		return nil, err
	}

	fileService, err := service.NewFileService(fileRepo, cfg.UploadDir)
	if err != nil {
		return nil, err
	}
	fileHandler := handler.NewFileHandler(fileService)
	noteService := service.NewNoteService(noteRepo)
	noteHandler := handler.NewNoteHandler(noteService)
	shareService := service.NewShareService(cfg, shareRepo, fileRepo)
	shareHandler := handler.NewShareHandler(shareService, fileHandler)

	engine.GET("/healthz", healthHandler.Healthz)

	api := engine.Group("/api")
	{
		api.POST("/auth/login", authHandler.Login)

		// Public share routes — no JWT required
		api.GET("/s/:token", shareHandler.GetShareInfo)
		api.GET("/s/:token/download", shareHandler.DownloadShare)

		authed := api.Group("", middleware.JWTAuth(cfg.JWTSecret))
		{
			authed.GET("/tree", fileHandler.NotImplemented)
			authed.POST("/folders", fileHandler.NotImplemented)
			authed.PATCH("/folders/:folderId/move", fileHandler.NotImplemented)
			authed.GET("/files", fileHandler.List)
			authed.POST("/files/upload", fileHandler.Upload)
			authed.GET("/files/:fileId/download", fileHandler.DownloadByID)
			authed.PATCH("/files/:fileId/move", fileHandler.NotImplemented)
			authed.GET("/files/:fileId/preview", fileHandler.NotImplemented)
			authed.GET("/files/:fileId/content", fileHandler.NotImplemented)
			authed.DELETE("/files/:fileId", fileHandler.DeleteByID)

			authed.GET("/notes", noteHandler.ListNotes)
			authed.POST("/notes", noteHandler.CreateNote)

			// Share create — requires auth
			authed.POST("/shares", shareHandler.CreateShare)
		}
	}

	return &Server{engine: engine, addr: cfg.Addr()}, nil
}

func (s *Server) Run() error {
	return s.engine.Run(s.addr)
}
