module catalog-service

go 1.22.0

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	docflow.local/pkg v0.0.0
	github.com/jackc/pgx/v5 v5.7.4
	google.golang.org/grpc v1.71.1
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace docflow.local/pkg => ../../pkg
