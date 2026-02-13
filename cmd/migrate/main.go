package main

import (
	"fmt"
	"os"

	"github.com/iLeoon/realtime-gateway/internal/db"
)

func main() {
	databaseURL := os.Getenv("SCALINGO_POSTGRESQL_URL")
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "ERROR: DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	fmt.Println("Starting database migrations...")

	if err := db.RunMigrations(databaseURL); err != nil {
		fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Database migrations completed successfully")
}
