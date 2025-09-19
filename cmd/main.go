package main

import (
	"cyberhunt/internal/database"
	"cyberhunt/internal/handlers"
	"flag"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	var addr = flag.String("addr", ":8080", "Address and port to run the server")
	flag.Parse()

	myEnv, err := godotenv.Read()
	if err != nil {
		log.Panic("No .env file found")
	}

	pgUser := myEnv["POSTGRES_USER"]
	pgPassword := myEnv["POSTGRES_PASSWORD"]
	pgHost := myEnv["POSTGRES_HOST"]
	pgPort := myEnv["POSTGRES_PORT"]
	pgDB := myEnv["POSTGRES_DB"]

	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		pgUser, pgPassword, pgHost, pgPort, pgDB,
	)
	// Initialize database
	db, err := database.InitDB(dbURL)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize handlers
	jwtSecret := myEnv["JWT_SECRET"]
	h := handlers.NewHandler(db, jwtSecret)
	r := SetupRoutes(h)

	// Start server
	log.Println("Server starting on", *addr)
	if err := r.Run(*addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
