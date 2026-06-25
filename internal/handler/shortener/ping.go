package shortener

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/kirillshkro/gshortener/internal/config"

	_ "github.com/lib/pq"
)

// Pinger interface defines the contract for database ping functionality
type Pinger interface {
	// Ping handles the database connection ping request
	Ping(w http.ResponseWriter, req *http.Request)
}

// Ping handles the database connection ping request
// It checks database connectivity
// Returns HTTP 200 OK if database is reachable, otherwise returns appropriate error
func (s Service) Ping(w http.ResponseWriter, r *http.Request) {
	var (
		db  *sql.DB
		err error
	)
	cfg := config.GetConfig()

	dsn := cfg.DSN
	if db, err = sql.Open("postgres", dsn); err != nil {
		http.Error(w, "database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		http.Error(w, "database ping error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
