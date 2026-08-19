package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"wongnok/internal/config"
	"wongnok/internal/httputil"
	"wongnok/internal/middleware"
	"wongnok/internal/platform/database"
	"wongnok/internal/user"

	_ "wongnok/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title			Wongnok API
//	@version		1.0
//	@description	API สำหรับจัดการกับระบบสูตรอาหาร
//	@host			localhost:8080
//	@BasePath		/api/v1
//	@schemas		http https

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				พิมพ์ "Bearer" ตามด้วย space แล้วตามด้วย JWT token เช่น "Bearer eyJhbGci..."
func main() {
	if err := run(); err != nil {
		slog.Error("service stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Default logger
	slog.SetDefault(newLogger(os.Stdout, "wongnok", config.Logging{Level: "DEBUG", Format: "text"}))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration:\n%s", config.Humanize(err))
	}

	// Setup logger
	slog.SetDefault(newLogger(os.Stdout, cfg.App.Name, cfg.Logging))

	// Signal context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, sqldb, err := database.Open(ctx, cfg.Database.PostgresDSN)
	if err != nil {
		log.Fatal("database connection:", err)
	}
	defer sqldb.Close()

	// Dependency injection
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	// Register path
	router := gin.Default()
	if cfg.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Register cors
	router.Use(cors.Default())

	// Group version
	v1 := router.Group("/api/v1")

	// Auth resource
	authRoute := v1.Group("/auth")

	// Inline code เพื่อ demo
	authRoute.GET("/login", func(ctx *gin.Context) {
		token, err := middleware.GenerateToken("user123")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, httputil.ErrorResponse{Message: "cannot generate token"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"token": token})
	})

	// User resource
	userRoute := v1.Group("/users")

	// JWT Verify middleware
	userRoute.Use(middleware.JWT())

	// curl -X GET \
	// -H "Authorization: Bearer {token}" \n
	// http://localhost:8080/api/v1/users/:id
	userRoute.GET("/:id", userHandler.GetUser)

	// curl -X POST http://localhost:8080/api/v1/users -H "Content-Type: application/json" -d '{"email":"taro@devpool.pea"}'
	userRoute.POST("", userHandler.CreateUser)

	// Register swagger
	router.GET("swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Server
	serv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := serv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "server error", "error", serv.Addr)
		}
	}()
	slog.InfoContext(ctx, "server started", "addr", serv.Addr)

	// Graceful shutdown
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	slog.InfoContext(shutdownCtx, "shutting down server")

	return serv.Shutdown(shutdownCtx)
}

func newLogger(writer io.Writer, name string, log config.Logging) *slog.Logger {
	opts := &slog.HandlerOptions{Level: log.SlogLevel()}

	var handler slog.Handler = slog.NewJSONHandler(writer, opts)
	if log.Format == "text" {
		handler = slog.NewTextHandler(writer, opts)
	}

	return slog.New(handler).With("service", name)
}
