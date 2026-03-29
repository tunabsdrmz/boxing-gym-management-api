package db

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/tunabsdrmz/boxing-gym-management/internal/config"
)

func NewDB(config config.DbConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.Addr)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}