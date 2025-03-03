package dependencies

import (
	"log/slog"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/dev3mike/go-api-swagger-boilerplate/internal/environments"
	clerkhandler "github.com/dev3mike/go-api-swagger-boilerplate/internal/handlers/clerk_handler"
	indexhandler "github.com/dev3mike/go-api-swagger-boilerplate/internal/handlers/index_handler"
	"github.com/dev3mike/go-api-swagger-boilerplate/internal/interfaces"
	clerkservice "github.com/dev3mike/go-api-swagger-boilerplate/internal/services/clerk"
)

var ClerkHandler interfaces.ClerkHandler
var IndexHandler interfaces.IndexHandler

func Initialize(logger *slog.Logger) {
	logger.Info("🔄 Initializing Dependencies")

	clerkClient, err := clerk.NewClient(environments.Env.ClerkSecretKey)
	if err != nil {
		panic(err)
	}

	ClerkHandler = clerkhandler.NewClerkHandler(clerkservice.NewClerkService(&clerkClient))
	IndexHandler = indexhandler.NewIndexHandler()

	logger.Info("✅ Dependencies Initialized Successfully")
}
