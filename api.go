package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	DB *sql.DB
}

type Books struct {
	Books []Book `json:"books"`
}

type Book struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Author   string `json:"author"`
	Finished bool   `json:"finished"`
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

func listBooks(db *sql.DB) Books {
	return loadBooks()
}

func addBook(name string, author string, finished bool, db *sql.DB) {
	_, err := db.Exec("INSERT INTO Books(name, author, finished) VALUES (?, ?, ?)", name, author, finished)
	if err != nil {
		log.Fatal(err)
	}
}

func maxId() int {
	id := 0
	books := loadBooks()
	for _, element := range books.Books {
		id = max(id, element.ID)
	}
	return id
}

func main() {
	db, err := sql.Open("sqlite3", "file:app.db?cache=shared&_foreign_keys=1")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Books(id INTEGER PRIMARY KEY, name TEXT, author TEXT, finished BIT)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO Books(id, name, author, finished) VALUES (?, ?, ?, ?)", 1, "The Stranger", "Albert Camus", 1)
	if err != nil {
		log.Fatal(err)
	}

	row := db.QueryRow("SELECT name FROM Books WHERE id = 1")
	var name string
	if err := row.Scan(&name); err != nil {
		log.Fatal(err)
	}
	fmt.Println(name)

	app := &App{DB: db}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/books", app.booksHandler)
	mux.HandleFunc("/api/addBook", app.addBookHandler)

	// start server on port 8080
	log.Println("Listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil { // initialize var; check condition
		log.Fatal(err)
	}
}

func (a *App) booksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	books := listBooks(a.DB)
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(books); err != nil {
		http.Error(w, "could not encode books", http.StatusInternalServerError)
		return
	}
}

func (a *App) addBookHandler(w http.ResponseWriter, r *http.Request) {
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
	addBook(b.Name, b.Author, b.Finished, a.DB)
	books := listBooks(a.DB)
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(books); err != nil {
		http.Error(w, "could not encode books", http.StatusInternalServerError)
		return
	}
}
