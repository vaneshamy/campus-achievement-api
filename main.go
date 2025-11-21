package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-fiber/config"
	"go-fiber/database"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Setup logger
	config.SetupLogger()

	// Connect to database
	db, err := database.ConnectPostgres()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// ===============================
	//  CLI COMMANDS (migrate / seed)
	// ===============================
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			database.Migrate(db)
			return
		case "seed":
			database.Seed(db)
			return
		default:
			log.Println("Unknown command:", os.Args[1])
			log.Println("Available commands: migrate, seed")
			return
		}
	}

	// Create Fiber app with routes
	app := config.NewApp(db)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	// Start server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
