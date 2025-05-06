package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// register handler
	http.HandleFunc("/api/books", booksHandler)

	// start server on port 8080
	log.Println("Listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func booksHandler(http.ResponseWriter, *http.Request) {
	fmt.Print("received")
	return
}
