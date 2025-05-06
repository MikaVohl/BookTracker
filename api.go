package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type Books struct {
	Books []Book `json:"books"`
}

type Book struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Author string `json:"author"`
}

func loadBooks() Books {
	f, err := os.Open("books.json")
	if err != nil {
		fmt.Println("Error opening json:")
		log.Fatal(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		fmt.Println("Error reading json:")
		log.Fatal(err)
	}

	var books Books
	if err := json.Unmarshal(data, &books); err != nil {
		fmt.Println("Error decoding json:")
		log.Fatal(err)
	}
	return books
}

func saveBooks(books Books) {
	out, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		fmt.Println("JSON marshal error:")
		log.Fatal(err)
	}

	if err := os.WriteFile("books.json", out, 0644); err != nil {
		fmt.Println("Write file error:")
		log.Fatal(err)
	}
}

func listBooks() Books {
	return loadBooks()
}

func addBook(book Book) {
	books := loadBooks()
	books.Books = append(books.Books, book)
	saveBooks(books)
}

func main() {
	// register handler
	http.HandleFunc("/api/books", booksHandler)
	http.HandleFunc("/api/addBook", addBookHandler)

	// start server on port 8080
	log.Println("Listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil { // initialize var; check condition
		log.Fatal(err)
	}
}

func booksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	books := listBooks()
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(books); err != nil {
		http.Error(w, "could not encode books", http.StatusInternalServerError)
		return
	}
}

func addBookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		if err == io.EOF {
			http.Error(w, "empty body", http.StatusBadRequest)
		} else {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		}
		return
	}

	addBook(b)
	books := listBooks()
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(books); err != nil {
		http.Error(w, "could not encode books", http.StatusInternalServerError)
		return
	}
}
