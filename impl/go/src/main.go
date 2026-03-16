package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type NewUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateUser struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

type App struct {
	db *pgxpool.Pool
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL env var must be set for the API")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to create postgres pool: %v", err)
	}
	defer pool.Close()

	app := &App{db: pool}

	http.HandleFunc("/users", app.handleUsers)
	http.HandleFunc("/users/", app.handleUserByID)

	log.Println("Go API listening on http://0.0.0.0:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func (a *App) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.listUsers(w, r)
	case http.MethodPost:
		a.createUser(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *App) handleUserByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/users/"):]
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		a.getUser(w, r, id)
	case http.MethodPut:
		a.updateUser(w, r, id)
	case http.MethodDelete:
		a.deleteUser(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *App) listUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.Query(r.Context(), `SELECT id, name, email, created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	writeJSON(w, http.StatusOK, users)
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request, id string) {
	var u User
	err := a.db.QueryRow(r.Context(), `SELECT id, name, email, created_at FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, u)
}

func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	var payload NewUser
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	var u User
	err := a.db.QueryRow(
		r.Context(),
		`INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, name, email, created_at`,
		payload.Name,
		payload.Email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, u)
}

func (a *App) updateUser(w http.ResponseWriter, r *http.Request, id string) {
	var payload UpdateUser
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	var current User
	err := a.db.QueryRow(r.Context(), `SELECT id, name, email, created_at FROM users WHERE id = $1`, id).
		Scan(&current.ID, &current.Name, &current.Email, &current.CreatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	name := current.Name
	if payload.Name != nil {
		name = *payload.Name
	}
	email := current.Email
	if payload.Email != nil {
		email = *payload.Email
	}

	var u User
	err = a.db.QueryRow(
		r.Context(),
		`UPDATE users SET name = $1, email = $2 WHERE id = $3 RETURNING id, name, email, created_at`,
		name, email, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, u)
}

func (a *App) deleteUser(w http.ResponseWriter, r *http.Request, id string) {
	cmd, err := a.db.Exec(r.Context(), `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	if cmd.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
