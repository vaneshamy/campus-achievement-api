package database

import (
	"database/sql"
	"log"

	"go-fiber/database/migrations"
	"go-fiber/database/seeders"
)

func Migrate(db *sql.DB) {
	log.Println("🔥 Running migrations...")

	if err := migrations.CreateTables(db); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	log.Println("✅ Migration completed")
}

func Seed(db *sql.DB) {
	log.Println("🌱 Running seeders...")

	if err := seeders.SeedRolesPermissions(db); err != nil {
		log.Fatalf("Seeder error: %v", err)
	}

	log.Println("✅ Seeder completed")
}
