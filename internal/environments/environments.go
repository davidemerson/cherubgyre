package environments

import (
	"log/slog"

	"github.com/dev3mike/go-xenv"
)

var Env struct {
	Port           string `json:"PORT" validators:"required,minLength:4,maxLength:5" transformers:"trim,lowercase"`
	DbConnection   string `json:"DB_CONNECTION" validators:"required,minLength:4" transformers:"trim"`
	MigrationPath  string `json:"MIGRATION_PATH" validators:"required,minLength:4" transformers:"trim,lowercase"`
	Environment    string `json:"ENV" validators:"enum:development-staging-production" transformers:"trim,lowercase"`
	ClerkSecretKey string `json:"CLERK_SECRECT_KEY" validators:"required" transformers:"trim"`
}

func InitializeEnvironments(logger *slog.Logger) {
	// Load environment variables from a .env file
	if err := xenv.LoadEnvFile(".env"); err != nil {
		logger.Info("Error loading .env file: ", err)
	}

	// Validate environment variables
	if err := xenv.ValidateEnv(&Env); err != nil {
		logger.Error("Failed to validate environment: ", err)
		panic("Failed to validate environment")
	}

	logger.Info("✅ Environment validated successfully!")
}
