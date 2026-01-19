package main

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	
	"besart951/go_infra_link/backend/internal/domain"
	"besart951/go_infra_link/backend/internal/repository"
	"besart951/go_infra_link/backend/internal/service"
)

func main() {
	// 1. DB Connection
	dsn := os.Getenv("DATABASE_URL")
    if dsn == "" { dsn = "host=localhost user=postgres password=postgres dbname=mydb port=5432 sslmode=disable" }
    
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 2. Migration
	log.Println("Migrating database...")
	db.AutoMigrate(
        &domain.User{}, 
        &domain.BusinessDetails{},
        &domain.Project{}, 
        &domain.Phase{},
        &domain.Building{},
        &domain.ControlCabinet{},
    )

	// 3. Wiring
	projRepo := repository.NewProjectRepository(db)
	projService := service.NewProjectService(projRepo)

	log.Printf("Service initialized: %v", projService)
    log.Println("Server ready to start...")
}
