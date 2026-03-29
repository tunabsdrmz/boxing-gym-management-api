package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/tunabsdrmz/boxing-gym-management/internal/config"
	"github.com/tunabsdrmz/boxing-gym-management/internal/db"
	"github.com/tunabsdrmz/boxing-gym-management/internal/env"
	"github.com/tunabsdrmz/boxing-gym-management/internal/handler"
	"github.com/tunabsdrmz/boxing-gym-management/internal/middleware"
	"github.com/tunabsdrmz/boxing-gym-management/internal/repository"
	"github.com/tunabsdrmz/boxing-gym-management/internal/routes"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	cfg := config.Config{
		Port: env.GetStr("PORT", ":3000"),
		DB: config.DbConfig{
			// lib/pq expects a URL or key=value DSN, not a bare host:port.
			Addr: env.GetStr("DB_ADDR", "postgres://postgres:postgres@localhost:5433/boxing-gym-management?sslmode=disable"),
		},
		JWT: config.JWTConfig{
			Secret: env.GetStr("JWT_SECRET", "secret"),
		},
	}
	config.App = cfg
	database, err := db.NewDB(cfg.DB)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer database.Close()
	repository := repository.NewRepository(database)
	handler := handler.NewHandler(repository)
	app := &application{
		config:     config.App,
		repository: repository,
		handler:    handler,
	}

	apiRouter := routes.NewRouter(handler, middleware.Authenticate)
	mux := app.mount(apiRouter)
	log.Printf("Server is running on port %s", cfg.Port)
	log.Fatal(app.run(mux))

}
