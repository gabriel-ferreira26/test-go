// Command test-go runs a small HTTP API for managing notes.
package main

import (
	"log"
	"net/http"

	"github.com/gabriel-ferreira26/test-go/internal/auth"
	"github.com/gabriel-ferreira26/test-go/internal/notes"
)

func main() {
	authStore := auth.NewStore()
	authHandlers := auth.NewHandlers(authStore)

	notesStore := notes.NewStore()
	notesHandlers := notes.NewHandlers(notesStore)

	mux := http.NewServeMux()
	authHandlers.Routes(mux)
	notesHandlers.Routes(mux, authStore.RequireAuth)

	const addr = ":8080"
	log.Printf("notes API listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
