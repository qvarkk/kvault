package main

import (
	"fmt"
	"log"
	"os"
	"qvarkk/kvault/config"
	"qvarkk/kvault/internal/postgres"
	"qvarkk/kvault/logger"
	"qvarkk/kvault/migrations"
	"strconv"
	"time"

	"go.uber.org/zap"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	err = logger.Init("migrate", config.Debug)
	if err != nil {
		log.Fatalf("Failed to initialize zap logger: %v", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", config.DB.Username, config.DB.Password, config.DB.Host, config.DB.Port, config.DB.Database)
	pgConfig := postgres.Config{
		DSN:             dsn,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Minute * 5,
	}

	pg, err := postgres.NewPostgres(pgConfig)
	if err != nil {
		logger.Logger.Fatal("Connection to database failed", zap.Error(err))
	}
	defer pg.Close()

	migrator, err := migrations.NewMigrator(pg.DB.DB, config.DB.Database)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize migration module", zap.Error(err))
	}

	if len(os.Args) < 2 {
		logger.Logger.Fatal("usage: migrate [up|down|steps N|force V|version]")
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
		logger.Logger.Info("version status", zap.Uint("version", v), zap.Bool("dirty", dirty))
		err = verr
	default:
		logger.Logger.Fatal("unknown command", zap.String("command", os.Args[1]))
	}

	if err != nil {
		logger.Logger.Fatal("unknown error", zap.Error(err))
	}

	logger.Logger.Info("command run successfully")
}
