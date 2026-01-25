package app

import (
	"context"
	"log"
	stdhttp "net/http"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/nelfander/Playingfield/internal/domain/messages"
	"github.com/nelfander/Playingfield/internal/domain/projects"
	"github.com/nelfander/Playingfield/internal/domain/tasks"
	"github.com/nelfander/Playingfield/internal/domain/user"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"github.com/nelfander/Playingfield/internal/infrastructure/postgres"
	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
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
	taskService := tasks.NewService(taskRepo, projectsRepo, hub)
	taskHandler := handlers.NewTaskHandler(taskService)

	// --- Chat/Messages repo + service + handler ---
	messageRepo := postgres.NewMessageRepository(db)
	chatService := messages.NewService(messageRepo, projectsRepo, hub)
	chatHandler := handlers.NewChatHandler(chatService)

	//  Start the Hub in a background goroutine
	go hub.Run()

	// --- Seed default admin ---
	if err := postgres.SeedAdminUser(context.Background(), userRepo); err != nil {
		log.Fatal("failed to seed admin user:", err)
	}

	// WebSocket handler creation
	wsHandler := handlers.NewWSHandler(jwtManager, hub, chatService)
	// --- Handler ---

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
	authGroup.Use(middleware.JWTMiddleware(jwtManager))
	authGroup.GET("/me", userHandler.Me)
	authGroup.GET("/users", userHandler.List)
	// DM Chat History: /messages/direct/:other_id
	authGroup.GET("/messages/direct/:other_id", chatHandler.GetDMHistory)

	http.RegisterRoutes(e, userHandler)

	// a group for all project-related routes w
	r := e.Group("/projects")
	r.Use(middleware.JWTMiddleware(jwtManager))

	// a group for specific task actions
	t := e.Group("/tasks")
	t.Use(middleware.JWTMiddleware(jwtManager))

	// --- Routes ---
	e.POST("/register", userHandler.Register)
	e.GET("/admin", userHandler.Admin, middleware.RequireRole(jwtManager, "admin"))
	e.POST("/users", userHandler.Register) // for now i leave it public to allow user creation
	e.POST("/login", userHandler.Login)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(stdhttp.StatusOK, map[string]string{"status": "ok"})
	})

	// project routes
	r.POST("", projectHandler.Create)
	r.PUT("/:id", projectHandler.Update)
	r.GET("", projectHandler.List)
	r.GET("/:id", projectHandler.GetByID)
	r.DELETE("/:id", projectHandler.DeleteProject)
	r.POST("/users", projectHandler.AddUserToProject)
	r.GET("/users", projectHandler.ListUsersInProject)
	r.DELETE("/users", projectHandler.RemoveUserFromProject)

	// task routes
	t.POST("", taskHandler.CreateTask)
	t.PUT("/:id", taskHandler.UpdateTask)
	t.DELETE("/:id", taskHandler.DeleteTask)
	t.GET("/:id/history", taskHandler.GetTaskHistory)

	// project task list: /projects/:id/tasks
	r.GET("/:id/tasks", taskHandler.ListTaskByProject)
	// project chat history: /projects/:id/messages
	r.GET("/:id/messages", chatHandler.GetProjectHistory)

	// websocket route
	e.GET("/ws", wsHandler.HandleConnection)

	// --- Start server ---
	logger.Println("starting HTTP server on :" + cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		logger.Println("server stopped:", err)
	}
}
