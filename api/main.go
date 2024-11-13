package main

import (
	"log"

	"stockTicker/server"
)

func main() {
	s := server.NewServer()
	log.Println("Starting server on :8080")
	if err := s.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
