package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB
var templates *template.Template

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "./users.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS users (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		username   TEXT NOT NULL UNIQUE,
		password   TEXT NOT NULL,
		email      TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);`

	if _, err = db.Exec(createTable); err != nil {
		log.Fatalf("failed to create users table: %v", err)
	}

	// Seed with sample users if the table is empty.
	var count int
	if err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		log.Fatalf("failed to count users: %v", err)
	}
	if count == 0 {
		seed := []struct{ username, password, email string }{
			{"admin", "S3cr3tP@ss!", "admin@example.com"},
			{"alice", "alice1234", "alice@example.com"},
			{"bob", "b0bpass", "bob@example.com"},
		}
		now := time.Now().UTC().Format("2006-01-02 15:04:05")
		for _, u := range seed {
			_, err = db.Exec(
				"INSERT INTO users (username, password, email, created_at) VALUES (?, ?, ?, ?)",
				u.username, u.password, u.email, now,
			)
			if err != nil {
				log.Fatalf("failed to seed user %s: %v", u.username, err)
			}
		}
		log.Println("database seeded with sample users")
	}
}

// indexHandler serves the landing page.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("template error: %v", err)
	}
}

// loginHandler serves the login form (GET) and processes login (POST).
func loginHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Error":   "",
		"Success": "",
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		// VULNERABILITY: SQL injection – user input is concatenated directly
		// into the query without parameterisation.
		query := fmt.Sprintf(
			"SELECT id, username, email FROM users WHERE username = '%s' AND password = '%s'",
			username, password,
		)
		log.Printf("login attempt for username: %s", username)

		rows, err := db.Query(query)
		if err != nil {
			data["Error"] = "Database error: " + err.Error()
		} else {
			defer rows.Close()
			if rows.Next() {
				var id int
				var uname, email string
				_ = rows.Scan(&id, &uname, &email)
				data["Success"] = fmt.Sprintf("Welcome back, %s! (email: %s)", uname, email)
			} else {
				data["Error"] = "Invalid username or password."
			}
		}
	}

	if err := templates.ExecuteTemplate(w, "login.html", data); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("template error: %v", err)
	}
}

func main() {
	initDB()
	defer db.Close()

	var err error
	templates, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/login", loginHandler)

	addr := ":8080"
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
