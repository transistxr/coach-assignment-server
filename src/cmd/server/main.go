package main

import (
	"log"

	"github.com/transistxr/coach-assignment-server/src/internal/db"
	"github.com/transistxr/coach-assignment-server/src/internal/server"
)

func main() {
	sqlDB := db.Init()
	rdb := db.InitRDB(sqlDB)
	s := server.New(sqlDB, rdb)
	log.Println("Starting server on :3000")
	if err := s.Start(":3000"); err != nil {
		log.Fatal(err)
	}
}
