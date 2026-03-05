package shortener

import (
	"database/sql"
	"net/http"

	"github.com/kirillshkro/gshortener/internal/config"
)

type Pinger interface {
	Ping(w http.ResponseWriter, req *http.Request)
}

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

	w.WriteHeader(http.StatusOK)
}
