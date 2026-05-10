package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	catalogpb "docflow.local/pkg/pb/catalog"
	iampb "docflow.local/pkg/pb/iam"
	"orchestrator-service/api"
	db "orchestrator-service/db/generated"
	"orchestrator-service/publisher"
	"orchestrator-service/service"
)

func main() {
	cfg := loadConfig()

	conn, err := sql.Open("pgx", cfg.dsn())
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer conn.Close()
	if err = conn.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	log.Println("orchestrator-service: db connected")

	queries := db.New(conn)

	iamConn, err := grpc.NewClient(cfg.IAMServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("iam grpc client: %v", err)
	}
	defer iamConn.Close()

	catalogConn, err := grpc.NewClient(cfg.CatalogServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("catalog grpc client: %v", err)
	}
	defer catalogConn.Close()

	pub, err := publisher.New(cfg.AmqpURL)
	if err != nil {
		log.Printf("WARNING: RabbitMQ not available: %v — events will be logged only", err)
		pub = nil
	} else {
		defer pub.Close()
	}

	svc := service.New(
		queries,
		catalogpb.NewCatalogServiceClient(catalogConn),
		iampb.NewIAMServiceClient(iamConn),
		pub,
		cfg.MetricsCSVDir,
	)

	if cfg.MetricsLogIntervalSeconds > 0 {
		interval := time.Duration(cfg.MetricsLogIntervalSeconds) * time.Second
		log.Printf("orchestrator-service: metrics logger enabled (interval=%s)", interval)
		go func() {
			if err := svc.LogMetricsSnapshot(context.Background()); err != nil {
				log.Printf("metrics logger error: %v", err)
			}
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for range ticker.C {
				if err := svc.LogMetricsSnapshot(context.Background()); err != nil {
					log.Printf("metrics logger error: %v", err)
				}
			}
		}()
	}

	router := api.NewRouter(svc, cfg.AllowedOrigin)

	log.Printf("orchestrator-service: listening on :%s", cfg.HTTPPort)
	srv := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: router}
	if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("http server: %v", err)
	}
}
