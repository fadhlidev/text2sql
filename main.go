package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/jackc/pgx/v5/stdlib" // Postgres driver
	"github.com/joho/godotenv"
	"github.com/liushuangls/go-anthropic/v2"
	_ "modernc.org/sqlite" // SQLite driver
	openai "github.com/sashabaranov/go-openai"
	"google.golang.org/genai"

	"github.com/fadhlidev/text2sql/handler"
	"github.com/fadhlidev/text2sql/schema"
	"github.com/fadhlidev/text2sql/text2sql"
)

func main() {
	// Load .env file (ignored in production if file doesn't exist)
	_ = godotenv.Load()

	dbURI := mustEnv("DB_URI")

	// Infer dialect and select driver
	dialect := dialectFromURI(dbURI)
	driver := driverForDialect(dialect)

	// Connect to database
	db, err := sql.Open(driver, dbURI)
	if err != nil {
		log.Fatalf("failed to open database (%s): %v", dialect, err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify the connection is alive
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}
	log.Println("database connected")

	// Load schema ONCE at startup
	s, err := schema.Introspect(ctx, db, dialect)
	if err != nil {
		log.Fatalf("schema introspection failed: %v", err)
	}
	log.Printf("schema loaded: %d tables", len(s.Tables))

	// Wire up LLM provider
	llmClient := initLLM()
	conv := text2sql.New(llmClient, s)
	exec := text2sql.NewExecutor(db)
	qh := handler.NewQueryHandler(conv, exec)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "github.com/fadhlidev/text2sql v1.0",
	})

	// Middleware
	app.Use(recover.New()) // catches panics and returns 500 instead of crashing
	app.Use(logger.New())  // logs every request to stdout

	// Routes
	app.Post("/query", qh.Query)
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Start server
	log.Println("server starting on :3000")
	log.Fatal(app.Listen(":3000"))
}

// mustEnv reads an environment variable and exits if it is not set
func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable not set: %s", key)
	}
	return v
}

// dialectFromURI infers the database dialect from the connection string prefix
func dialectFromURI(uri string) string {
	switch {
	case strings.HasPrefix(uri, "postgres"):
		return "postgres"
	case strings.HasPrefix(uri, "mysql"):
		return "mysql"
	default:
		return "sqlite"
	}
}

// driverForDialect returns the registered database/sql driver name for the given dialect
func driverForDialect(dialect string) string {
	switch dialect {
	case "postgres":
		return "pgx"
	case "mysql":
		return "mysql"
	case "sqlite":
		return "sqlite"
	default:
		return "pgx"
	}
}

// initLLM initializes the LLM client based on environment variables
func initLLM() text2sql.LLMClient {
	provider := strings.ToLower(os.Getenv("LLM_PROVIDER"))
	if provider == "" {
		provider = "openai" // default
	}

	model := os.Getenv("LLM_MODEL")

	switch provider {
	case "anthropic":
		apiKey := mustEnv("ANTHROPIC_API_KEY")
		if model == "" {
			model = "claude-3-5-sonnet-20240620"
		}
		return text2sql.NewAnthropicClient(anthropic.NewClient(apiKey), model)
	case "gemini":
		apiKey := mustEnv("GEMINI_API_KEY")
		if model == "" {
			model = "gemini-1.5-flash"
		}
		client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
			APIKey: apiKey,
		})
		if err != nil {
			log.Fatalf("failed to create gemini client: %v", err)
		}
		return text2sql.NewGeminiClient(client, model)
	case "openai":
		apiKey := mustEnv("OPENAI_API_KEY")
		if model == "" {
			model = "gpt-4o"
		}
		return text2sql.NewOpenAIClient(openai.NewClient(apiKey), model)
	default:
		log.Fatalf("unsupported LLM provider: %s", provider)
		return nil
	}
}
