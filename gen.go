package main

//go:generate goapi-gen --package=spec --out ./internal/api/spec/course.gen.spec.go ./internal/api/spec/course.spec.json
//go:generate tern migrate --migrations ./internal/pgstore/migrations/ --config ./internal/pgstore/migrations/tern.conf
//go:generate sqlc generate -f ./internal/pgstore/sqlc.yaml
