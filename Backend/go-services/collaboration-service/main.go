package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	db "collaboration-service/db/generated"
	"collaboration-service/api"
	"collaboration-service/consumer"
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
	log.Println("collaboration-service: db connected")

	queries := db.New(conn)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		for {
			if err := consumer.Run(ctx, cfg.AmqpURL, queries); err != nil {
				log.Printf("audit consumer error: %v — reconnecting in 5s", err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
				}
			} else {
				return
			}
		}
	}()

	router := api.NewRouter(queries, cfg.AllowedOrigin)

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	go func() {
		log.Printf("collaboration-service: listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("collaboration-service: shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
