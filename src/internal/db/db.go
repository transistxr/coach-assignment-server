package db

import (
	"database/sql"
	//"fmt"
	"log"
	//"os"
	_ "github.com/lib/pq"

)

var DB *sql.DB

func Init() *sql.DB {

    log.Println("Initializing PostgreSQL Database Connection")

    dsn := "postgresql://postgres:postgres123@postgres:5432/scheduling_db?sslmode=disable"
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatalf("failed to connect to database: %v", err)
    }

    if err := db.Ping(); err != nil {
        log.Fatalf("failed to ping database: %v", err)
    }

    log.Println("Connected to PostgreSQL")
    return db
}
