package app

import (
	"TransactiStream/internal/config"
	httphandler "TransactiStream/internal/delivery/http"
	"TransactiStream/internal/logger"
	"TransactiStream/internal/repository/postgres"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"net/http"
	"os"
)

func Run(confDir string) {
	logger.InitLogger()

	cfg, err := config.MustLoad(confDir)
	if err != nil {
		logger.Errorf("failed to load config: %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.DBName,
	)
	logger.Infof("Connect string: %s", connString)

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		logger.Errorf("Unable to establish connection: %v", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())
	logger.Info("Connection established")

	err = postgres.CreateTables(ctx, *conn)
	if err != nil {
		logger.Errorf("Unable to create tables: %v", err)
		os.Exit(1)
	}
	logger.Info("Tables created")

	repo := postgres.NewPostgres(conn)
	handler := httphandler.NewHandler(repo)

	http.HandleFunc("/transaction", handler.CreateTransaction)

	srv := &http.Server{
		Addr: cfg.HTTP.Host + ":" + cfg.HTTP.Port,
	}

	go func() {
		logger.Info("Server started")
		if err := srv.ListenAndServe(); err != nil {
			slog.Info("error: ", err)
			panic(err)
		}
		logger.Info("Server stopped ")
	}()

	select {}
}
