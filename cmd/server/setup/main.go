package setup

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dev3mike/go-api-swagger-boilerplate/cmd/server/dependencies"
	"github.com/dev3mike/go-api-swagger-boilerplate/cmd/server/logger"
	"github.com/dev3mike/go-api-swagger-boilerplate/cmd/server/router"
	"github.com/dev3mike/go-api-swagger-boilerplate/internal/database"
	"github.com/dev3mike/go-api-swagger-boilerplate/internal/environments"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

var r http.Handler

func SetupServerPrerequisites() {
	logger.Logger = slog.Default()

	// Initialize environment settings
	environments.InitializeEnvironments(logger.Logger)

	logger.Logger.Info("🔄 Initializing server on environment: " + environments.Env.Environment)

	// Connect to the database
	database.Connect(logger.Logger)

	// Initialize the router
	r = initializeRouter(logger.Logger)
}

func StartServer() {
	// Create the HTTP server
	serverPort := environments.Env.Port
	server := &http.Server{
		Addr:    ":" + serverPort,
		Handler: r,
	}

	// Start the server
	startServer(server, logger.Logger)
}

func DisconnectDatabase() {
	// Close the database connection
	if err := database.DB.Close(); err != nil {
		logger.Logger.Error("❌ Failed to disconnect from the database: %v", err)
		return
	}

	logger.Logger.Info("✅ Disconnected from the database successfully!")
}

// initializeRouter sets up the router with all the necessary middlewares and routes
func initializeRouter(logger *slog.Logger) http.Handler {

	// Initialize dependencies
	dependencies.Initialize(logger)

	r := router.InitializeRouter()

	// Set middlewares
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.StripSlashes)

	// Setup routes
	r = router.SetupRoutes(r, logger)

	return r
}

func startServer(server *http.Server, logger *slog.Logger) {
	// Channel to listen for interrupt or terminate signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("✅ Starting server on port: " + server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("ListenAndServe(): %v", err)
		}
	}()

	// Wait for a signal
	<-stop

	// Initiate graceful shutdown
	logger.Info("❌ Shutting down server...")
	shutdownServer(server, logger)
}

func shutdownServer(server *http.Server, logger *slog.Logger) {
	// Create a deadline for the shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("❌ Server forced to shutdown: %v", err)
	}

	logger.Info("❌ Server exiting")
}
