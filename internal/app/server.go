package app

import (
	"context"
	"log"
	stdhttp "net/http"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/nelfander/Playingfield/internal/domain/messages"
	"github.com/nelfander/Playingfield/internal/domain/projects"
	"github.com/nelfander/Playingfield/internal/domain/user"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"github.com/nelfander/Playingfield/internal/infrastructure/postgres"
	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
	"github.com/nelfander/Playingfield/internal/infrastructure/ws"
	"github.com/nelfander/Playingfield/internal/interfaces/http"
	"github.com/nelfander/Playingfield/internal/interfaces/http/handlers"
	"github.com/nelfander/Playingfield/internal/interfaces/http/middleware"
	httpMiddleware "github.com/nelfander/Playingfield/internal/interfaces/http/middleware"
	"github.com/nelfander/Playingfield/pkg/config"
)

func Run() {
	// --- Load config ---

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	// --- Logger ---
	logger := log.Default()

	// --- Postgres pool ---
	dbPool, err := postgres.NewPool(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("failed to connect to database:", err)
	}
	defer dbPool.Close()

	// --- DB adapter for sqlc ---
	db := postgres.NewDBAdapter(dbPool)

	// --- SQLC wrapper ---
	queries := sqlc.New(db)

	// --- Repository ---
	userRepo := postgres.NewUserRepository(db, queries)

	//  Initialize the Hub
	hub := ws.NewHub()

	//  Start the Hub in a background Goroutine
	go hub.Run()

	// Projects repo + service
	projectsRepo := postgres.NewProjectRepository(db)
	projectsService := projects.NewService(projectsRepo, queries, hub)
	projectHandler := handlers.NewProjectHandler(projectsService)

	// --- Chat/Messages logic ---
	messageRepo := postgres.NewMessageRepository(db)
	chatService := messages.NewService(messageRepo, queries, hub)
	chatHandler := handlers.NewChatHandler(chatService)

	// --- Seed default admin ---
	if err := postgres.SeedAdminUser(context.Background(), userRepo); err != nil {
		log.Fatal("failed to seed admin user:", err)
	}

	// --- Service ---
	userService := user.NewService(userRepo)

	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry)

	// WebSocket handler creation
	wsHandler := handlers.NewWSHandler(jwtManager, hub, chatService)
	// --- Handler ---
	userHandler := handlers.NewUserHandler(userService, jwtManager)

	// --- Echo server ---
	e := echo.New()

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

	authGroup := e.Group("")
	authGroup.Use(httpMiddleware.JWTMiddleware(jwtManager))
	authGroup.GET("/me", userHandler.Me)
	authGroup.GET("/users", userHandler.List)
	// DM Chat History: /messages/direct/:other_id
	authGroup.GET("/messages/direct/:other_id", chatHandler.GetDMHistory)

	http.RegisterRoutes(e, userHandler)

	// a group for all project-related routes
	r := e.Group("/projects")
	r.Use(httpMiddleware.JWTMiddleware(jwtManager))

	// --- Routes ---
	e.POST("/register", userHandler.Register)
	e.GET("/me", userHandler.Me, middleware.JWTMiddleware(jwtManager)) //	For account panel
	e.GET("/admin", userHandler.Admin, middleware.RequireRole(jwtManager, "admin"))
	e.POST("/users", userHandler.Register) // for now i leave it public to allow user creation
	e.POST("/login", userHandler.Login)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(stdhttp.StatusOK, map[string]string{"status": "ok"})
	})
	// project routes
	r.POST("", projectHandler.Create)     // /projects
	r.GET("", projectHandler.List)        // /projects
	r.GET("/:id", projectHandler.GetByID) // This handles /projects/1, /projects/2, etc.
	r.DELETE("/:id", projectHandler.DeleteProject)
	r.POST("/users", projectHandler.AddUserToProject)        // /projects/users
	r.GET("/users", projectHandler.ListUsersInProject)       // /projects/users
	r.DELETE("/users", projectHandler.RemoveUserFromProject) // /projects/users
	// Project Chat History: /projects/:id/messages
	r.GET("/:id/messages", chatHandler.GetProjectHistory)

	// websocket route
	e.GET("/ws", wsHandler.HandleConnection)

	// --- Start server ---
	logger.Println("starting HTTP server on :" + cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		logger.Println("server stopped:", err)
	}
}
