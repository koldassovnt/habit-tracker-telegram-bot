package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, dsn())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	sql, err := os.ReadFile("migrations/v1__init.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	_, err = conn.Exec(ctx, string(sql))
	if err != nil {
		log.Fatalf("Failed to run migration: %v", err)
	}

	log.Println("Migration done")
}

func dsn() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	if host == "" || port == "" || user == "" || password == "" || name == "" {
		log.Fatal("Missing required database environment variables")
	}

	return "host=" + host + " port=" + port + " user=" + user +
		" password=" + password + " dbname=" + name + " sslmode=disable"
}
