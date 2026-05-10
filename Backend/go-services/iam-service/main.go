package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	db "iam-service/db/generated"
	iamhttp "iam-service/transport/http"
	grpcserver "iam-service/transport/grpc"

	_ "github.com/jackc/pgx/v5/stdlib"
	iampb "docflow.local/pkg/pb/iam"
	"google.golang.org/grpc"
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

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}
	grpcSrv := grpc.NewServer()
	iampb.RegisterIAMServiceServer(grpcSrv, grpcserver.New(queries))
	go func() {
		log.Printf("iam-service gRPC listening on :%s", cfg.GRPCPort)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	router := iamhttp.NewRouter(queries, cfg.AllowedOrigin)
	addr := ":" + cfg.HTTPPort
	log.Printf("iam-service HTTP listening on %s", addr)
	if err = http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

