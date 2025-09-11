package main

import (
	"expvar"
	"log"
	"runtime"
	"time"

	"github.com/joho/godotenv"
	"github.com/n-korel/social-api/internal/auth"
	"github.com/n-korel/social-api/internal/db"
	"github.com/n-korel/social-api/internal/env"
	"github.com/n-korel/social-api/internal/mailer"
	"github.com/n-korel/social-api/internal/ratelimiter"
	"github.com/n-korel/social-api/internal/service"
	"github.com/n-korel/social-api/internal/store"
	"github.com/n-korel/social-api/internal/store/cache"
	"github.com/redis/go-redis/v9"
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
		redisCfg: redisConfig{
			addr:     env.GetString("REDIS_ADDR", "localhost:6379"),
			password: env.GetString("REDIS_PW", ""),
			db:       env.Getint("REDIS_DB", 0),
			enabled:  env.GetBool("REDIS_ENABLED", false),
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
		auth: authConfig{
			basic: basicConfig{
				user: env.GetString("AUTH_BASIC_USER", ""),
				pass: env.GetString("AUTH_BASIC_PASS", ""),
			},
			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", "example"),
				exp:    time.Hour * 24 * 2, // 2 Days
				host:   env.GetString("AUTH_TOKEN_HOST", "example"),
			},
		},
		rateLimiter: ratelimiter.Config{
			RequestsPerTimeFrame: env.Getint("RATELIMITER_REQUESTS_COUNT", 20),
			TimeFrame:            time.Second * 5,
			Enabled:              env.GetBool("RATE_LIMITER_ENABLED", true),
		},
	}

	// Initialize Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Initialize Database
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

	// Initialize cache
	var rdb *redis.Client
	var cacheStorage service.CacheStorage

	if cfg.redisCfg.enabled {
		rdb = cache.NewRedisClient(
			cfg.redisCfg.addr,
			cfg.redisCfg.password,
			cfg.redisCfg.db,
		)
		logger.Info("Redis cache connected!")
		defer rdb.Close()

		cacheStorage = cache.NewRedisStorage(rdb)
	}

	// Initialize Rate limiter
	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.rateLimiter.RequestsPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	// Initialize repository layer
	store := store.NewStorage(db)

	// Initialize Mailer
	mailtrap, err := mailer.NewMailTrapClient(cfg.mail.mailTrap.username, cfg.mail.mailTrap.password, cfg.mail.fromEmail)
	if err != nil {
		logger.Fatal(err)
	}

	// Initialize Authenticator
	JWTAuthenticator := auth.NewJWTAuthenticator(
		cfg.auth.token.secret,
		cfg.auth.token.host,
		cfg.auth.token.host,
	)

	// Initialize Service layer
	userServiceConfig := service.UserServiceConfig{
		FrontendURL:     cfg.frontendURL,
		MailExpiration:  cfg.mail.exp,
		IsProductionEnv: cfg.env == "production",
	}

	authServiceConfig := service.AuthServiceConfig{
		TokenExpiration: cfg.auth.token.exp,
		TokenHost:       cfg.auth.token.host,
	}

	services := service.NewServices(
		store,
		cacheStorage,
		mailtrap,
		JWTAuthenticator,
		userServiceConfig,
		authServiceConfig,
	)

	app := &application{
		config:        cfg,
		store:         store,
		cacheStorage:  cacheStorage,
		services:      services,
		logger:        logger,
		mailer:        mailtrap,
		authenticator: JWTAuthenticator,
		rateLimiter:   rateLimiter,
	}

	// Metrics collected
	expvar.NewString("version").Set(version)
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	mux := app.mount()

	logger.Fatal((app.run(mux)))
}
