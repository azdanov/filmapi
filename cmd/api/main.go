package main

import (
	"context"
	"expvar"
	"flag"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"filmapi.azdanov.dev/internal/jsonlog"
	"filmapi.azdanov.dev/internal/mailer"

	"filmapi.azdanov.dev/internal/data"

	"github.com/jackc/pgx/v4/pgxpool"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://filmapi:secret@localhost/filmapi", "PostgreSQL DSN")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "32cb8a7e024879", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "260798f911e5b7", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "FilmApi <no-reply@filmapi.io>", "SMTP sender")
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()

	logger.PrintInfo("database connection pool established", nil)

	expvar.NewString("version").Set(version)

	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("database", expvar.Func(func() interface{} {
		return struct {
			AcquireCount         int64
			AcquireDuration      time.Duration
			AcquiredConns        int32
			CanceledAcquireCount int64
			ConstructingConns    int32
			EmptyAcquireCount    int64
			IdleConns            int32
			MaxConns             int32
			TotalConns           int32
		}{
			db.Stat().AcquireCount(),
			db.Stat().AcquireDuration(),
			db.Stat().AcquiredConns(),
			db.Stat().CanceledAcquireCount(),
			db.Stat().ConstructingConns(),
			db.Stat().EmptyAcquireCount(),
			db.Stat().IdleConns(),
			db.Stat().MaxConns(),
			db.Stat().TotalConns(),
		}
	}))

	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	err = app.serve()
	logger.PrintFatal(err, nil)
}

func openDB(cfg config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	poolDB, err := pgxpool.ConnectConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = poolDB.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return poolDB, nil
}
