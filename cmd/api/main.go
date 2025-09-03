package main

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/n-korel/social-api/internal/db"
	"github.com/n-korel/social-api/internal/env"
	"github.com/n-korel/social-api/internal/mailer"
	"github.com/n-korel/social-api/internal/store"
	"go.uber.org/zap"
)

const version = "0.0.1"

//	@title			Social Forum Golang API
//	@description	API for Social Forum Golang
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	cfg := config{
		addr:        ":" + env.GetString("PORT", "8080"),
		apiURL:      env.GetString("EXTERNAL_URL", "localhost:8080"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:5173"),
		db: dbConfig{
			dsn:          env.GetString("DSN", "host=localhost user=postgres password=my_pass dbname=social-api port=5432 sslmode=disable"),
			maxOpenConns: env.Getint("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.Getint("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("ENV", "development"),
		mail: mailConfig{
			exp:       time.Hour * 24 * 2, // 2 Days
			fromEmail: env.GetString("FROM_EMAIL", ""),
			mailTrap: mailTrapConfig{
				username: env.GetString("MAILTRAP_USER", ""),
				password: env.GetString("MAILTRAP_PASS", ""),
			},
		},
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Database
	db, err := db.New(
		cfg.db.dsn,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("Database has connected!")

	store := store.NewStorage(db)

	// Mailer
	mailtrap, err := mailer.NewMailTrapClient(cfg.mail.mailTrap.username, cfg.mail.mailTrap.password, cfg.mail.fromEmail)
	if err != nil {
		logger.Fatal(err)
	}

	app := &application{
		config: cfg,
		store:  store,
		logger: logger,
		mailer: mailtrap,
	}

	mux := app.mount()

	logger.Fatal((app.run(mux)))
}
