package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	DB *sql.DB
}

type Book struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Author   string `json:"author"`
	Finished bool   `json:"finished"`
}

func (a *App) listBooks() ([]Book, error) {
	rows, err := a.DB.Query("SELECT id, name, author, finished FROM Books")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Name, &b.Author, &b.Finished); err != nil {
			return books, err
		}
		books = append(books, b)
	}
	if err := rows.Err(); err != nil {
		return books, err
	}
	return books, nil
}

func (a *App) addBook(name string, author string, finished bool) error {
	_, err := a.DB.Exec("INSERT INTO Books(name, author, finished) VALUES (?, ?, ?)", name, author, finished)
	return err
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

	app := &App{DB: db}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/books", app.booksHandler)
	mux.HandleFunc("/api/addBook", app.addBookHandler)

	log.Println("Listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil { // initialize var; check condition
		log.Fatal(err)
	}
}

func (a *App) booksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	books, err := a.listBooks()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
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
	err := a.addBook(b.Name, b.Author, b.Finished)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	books, err := a.listBooks()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(books); err != nil {
		http.Error(w, "could not encode books", http.StatusInternalServerError)
		return
	}
}
