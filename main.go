// Command test-go runs a small HTTP API for managing notes.
package main

import (
	"log"
	"net/http"

	"github.com/gabriel-ferreira26/test-go/internal/notes"
)

func main() {
	store := notes.NewStore()
	handlers := notes.NewHandlers(store)

	const addr = ":8080"
	log.Printf("notes API listening on %s", addr)
	if err := http.ListenAndServe(addr, handlers.Routes()); err != nil {
		log.Fatal(err)
	}
}
