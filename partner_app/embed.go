package partnerapp

import "embed"

//go:embed migrations/*.sql
var MigrationFS embed.FS

//go:generate go tool swag init -g cmd/server/main.go -d ./ --parseInternal --v3.1 -o api --outputTypes json
//go:generate mv api/swagger.json api/openapi.json
//go:generate go run ./cmd/fixopenapi
//go:embed api/openapi.json
var OpenAPIFS embed.FS
