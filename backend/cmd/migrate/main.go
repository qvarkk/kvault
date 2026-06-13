package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"qvarkk/kvault/config"
	"qvarkk/kvault/internal/postgres"
	"qvarkk/kvault/logger"
	"qvarkk/kvault/migrations"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	err = logger.Init("migrate", cfg.Debug)
	if err != nil {
		log.Fatalf("Failed to initialize zap logger: %v", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Database)
	pgConfig := postgres.Config{
		DSN:             dsn,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Minute * 5,
	}

	pg, err := postgres.NewPostgres(pgConfig)
	if err != nil {
		zap.L().Fatal("Connection to database failed", zap.Error(err))
	}
	defer func() {
		if err := pg.Close(); err != nil {
			zap.L().Error("failed to close database connection", zap.Error(err))
		}
	}()

	migrator, err := migrations.NewMigrator(pg.DB.DB, cfg.DB.Database)
	if err != nil {
		zap.L().Fatal("Failed to initialize migration module", zap.Error(err))
	}

	if len(os.Args) < 2 {
		zap.L().Fatal("usage: migrate [up|down|steps N|force V|version]")
	}

	switch os.Args[1] {
	case "up":
		err = migrator.Up()
	case "down":
		err = migrator.Down()
	case "steps":
		n, _ := strconv.Atoi(os.Args[2])
		err = migrator.Steps(n)
	case "force":
		v, _ := strconv.Atoi(os.Args[2])
		err = migrator.Force(v)
	case "version":
		v, dirty, verr := migrator.Version()
		zap.L().Info("version status", zap.Uint("version", v), zap.Bool("dirty", dirty))
		err = verr
	default:
		zap.L().Fatal("unknown command", zap.String("command", os.Args[1]))
	}

	if err != nil {
		zap.L().Fatal("unknown error", zap.Error(err))
	}

	zap.L().Info("command run successfully")
}
