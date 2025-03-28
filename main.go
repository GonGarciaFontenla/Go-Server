package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"database/sql"

	"github.com/GonGarciaFontenla/rssagg/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// apiConfig holds the database instance for API handlers
type apiConfig struct {
	DB *database.Queries
}

func main() {
	// Load environment variables from .env file (useful for local dev)
	godotenv.Load(".env")

	// Get the server port from the environment
	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT is not found in the environment")
	}

	// Get the database URL from the environment
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB_URL is not found in the environment")
	}

	// Open a connection to the PostgreSQL database
	conn, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("Error opening connection to the database: ", err)
	}

	// Create a new instance of the database queries struct
	db := database.New(conn)
	apiCfg := apiConfig{
		DB: db,
	}

	// Start the background RSS scraping process
	go startScraping(db, 10, time.Minute)

	// Set up the HTTP router
	router := chi.NewRouter()

	// Configure CORS to allow requests from any origin (can be restricted if needed)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Create a subrouter for version 1 of the API
	v1Router := chi.NewRouter()
	v1Router.Get("/healthz", handlerReadiness) // Health check endpoint
	v1Router.Get("/err", handlerErr)           // Test error response

	// User-related endpoints
	v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUser))
	v1Router.Post("/users", apiCfg.handlerCreateUser)

	// Feed-related endpoints
	v1Router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))
	v1Router.Get("/feeds", apiCfg.handlerGetFeeds)

	// Feed following endpoints
	v1Router.Post("/feedsFollows", apiCfg.middlewareAuth(apiCfg.handlerCreateFeedFollow))
	v1Router.Get("/feedsFollows", apiCfg.middlewareAuth(apiCfg.handlerGetFeedFollows))
	v1Router.Delete("/feedsFollows/{feedFollowID}", apiCfg.middlewareAuth(apiCfg.handlerDeleteFeedFollows))

	// Posts endpoint
	v1Router.Get("/posts", apiCfg.middlewareAuth(apiCfg.handlerGetPostsForUser))

	// Mount the API routes under /v1
	router.Mount("/v1", v1Router)

	// Configure and start the HTTP server
	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}

	log.Printf("Server starting on port %v", portString)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
