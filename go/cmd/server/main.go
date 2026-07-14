package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	osClientLib "github.com/opensearch-project/opensearch-go/v2"
	todoHttp "github.com/rezafahlevi/gotodo/internal/delivery/http"
	opensearchRepo "github.com/rezafahlevi/gotodo/internal/repository/opensearch"
	"github.com/rezafahlevi/gotodo/internal/repository/postgres"
	"github.com/rezafahlevi/gotodo/internal/usecase"
)

func main() {
	// 1. Load configuration (Environment variables)
	// godotenv loads variables from .env file into the system environment
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// 2. Setup Database Connection (pgxpool)
	// pgxpool manages a connection pool automatically which is best practice for prod.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	// Check if connection is actually working
	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}
	log.Println("Connected to PostgreSQL database successfully.")

	// Setup OpenSearch Connection
	osURL := os.Getenv("OPENSEARCH_URL")
	if osURL == "" {
		osURL = "http://localhost:9200"
	}
	
	osClient, err := osClientLib.NewClient(osClientLib.Config{
		Addresses: []string{osURL},
	})
	if err != nil {
		log.Fatalf("Unable to create OpenSearch client: %v\n", err)
	}
	
	res, err := osClient.Info()
	if err != nil {
		log.Printf("OpenSearch is not reachable: %v\n", err)
	} else if res.IsError() {
		log.Printf("OpenSearch error response: %s\n", res.String())
	} else {
		log.Println("Connected to OpenSearch successfully.")
		res.Body.Close()
	}

	// 3. Dependency Injection
	// --- OLD STACK ---
	todoRepo := postgres.NewTodoRepository(dbPool)
	todoUsecase := usecase.NewTodoUseCase(todoRepo, osClient)

	// --- CQRS STACK ---
	todoQueryRepo := opensearchRepo.NewTodoQueryRepository(osClient)
	todoQueryUseCase := usecase.NewTodoQueryHandler(todoQueryRepo)

	// 4. Setup HTTP Router
	mux := http.NewServeMux()
	
	// Register old route (/todos)
	todoHttp.NewTodoHandler(mux, todoUsecase)
	
	// Register CQRS route (/cqrs/todos)
	todoHttp.NewTodoCQRSHandler(mux, todoQueryUseCase)

	// Wrap mux with CORS middleware
	handlerWithCORS := corsMiddleware(mux)

	// 5. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handlerWithCORS); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// corsMiddleware is a simple middleware to handle CORS for our React frontend.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
		if allowedOrigin == "" {
			allowedOrigin = "*" // Default to allow all for development
		}
		
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
