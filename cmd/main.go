package main

import (
	"log"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/tunabsdrmz/boxing-gym-management/internal/config"
	"github.com/tunabsdrmz/boxing-gym-management/internal/db"
	"github.com/tunabsdrmz/boxing-gym-management/internal/env"
	"github.com/tunabsdrmz/boxing-gym-management/internal/handler"
	"github.com/tunabsdrmz/boxing-gym-management/internal/middleware"
	"github.com/tunabsdrmz/boxing-gym-management/internal/repository"
	"github.com/tunabsdrmz/boxing-gym-management/internal/routes"
)

func parseOriginsCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	accessMin := env.GetInt("JWT_ACCESS_TTL_MINUTES", 15)
	refreshDays := env.GetInt("JWT_REFRESH_TTL_DAYS", 7)
	cfg := config.Config{
		Port: env.GetStr("PORT", ":3000"),
		DB: config.DbConfig{
			Addr: env.GetStr("DB_ADDR", "postgres://postgres:postgres@localhost:5433/boxing-gym-management?sslmode=disable"),
		},
		JWT: config.JWTConfig{
			Secret:     env.GetStr("JWT_SECRET", "secret"),
			AccessTTL:  time.Duration(accessMin) * time.Minute,
			RefreshTTL: time.Duration(refreshDays) * 24 * time.Hour,
		},
		CORS: config.CORSConfig{
			AllowedOrigins: parseOriginsCSV(env.GetStr("CORS_ALLOWED_ORIGINS", "")),
		},
		Security: config.SecurityConfig{
			RateLimitRPM: env.GetInt("RATE_LIMIT_RPM", 0),
		},
		Dev: config.DevConfig{
			ReturnPasswordResetToken: env.GetBool("DEV_RETURN_PASSWORD_RESET_TOKEN", false),
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
