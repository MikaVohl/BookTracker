package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

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

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (a *App) ListBooks(ctx context.Context) ([]Book, error) {
	rows, err := a.DB.QueryContext(ctx, "SELECT id, name, author, finished FROM Books")
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

func (a *App) CreateBook(ctx context.Context, b *Book) error {
	res, err := a.DB.ExecContext(ctx, "INSERT INTO Books(name, author, finished) VALUES (?, ?, ?)", b.Name, b.Author, b.Finished)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	b.ID = int(id)
	return nil
}

func main() {
	db, err := sql.Open("sqlite3", "file:app.db?cache=shared&_foreign_keys=1")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	_, err = db.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS Books(id INTEGER PRIMARY KEY, name TEXT, author TEXT, finished INTEGER)")
	if err != nil {
		log.Fatal(err)
	}

	app := &App{DB: db}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/books", app.booksHandler)

	log.Println("Listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil { // initialize var; check condition
		log.Fatal(err)
	}
}

func (a *App) booksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		books, err := a.ListBooks(r.Context())
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(books); err != nil {
			http.Error(w, "could not encode books", http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
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
		if strings.TrimSpace(b.Name) == "" || strings.TrimSpace(b.Author) == "" {
			http.Error(w, "name and author are required", http.StatusBadRequest)
			return
		}

		if err := a.CreateBook(r.Context(), &b); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", fmt.Sprintf("/api/books/%d", b.ID))
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(b); err != nil {
			http.Error(w, "could not encode response", http.StatusInternalServerError)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
