package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	handler := NewServer(NewStore())
	log.Printf("budget API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
