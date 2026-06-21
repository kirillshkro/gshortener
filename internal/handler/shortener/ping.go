package shortener

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/kirillshkro/gshortener/internal/config"

	_ "github.com/lib/pq"
)

// Pinger interface defines the contract for handling ping requests.
type Pinger interface {
	Ping(w http.ResponseWriter, r *http.Request)
}

// Ping handles the health check request to verify database connectivity.
//
// Parameters:
//   - w: The HTTP response writer to send the response back to the client.
//   - r: The HTTP request containing the health check information.
func (s Service) Ping(w http.ResponseWriter, r *http.Request) {
	var (
		db  *sql.DB
		err error
	)
	cfg := config.GetConfig()
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
