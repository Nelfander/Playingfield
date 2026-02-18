package app

import (
	"context"
	"log/slog"
	stdhttp "net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/nelfander/Playingfield/internal/domain/messages"
	"github.com/nelfander/Playingfield/internal/domain/projects"
	"github.com/nelfander/Playingfield/internal/domain/tasks"
	"github.com/nelfander/Playingfield/internal/domain/user"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"github.com/nelfander/Playingfield/internal/infrastructure/postgres"
	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
	"github.com/nelfander/Playingfield/internal/infrastructure/storage"
	"github.com/nelfander/Playingfield/internal/infrastructure/ws"
	"github.com/nelfander/Playingfield/internal/interfaces/http"
	"github.com/nelfander/Playingfield/internal/interfaces/http/handlers"
	"github.com/nelfander/Playingfield/internal/interfaces/http/middleware"
	"github.com/nelfander/Playingfield/pkg/config"
)

func Run() {
	// --- Load config ---
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// --- Logger Setup (The slog way) ---
	slogHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // LevelDebug, Info, Warn, Error
	})
	logger := slog.New(slogHandler)
	slog.SetDefault(logger) // This makes slog.Info() work everywhere else(all packages)!

	// --- Postgres pool ---
	dbPool, err := postgres.NewPool(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// --- DB adapter for sqlc ---
	db := postgres.NewDBAdapter(dbPool)

	// --- SQLC wrapper ---
	queries := sqlc.New(db)

	// --- S3 / MinIO Storage Setup ---

	//  Create the AWS Credentials Provider
	creds := credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, "")

	//  Load the SDK configuration
	awscfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(cfg.S3Region),
		awsConfig.WithCredentialsProvider(creds),
	)
	if err != nil {
		logger.Error("unable to load AWS SDK config", "error", err)
		os.Exit(1)
	}

	// create S3 Client
	s3Client := s3.NewFromConfig(awscfg, func(o *s3.Options) {
		if cfg.S3Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.S3Endpoint)
		}
		o.UsePathStyle = cfg.S3UsePathStyle
	})

	// Initialize  StorageProvider
	storageProvider := storage.NewS3Storage(
		s3Client,
		cfg.S3BucketName,
		cfg.S3PublicURL,
		logger,
	)

	//  Quick check to see if storage is alive
	_, err = s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		logger.Warn("Storage (MinIO) is not reachable. Is Docker running?", "error", err)
	} else {
		logger.Info("Successfully connected to S3/MinIO storage", "bucket", cfg.S3BucketName)
	}

	// --- JWT Manager ---
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry)

	//  Initialize the Hub
	hub := ws.NewHub()

	// --- user repo + service + handler ---
	userRepo := postgres.NewUserRepository(db, queries)
	userService := user.NewService(userRepo)
	userHandler := handlers.NewUserHandler(userService, jwtManager)

	// Projects repo + service + handler
	projectsRepo := postgres.NewProjectRepository(db)
	projectsService := projects.NewService(projectsRepo, hub)
	projectHandler := handlers.NewProjectHandler(projectsService)

	// --- Task repo + service + handler ---
	taskRepo := postgres.NewTaskRepository(db)
	taskService := tasks.NewService(taskRepo, projectsRepo, storageProvider, hub)
	taskHandler := handlers.NewTaskHandler(taskService)

	// --- Chat/Messages repo + service + handler ---
	messageRepo := postgres.NewMessageRepository(db)
	chatService := messages.NewService(messageRepo, projectsRepo, hub)
	chatHandler := handlers.NewChatHandler(chatService)

	//  Start the Hub in a background goroutine
	go hub.Run()

	// --- Performance Profiling (pprof) ---
	// This starts a private server on port 6060 just for internal checks.
	// In production, I would block this port via firewall.
	go func() {
		slog.Info("Starting pprof sidecar", "port", "6060")
		if err := stdhttp.ListenAndServe("localhost:6060", nil); err != nil {
			slog.Error("pprof server failed", "error", err)
		}
	}()

	// WebSocket handler creation
	wsHandler := handlers.NewWSHandler(jwtManager, hub, chatService)

	// --- Echo server ---
	e := echo.New()

	// Override default error handler to centralize JSON formatting and slog logging.
	// This prevents sensitive internal error leakage to the client.
	e.HTTPErrorHandler = http.CustomHTTPErrorHandler

	// Apply globally to all routes
	e.Use(middleware.RateLimitMiddleware(jwtManager))

	// --- CORS Middleware ---
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			stdhttp.MethodGet,
			stdhttp.MethodPost,
			stdhttp.MethodPut,
			stdhttp.MethodDelete,
		},
		AllowHeaders: []string{
			echo.HeaderAuthorization,
			echo.HeaderContentType,
		},
	}))

	// --- Admin Group ---
	// Only accessible by users with the 'admin' role
	adminGroup := e.Group("/admin")
	adminGroup.Use(middleware.JWTMiddleware(jwtManager))
	adminGroup.Use(middleware.RequireRole(jwtManager, "admin"))

	// --- Auth Group ---
	authGroup := e.Group("")
	authGroup.Use(middleware.JWTMiddleware(jwtManager))

	http.RegisterRoutes(e, userHandler)

	// a group for all project-related routes w
	projectGroup := e.Group("/projects")
	projectGroup.Use(middleware.JWTMiddleware(jwtManager))

	// a group for specific task actions
	taskGroup := e.Group("/tasks")
	taskGroup.Use(middleware.JWTMiddleware(jwtManager))

	// --- Routes ---
	e.POST("/register", userHandler.Register)
	e.GET("/admin", userHandler.Admin, middleware.RequireRole(jwtManager, "admin"))
	e.POST("/users", userHandler.Register) // for now i leave it public to allow user creation
	e.POST("/login", userHandler.Login)
	e.GET("/health", func(c echo.Context) error { // so AWS can verify the container is running.
		return c.JSON(stdhttp.StatusOK, map[string]string{"status": "ok"})
	})
	authGroup.GET("/me", userHandler.Me)
	authGroup.GET("/users", userHandler.List)
	// DM Chat History: /messages/direct/:other_id
	authGroup.GET("/messages/direct/:other_id", chatHandler.GetDMHistory)

	//  Admin Routes
	adminGroup.GET("", userHandler.Admin)
	adminGroup.GET("/users", userHandler.AdminListAllUsers)
	adminGroup.DELETE("/users/:id", userHandler.ScrubUser)
	adminGroup.POST("/users/:id/toggle", userHandler.ToggleStatus)

	// project routes
	projectGroup.POST("", projectHandler.Create)
	projectGroup.PUT("/:id", projectHandler.Update)
	projectGroup.GET("", projectHandler.List)
	projectGroup.GET("/:id", projectHandler.GetByID)
	projectGroup.DELETE("/:id", projectHandler.DeleteProject)
	projectGroup.POST("/users", projectHandler.AddUserToProject)
	projectGroup.GET("/users", projectHandler.ListUsersInProject)
	projectGroup.DELETE("/users", projectHandler.RemoveUserFromProject)
	// project task list: /projects/:id/tasks
	projectGroup.GET("/:id/tasks", taskHandler.ListTaskByProject)
	// project chat history: /projects/:id/messages
	projectGroup.GET("/:id/messages", chatHandler.GetProjectHistory)

	// task routes
	taskGroup.POST("", taskHandler.CreateTask)
	taskGroup.PUT("/:id", taskHandler.UpdateTask)
	taskGroup.DELETE("/:id", taskHandler.DeleteTask)
	taskGroup.GET("/:id/history", taskHandler.GetTaskHistory)
	taskGroup.POST("/:id/attachments", taskHandler.UploadAttachment)
	taskGroup.GET("/:id/attachments", taskHandler.GetAttachments)
	taskGroup.DELETE("/attachments/:attachment_id", taskHandler.DeleteAttachment)
	taskGroup.GET("/attachments/:attachment_id/file", taskHandler.GetAttachmentFile)

	// websocket route
	e.GET("/ws", wsHandler.HandleConnection)

	// --- Graceful shutdown prep---
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Interrupt, syscall.SIGTERM)

	// --- Start server in a goroutine---
	go func() {
		logger.Info("Starting HTTP server", "port", cfg.Port)
		if err := e.Start(":" + cfg.Port); err != nil && err != stdhttp.ErrServerClosed {
			logger.Error("error starting server", "error", err)
		}
	}()

	// stop everything and wait for the signal ( ctrl + c)
	<-quitCh
	logger.Info("🚀 Shutdown signal received")

	// stop broadcasting and cleanup clients
	hub.Stop()

	// "Deadline" 10 secs
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() //  prevent memory leaks

	// stop accepting new requests and finish
	// current requests(up until the 10s deadline).
	if err := e.Shutdown(ctx); err != nil {
		logger.Error("❌ Server Shutdown Failed", "error", err)
		os.Exit(1)
	}

	logger.Info("👋 Server exited gracefully")
}
