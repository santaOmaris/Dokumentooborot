package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	catalogpb "docflow.local/pkg/pb/catalog"
	db "catalog-service/db/generated"
	"catalog-service/api"
	grpcserver "catalog-service/api/grpc"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := loadConfig()

	conn, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer conn.Close()

	if err = conn.PingContext(context.Background()); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}
	log.Println("connected to postgres")

	queries := db.New(conn)

	fileConn, err := grpc.NewClient(cfg.FileServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to file-service: %v", err)
	}
	defer fileConn.Close()

	grpcSrv := grpc.NewServer()
	catalogpb.RegisterCatalogServiceServer(grpcSrv, grpcserver.New(queries))

	grpcLis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen grpc: %v", err)
	}
	go func() {
		log.Printf("catalog gRPC listening on :%s", cfg.GRPCPort)
		if err := grpcSrv.Serve(grpcLis); err != nil {
			log.Fatalf("grpc serve error: %v", err)
		}
	}()

	router := api.NewRouter(queries, fileConn, cfg.AllowedOrigin)
	addr := ":" + cfg.HTTPPort
	log.Printf("catalog-service HTTP listening on %s", addr)
	if err = http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("http server error: %v", err)
	}
}
