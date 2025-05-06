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

func listBooks() {
	books := loadBooks()

	for _, book := range books.Books {
		fmt.Println(book.ID)
		fmt.Println(book.Name)
		fmt.Println(book.Author)
	}
}

func addBook(id int, name string, author string) {
	book := Book{ID: id, Name: name, Author: author}
	books := loadBooks()
	books.Books = append(books.Books, book)
	saveBooks(books)
}

func main() {
	listBooks()
	addBook(2, "Notes from Underground", "Fyodor Dostoevsky")
	listBooks()
	// register handler
	// http.HandleFunc("/api/books", booksHandler)

	// // start server on port 8080
	// log.Println("Listening on http://localhost:8080")
	// if err := http.ListenAndServe(":8080", nil); err != nil { // initialize var; check condition
	// 	log.Fatal(err)
	// }
}

func booksHandler(http.ResponseWriter, *http.Request) {
	fmt.Println("received")
	return
}
